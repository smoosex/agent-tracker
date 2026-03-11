package sync

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"agent-tracker/internal/database"
	"agent-tracker/internal/models"
)

type GitHubRelease struct {
	ID          int64     `json:"id"`
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	HTMLURL     string    `json:"html_url"`
	PublishedAt time.Time `json:"published_at"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	CreatedAt   time.Time `json:"created_at"`
}

func computeHash(body string) string {
	hash := sha256.Sum256([]byte(body))
	return hex.EncodeToString(hash[:])
}

func fetchReleases(repo, token string, etag string) ([]GitHubRelease, string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases?per_page=100", repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return nil, etag, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("github api returned status %d", resp.StatusCode)
	}

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, "", err
	}

	newEtag := resp.Header.Get("ETag")
	return releases, newEtag, nil
}

func SyncTool(tool *models.Tool, fullSync bool) error {
	token := os.Getenv("GITHUB_TOKEN")

	var syncState models.SyncState
	database.DB.FirstOrCreate(&syncState, models.SyncState{ToolID: tool.ID})

	now := time.Now()
	syncState.LastAttemptAt = &now

	releases, newEtag, err := fetchReleases(tool.SourceRepo, token, syncState.EtagPage1)
	if err != nil {
		syncState.LastError = err.Error()
		database.DB.Save(&syncState)
		return err
	}

	if len(releases) == 0 && newEtag != "" {
		syncState.EtagPage1 = newEtag
		database.DB.Save(&syncState)
		return nil
	}

	for _, r := range releases {
		if r.Draft {
			continue
		}

		sourceID := fmt.Sprintf("%d", r.ID)
		contentHash := computeHash(r.Body)

		var entry models.Entry
		result := database.DB.Where("tool_id = ? AND source_entry_id = ?", tool.ID, sourceID).First(&entry)

		if result.Error == nil {
			if entry.ContentHash != contentHash {
				entry.Title = r.Name
				entry.BodyMD = r.Body
				entry.ContentHash = contentHash
				entry.URL = r.HTMLURL
				entry.SourceUpdatedAt = r.CreatedAt
				entry.IsPrerelease = 0
				if r.Prerelease {
					entry.IsPrerelease = 1
				}
				entry.UpdatedAt = time.Now()
				database.DB.Save(&entry)
			}
		} else {
			entry = models.Entry{
				ToolID:          tool.ID,
				SourceEntryID:   sourceID,
				Version:         r.TagName,
				Title:           r.Name,
				URL:             r.HTMLURL,
				BodyMD:          r.Body,
				PublishedAt:     r.PublishedAt,
				SourceUpdatedAt: r.CreatedAt,
				ContentHash:     contentHash,
				IsPrerelease:    0,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}
			if r.Prerelease {
				entry.IsPrerelease = 1
			}
			database.DB.Create(&entry)
		}
	}

	successAt := time.Now()
	syncState.LastSuccessAt = &successAt
	if newEtag != "" {
		syncState.EtagPage1 = newEtag
	}
	if fullSync {
		syncState.LastFullSyncAt = &successAt
	}
	syncState.LastError = ""
	database.DB.Save(&syncState)

	return nil
}

func SyncAll(fullSync bool) {
	var tools []models.Tool
	database.DB.Where("is_active = ?", 1).Find(&tools)

	for _, tool := range tools {
		SyncTool(&tool, fullSync)
	}
}

func InitTools() {
	tools := []models.Tool{
		{Slug: "claude-code", Name: "Claude Code", SourceType: "github", SourceRepo: "anthropics/claude-code", Homepage: "https://claude.ai/code", IsActive: 1, CreatedAt: time.Now()},
		{Slug: "codex", Name: "OpenAI Codex", SourceType: "github", SourceRepo: "openai/codex", Homepage: "https://github.com/openai/codex", IsActive: 1, CreatedAt: time.Now()},
		{Slug: "gemini-cli", Name: "Gemini CLI", SourceType: "github", SourceRepo: "google-gemini/gemini-cli", Homepage: "https://github.com/google-gemini/gemini-cli", IsActive: 1, CreatedAt: time.Now()},
		{Slug: "opencode", Name: "OpenCode", SourceType: "github", SourceRepo: "opencode-ai/opencode", Homepage: "https://github.com/opencode-ai/opencode", IsActive: 1, CreatedAt: time.Now()},
	}

	for _, t := range tools {
		var existing models.Tool
		if err := database.DB.Where("slug = ?", t.Slug).First(&existing).Error; err != nil {
			database.DB.Create(&t)
		}
	}
}