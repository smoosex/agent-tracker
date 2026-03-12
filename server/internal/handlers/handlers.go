package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	stdsync "sync"
	"time"

	"agent-tracker/internal/database"
	"agent-tracker/internal/logging"
	"agent-tracker/internal/models"
	trackerSync "agent-tracker/internal/sync"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	syncCooldown      = 10 * time.Second
	syncStateMu       stdsync.Mutex
	syncRunning       bool
	syncStartedAt     *time.Time
	syncCooldownUntil *time.Time
	errSyncInProgress = errors.New("sync already in progress")
	syncEventMu       stdsync.Mutex
	syncSubscribers   = map[chan SyncStatusResponse]struct{}{}
)

type SyncStatusResponse struct {
	Running                  bool       `json:"running"`
	StartedAt                *time.Time `json:"started_at,omitempty"`
	CooldownUntil            *time.Time `json:"cooldown_until,omitempty"`
	CooldownRemainingSeconds int        `json:"cooldown_remaining_seconds"`
}

type RecentLogsResponse struct {
	Path  string   `json:"path"`
	Lines []string `json:"lines"`
}

type SyncFailureResponse struct {
	ID        uint      `json:"id"`
	ToolSlug  string    `json:"tool_slug"`
	Error     string    `json:"error"`
	FullSync  int       `json:"full_sync"`
	CreatedAt time.Time `json:"created_at"`
}

func currentSyncStatus() SyncStatusResponse {
	syncStateMu.Lock()
	response := SyncStatusResponse{
		Running:   syncRunning,
		StartedAt: syncStartedAt,
	}
	if syncCooldownUntil != nil && time.Now().Before(*syncCooldownUntil) {
		response.CooldownUntil = syncCooldownUntil
		response.CooldownRemainingSeconds = int(time.Until(*syncCooldownUntil).Seconds()) + 1
	}
	syncStateMu.Unlock()

	return response
}

func publishSyncStatus(status SyncStatusResponse) {
	syncEventMu.Lock()
	for subscriber := range syncSubscribers {
		select {
		case subscriber <- status:
		default:
		}
	}
	syncEventMu.Unlock()
}

func RunSync() (result trackerSync.SyncResult, err error) {
	syncStateMu.Lock()
	now := time.Now()
	if syncRunning {
		syncStateMu.Unlock()
		return trackerSync.SyncResult{}, errSyncInProgress
	}
	if syncCooldownUntil != nil && now.Before(*syncCooldownUntil) {
		remaining := int(time.Until(*syncCooldownUntil).Seconds()) + 1
		syncStateMu.Unlock()
		return trackerSync.SyncResult{}, fmt.Errorf("sync cooldown active, try again in %ds", remaining)
	}
	syncRunning = true
	syncStartedAt = &now
	syncStateMu.Unlock()
	publishSyncStatus(SyncStatusResponse{
		Running:   true,
		StartedAt: &now,
	})

	defer func() {
		syncStateMu.Lock()
		syncRunning = false
		syncStartedAt = nil
		if err == nil {
			cooldownUntil := time.Now().Add(syncCooldown)
			syncCooldownUntil = &cooldownUntil
		} else {
			syncCooldownUntil = nil
		}
		syncStateMu.Unlock()
		publishSyncStatus(currentSyncStatus())
	}()

	trackerSync.InitTools()
	result, err = trackerSync.SyncAll(false)
	return result, err
}

func GetSyncStatus(c *gin.Context) {
	c.JSON(http.StatusOK, currentSyncStatus())
}

func GetSyncEvents(c *gin.Context) {
	statuses := make(chan SyncStatusResponse, 4)

	syncEventMu.Lock()
	syncSubscribers[statuses] = struct{}{}
	syncEventMu.Unlock()

	defer func() {
		syncEventMu.Lock()
		delete(syncSubscribers, statuses)
		syncEventMu.Unlock()
		close(statuses)
	}()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}

	writeStatus := func(status SyncStatusResponse) bool {
		payload, err := json.Marshal(status)
		if err != nil {
			return false
		}
		if _, err := fmt.Fprintf(c.Writer, "event: sync-status\ndata: %s\n\n", payload); err != nil {
			return false
		}
		flusher.Flush()
		return true
	}

	if !writeStatus(currentSyncStatus()) {
		return
	}

	heartbeat := time.NewTicker(25 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case status := <-statuses:
			if !writeStatus(status) {
				return
			}
		case <-heartbeat.C:
			if _, err := fmt.Fprint(c.Writer, ": keep-alive\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func parseLimit(raw string, defaultValue int) int {
	if raw == "" {
		return defaultValue
	}

	limit, err := strconv.Atoi(raw)
	if err != nil || limit <= 0 {
		return defaultValue
	}
	if limit > 500 {
		return 500
	}

	return limit
}

func GetRecentLogs(c *gin.Context) {
	lines, err := logging.ReadRecentLines(parseLimit(c.Query("limit"), 200))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read logs"})
		return
	}

	c.JSON(http.StatusOK, RecentLogsResponse{
		Path:  logging.Path(),
		Lines: lines,
	})
}

func GetRecentSyncFailures(c *gin.Context) {
	limit := parseLimit(c.Query("limit"), 50)

	var failures []models.SyncFailureRecord
	if err := database.DB.Order("created_at desc").Limit(limit).Find(&failures).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch sync failures"})
		return
	}

	response := make([]SyncFailureResponse, len(failures))
	for i, failure := range failures {
		response[i] = SyncFailureResponse{
			ID:        failure.ID,
			ToolSlug:  failure.ToolSlug,
			Error:     failure.Error,
			FullSync:  failure.FullSync,
			CreatedAt: failure.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"failures": response,
	})
}

type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

func Health(c *gin.Context) {
	sqlDB, err := database.DB.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, HealthResponse{
			Status:   "error",
			Database: "connection failed",
		})
		return
	}

	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, HealthResponse{
			Status:   "error",
			Database: "ping failed",
		})
		return
	}

	c.JSON(http.StatusOK, HealthResponse{
		Status:   "ok",
		Database: "connected",
	})
}

type ToolResponse struct {
	ID         uint   `json:"id"`
	Slug       string `json:"slug"`
	Name       string `json:"name"`
	SourceType string `json:"source_type"`
	SourceRepo string `json:"source_repo"`
	Homepage   string `json:"homepage"`
	IsActive   int    `json:"is_active"`
}

func activeEntriesQuery() *gorm.DB {
	return database.DB.Model(&models.Entry{}).
		Joins("JOIN tools ON tools.id = entries.tool_id").
		Where("tools.is_active = ?", 1)
}

func GetTools(c *gin.Context) {
	var tools []models.Tool
	if err := database.DB.Where("is_active = ?", 1).Find(&tools).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch tools"})
		return
	}

	response := make([]ToolResponse, len(tools))
	for i, t := range tools {
		response[i] = ToolResponse{
			ID:         t.ID,
			Slug:       t.Slug,
			Name:       t.Name,
			SourceType: t.SourceType,
			SourceRepo: t.SourceRepo,
			Homepage:   t.Homepage,
			IsActive:   t.IsActive,
		}
	}

	c.JSON(http.StatusOK, response)
}

func GetTool(c *gin.Context) {
	slug := c.Param("slug")
	var tool models.Tool
	if err := database.DB.Where("slug = ? AND is_active = ?", slug, 1).First(&tool).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tool not found"})
		return
	}

	c.JSON(http.StatusOK, ToolResponse{
		ID:         tool.ID,
		Slug:       tool.Slug,
		Name:       tool.Name,
		SourceType: tool.SourceType,
		SourceRepo: tool.SourceRepo,
		Homepage:   tool.Homepage,
		IsActive:   tool.IsActive,
	})
}

type EntryResponse struct {
	ID           uint      `json:"id"`
	ToolID       uint      `json:"tool_id"`
	ToolSlug     string    `json:"tool_slug"`
	ToolName     string    `json:"tool_name"`
	Version      string    `json:"version"`
	Title        string    `json:"title"`
	URL          string    `json:"url"`
	BodyMD       string    `json:"body_md"`
	Excerpt      string    `json:"excerpt"`
	PublishedAt  time.Time `json:"published_at"`
	IsPrerelease int       `json:"is_prerelease"`
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

func GetEntries(c *gin.Context) {
	toolSlug := c.Query("tool")
	cursor := c.Query("cursor")
	limit := 20

	query := database.DB.Model(&models.Entry{}).
		Joins("JOIN tools ON tools.id = entries.tool_id").
		Where("tools.is_active = ?", 1).
		Preload("Tool").
		Order("entries.published_at DESC").
		Limit(limit + 1)

	if toolSlug != "" {
		var tool models.Tool
		if err := database.DB.Where("slug = ? AND is_active = ?", toolSlug, 1).First(&tool).Error; err == nil {
			query = query.Where("tool_id = ?", tool.ID)
		}
	}

	if cursor != "" {
		var cursorEntry models.Entry
		if err := database.DB.First(&cursorEntry, cursor).Error; err == nil {
			query = query.Where("entries.published_at < ?", cursorEntry.PublishedAt)
		}
	}

	var entries []models.Entry
	if err := query.Find(&entries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch entries"})
		return
	}

	hasMore := len(entries) > limit
	if hasMore {
		entries = entries[:limit]
	}

	response := make([]EntryResponse, len(entries))
	for i, e := range entries {
		excerpt := truncate(e.BodyMD, 200)
		toolSlug := ""
		toolName := ""
		if e.Tool != nil {
			toolSlug = e.Tool.Slug
			toolName = e.Tool.Name
		}
		response[i] = EntryResponse{
			ID:           e.ID,
			ToolID:       e.ToolID,
			ToolSlug:     toolSlug,
			ToolName:     toolName,
			Version:      e.Version,
			Title:        e.Title,
			URL:          e.URL,
			BodyMD:       e.BodyMD,
			Excerpt:      excerpt,
			PublishedAt:  e.PublishedAt,
			IsPrerelease: e.IsPrerelease,
		}
	}

	result := gin.H{
		"entries": response,
		"hasMore": hasMore,
	}
	if hasMore && len(entries) > 0 {
		result["nextCursor"] = entries[len(entries)-1].ID
	}

	c.JSON(http.StatusOK, result)
}

func GetEntry(c *gin.Context) {
	id := c.Param("id")
	var entry models.Entry
	if err := database.DB.Preload("Tool").First(&entry, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "entry not found"})
		return
	}

	toolSlug := ""
	toolName := ""
	if entry.Tool != nil {
		toolSlug = entry.Tool.Slug
		toolName = entry.Tool.Name
	}

	c.JSON(http.StatusOK, EntryResponse{
		ID:           entry.ID,
		ToolID:       entry.ToolID,
		ToolSlug:     toolSlug,
		ToolName:     toolName,
		Version:      entry.Version,
		Title:        entry.Title,
		URL:          entry.URL,
		BodyMD:       entry.BodyMD,
		Excerpt:      truncate(entry.BodyMD, 200),
		PublishedAt:  entry.PublishedAt,
		IsPrerelease: entry.IsPrerelease,
	})
}

func GetToolEntries(c *gin.Context) {
	slug := c.Param("slug")
	cursor := c.Query("cursor")
	limit := 20

	var tool models.Tool
	if err := database.DB.Where("slug = ? AND is_active = ?", slug, 1).First(&tool).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tool not found"})
		return
	}

	query := database.DB.Model(&models.Entry{}).
		Where("tool_id = ?", tool.ID).
		Order("published_at DESC").
		Limit(limit + 1)

	if cursor != "" {
		var cursorEntry models.Entry
		if err := database.DB.First(&cursorEntry, cursor).Error; err == nil {
			query = query.Where("published_at < ?", cursorEntry.PublishedAt)
		}
	}

	var entries []models.Entry
	if err := query.Find(&entries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch entries"})
		return
	}

	hasMore := len(entries) > limit
	if hasMore {
		entries = entries[:limit]
	}

	response := make([]EntryResponse, len(entries))
	for i, e := range entries {
		response[i] = EntryResponse{
			ID:           e.ID,
			ToolID:       e.ToolID,
			ToolSlug:     tool.Slug,
			ToolName:     tool.Name,
			Version:      e.Version,
			Title:        e.Title,
			URL:          e.URL,
			BodyMD:       e.BodyMD,
			Excerpt:      truncate(e.BodyMD, 200),
			PublishedAt:  e.PublishedAt,
			IsPrerelease: e.IsPrerelease,
		}
	}

	result := gin.H{
		"tool": ToolResponse{
			ID:         tool.ID,
			Slug:       tool.Slug,
			Name:       tool.Name,
			SourceType: tool.SourceType,
			SourceRepo: tool.SourceRepo,
			Homepage:   tool.Homepage,
			IsActive:   tool.IsActive,
		},
		"entries": response,
		"hasMore": hasMore,
	}
	if hasMore && len(entries) > 0 {
		result["nextCursor"] = entries[len(entries)-1].ID
	}

	c.JSON(http.StatusOK, result)
}

func Search(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter required"})
		return
	}

	var entries []models.Entry
	query := activeEntriesQuery().
		Preload("Tool").
		Where("entries.title LIKE ? OR entries.body_md LIKE ?", "%"+q+"%", "%"+q+"%").
		Order("entries.published_at DESC").
		Limit(20)

	if err := query.Find(&entries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
		return
	}

	response := make([]EntryResponse, len(entries))
	for i, e := range entries {
		toolSlug := ""
		toolName := ""
		if e.Tool != nil {
			toolSlug = e.Tool.Slug
			toolName = e.Tool.Name
		}
		response[i] = EntryResponse{
			ID:           e.ID,
			ToolID:       e.ToolID,
			ToolSlug:     toolSlug,
			ToolName:     toolName,
			Version:      e.Version,
			Title:        e.Title,
			URL:          e.URL,
			BodyMD:       e.BodyMD,
			Excerpt:      truncate(e.BodyMD, 200),
			PublishedAt:  e.PublishedAt,
			IsPrerelease: e.IsPrerelease,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"query":   q,
		"entries": response,
	})
}

func TriggerSync(c *gin.Context) {
	result, err := RunSync()
	if err != nil {
		status := currentSyncStatus()
		if errors.Is(err, errSyncInProgress) {
			c.JSON(http.StatusConflict, gin.H{
				"error":                      "sync already in progress",
				"running":                    status.Running,
				"started_at":                 status.StartedAt,
				"cooldown_until":             status.CooldownUntil,
				"cooldown_remaining_seconds": status.CooldownRemainingSeconds,
			})
			return
		}
		if status.CooldownRemainingSeconds > 0 {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":                      err.Error(),
				"running":                    status.Running,
				"started_at":                 status.StartedAt,
				"cooldown_until":             status.CooldownUntil,
				"cooldown_remaining_seconds": status.CooldownRemainingSeconds,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     err.Error(),
			"status":    "error",
			"total":     result.Total,
			"succeeded": result.Succeeded,
			"failed":    result.Failed,
			"failures":  result.Failures,
		})
		return
	}

	status := "ok"
	message := "Data refreshed"
	if result.HasFailures() {
		status = "partial"
		message = fmt.Sprintf("Data refreshed with %d failures", result.Failed)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    status,
		"message":   message,
		"total":     result.Total,
		"succeeded": result.Succeeded,
		"failed":    result.Failed,
		"failures":  result.Failures,
	})
}
