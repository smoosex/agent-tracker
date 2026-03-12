package models

import (
	"time"
)

type Tool struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Slug       string    `gorm:"uniqueIndex;not null" json:"slug"`
	Name       string    `gorm:"not null" json:"name"`
	SourceType string    `gorm:"not null" json:"source_type"`
	SourceRepo string    `gorm:"not null" json:"source_repo"`
	Homepage   string    `json:"homepage"`
	IsActive   int       `gorm:"not null;default:1" json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}

type Entry struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ToolID          uint      `gorm:"not null;index" json:"tool_id"`
	SourceEntryID   string    `gorm:"not null;uniqueIndex:idx_tool_entry" json:"source_entry_id"`
	Version         string    `gorm:"not null" json:"version"`
	Title           string    `gorm:"not null" json:"title"`
	URL             string    `gorm:"not null" json:"url"`
	BodyMD          string    `gorm:"not null" json:"body_md"`
	PublishedAt     time.Time `json:"published_at"`
	SourceUpdatedAt time.Time `json:"source_updated_at"`
	ContentHash     string    `gorm:"not null" json:"content_hash"`
	IsPrerelease    int       `gorm:"not null;default:0" json:"is_prerelease"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Tool            *Tool     `gorm:"foreignKey:ToolID" json:"tool,omitempty"`
}

type SyncState struct {
	ToolID         uint       `gorm:"primaryKey" json:"tool_id"`
	LastAttemptAt  *time.Time `json:"last_attempt_at"`
	LastSuccessAt  *time.Time `json:"last_success_at"`
	LastFullSyncAt *time.Time `json:"last_full_sync_at"`
	LastError      string     `json:"last_error"`
	EtagPage1      string     `json:"etag_page_1"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type SyncFailureRecord struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ToolSlug  string    `gorm:"index;not null" json:"tool_slug"`
	Error     string    `gorm:"not null" json:"error"`
	FullSync  int       `gorm:"not null;default:0" json:"full_sync"`
	CreatedAt time.Time `gorm:"index;not null" json:"created_at"`
}
