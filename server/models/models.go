package models

import (
	"time"

	"gorm.io/gorm"
)

type Tool struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	Slug       string         `gorm:"uniqueIndex;not null" json:"slug"`
	Name       string         `gorm:"not null" json:"name"`
	SourceType string         `gorm:"not null" json:"source_type"`
	SourceRepo string         `gorm:"not null" json:"source_repo"`
	Homepage   string         `json:"homepage"`
	IsActive   int            `gorm:"not null;default:1" json:"is_active"`
	CreatedAt  time.Time      `json:"created_at"`
	Entries    []Entry        `json:"entries,omitempty"`
}

type Entry struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	ToolID          uint           `gorm:"not null;index" json:"tool_id"`
	SourceEntryID   string         `gorm:"not null" json:"source_entry_id"`
	Version         string         `gorm:"not null" json:"version"`
	Title           string         `gorm:"not null" json:"title"`
	URL             string         `gorm:"not null" json:"url"`
	BodyMD          string         `gorm:"not null" json:"body_md"`
	PublishedAt     time.Time      `json:"published_at"`
	SourceUpdatedAt time.Time      `json:"source_updated_at"`
	ContentHash     string         `gorm:"not null" json:"content_hash"`
	IsPrerelease    int            `gorm:"not null;default:0" json:"is_prerelease"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	Tool            *Tool          `json:"tool,omitempty"`

	gorm.Model
}

func (Entry) TableName() string {
	return "entries"
}

type SyncState struct {
	ToolID         uint      `gorm:"primaryKey" json:"tool_id"`
	LastAttemptAt  *time.Time `json:"last_attempt_at"`
	LastSuccessAt  *time.Time `json:"last_success_at"`
	LastFullSyncAt *time.Time `json:"last_full_sync_at"`
	LastError      string    `json:"last_error"`
	EtagPage1      string    `json:"etag_page_1"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}