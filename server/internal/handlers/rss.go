package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"agent-tracker/internal/database"
	"agent-tracker/internal/models"

	"github.com/gin-gonic/gin"
)

func escapeXML(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&apos;",
	)
	return replacer.Replace(s)
}

func formatRFC822(t time.Time) string {
	return t.Format("Mon, 02 Jan 2006 15:04:05 -0700")
}

func buildRSSFeed(title string, entries []models.Entry, baseURL string) string {
	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sb.WriteString(`<rss version="2.0">`)
	sb.WriteString(`<channel>`)
	sb.WriteString(fmt.Sprintf("<title>%s</title>", escapeXML(title)))
	sb.WriteString(fmt.Sprintf("<link>%s</link>", escapeXML(baseURL)))
	sb.WriteString(fmt.Sprintf("<description>%s</description>", escapeXML(title)))
	sb.WriteString("<language>en-us</language>")

	for _, e := range entries {
		sb.WriteString("<item>")
		sb.WriteString(fmt.Sprintf("<title>%s</title>", escapeXML(e.Title)))
		sb.WriteString(fmt.Sprintf("<link>%s</link>", escapeXML(e.URL)))
		sb.WriteString(fmt.Sprintf("<description>%s</description>", escapeXML(e.BodyMD)))
		sb.WriteString(fmt.Sprintf("<pubDate>%s</pubDate>", formatRFC822(e.PublishedAt)))
		sb.WriteString(fmt.Sprintf("<guid>%s</guid>", escapeXML(e.URL)))
		sb.WriteString("</item>")
	}

	sb.WriteString("</channel>")
	sb.WriteString("</rss>")

	return sb.String()
}

func GetAllRSS(c *gin.Context) {
	var entries []models.Entry
	if err := database.DB.Model(&models.Entry{}).
		Joins("JOIN tools ON tools.id = entries.tool_id").
		Where("tools.is_active = ?", 1).
		Preload("Tool").
		Order("entries.published_at DESC").
		Limit(50).
		Find(&entries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch entries"})
		return
	}

	baseURL := fmt.Sprintf("http://%s", c.Request.Host)
	rss := buildRSSFeed("Agent Tracker - All Releases", entries, baseURL)

	c.Data(http.StatusOK, "application/xml; charset=utf-8", []byte(rss))
}

func GetToolRSS(c *gin.Context) {
	slug := c.Param("slug")

	var tool models.Tool
	if err := database.DB.Where("slug = ? AND is_active = ?", slug, 1).First(&tool).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tool not found"})
		return
	}

	var entries []models.Entry
	if err := database.DB.Where("tool_id = ?", tool.ID).
		Order("published_at DESC").
		Limit(50).
		Find(&entries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch entries"})
		return
	}

	baseURL := fmt.Sprintf("http://%s", c.Request.Host)
	title := fmt.Sprintf("Agent Tracker - %s Releases", tool.Name)
	rss := buildRSSFeed(title, entries, baseURL)

	c.Data(http.StatusOK, "application/xml; charset=utf-8", []byte(rss))
}
