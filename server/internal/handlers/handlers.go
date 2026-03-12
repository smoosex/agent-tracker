package handlers

import (
	"net/http"
	stdsync "sync"
	"time"

	"agent-tracker/internal/database"
	"agent-tracker/internal/models"
	trackerSync "agent-tracker/internal/sync"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	syncStateMu stdsync.Mutex
	syncRunning bool
)

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
	syncStateMu.Lock()
	if syncRunning {
		syncStateMu.Unlock()
		c.JSON(http.StatusConflict, gin.H{"error": "sync already in progress"})
		return
	}
	syncRunning = true
	syncStateMu.Unlock()

	defer func() {
		syncStateMu.Lock()
		syncRunning = false
		syncStateMu.Unlock()
	}()

	trackerSync.InitTools()
	if err := trackerSync.SyncAll(false); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
