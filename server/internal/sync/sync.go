package sync

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	htmlpkg "html"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"agent-tracker/internal/database"
	"agent-tracker/internal/models"

	"golang.org/x/net/html"
)

const (
	openAICodexChangelogURL = "https://developers.openai.com/codex/changelog"
	openAIBaseURL           = "https://developers.openai.com"
	openCodeChangelogURL    = "https://opencode.ai/changelog"
	openCodeBaseURL         = "https://opencode.ai"
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

type SourceEntry struct {
	SourceEntryID   string
	Version         string
	Title           string
	URL             string
	BodyMD          string
	PublishedAt     time.Time
	SourceUpdatedAt time.Time
	IsPrerelease    int
}

func computeHash(body string) string {
	hash := sha256.Sum256([]byte(body))
	return hex.EncodeToString(hash[:])
}

func newHTTPClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second}
}

func fetchGitHubReleases(repo, token string, etag string) ([]GitHubRelease, string, error) {
	releasesURL := fmt.Sprintf("https://api.github.com/repos/%s/releases?per_page=100", repo)

	req, err := http.NewRequest(http.MethodGet, releasesURL, nil)
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

	resp, err := newHTTPClient().Do(req)
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

	return releases, resp.Header.Get("ETag"), nil
}

func fetchGitHubEntries(repo, token, etag string) ([]SourceEntry, string, error) {
	releases, newEtag, err := fetchGitHubReleases(repo, token, etag)
	if err != nil {
		return nil, "", err
	}

	entries := make([]SourceEntry, 0, len(releases))
	for _, release := range releases {
		if release.Draft {
			continue
		}

		entries = append(entries, SourceEntry{
			SourceEntryID:   fmt.Sprintf("%d", release.ID),
			Version:         release.TagName,
			Title:           release.Name,
			URL:             release.HTMLURL,
			BodyMD:          release.Body,
			PublishedAt:     release.PublishedAt,
			SourceUpdatedAt: release.CreatedAt,
			IsPrerelease:    boolToInt(release.Prerelease),
		})
	}

	return entries, newEtag, nil
}

func fetchOpenAIChangelogEntries(topic, etag string) ([]SourceEntry, string, error) {
	pageURL := fmt.Sprintf("%s?type=%s", openAICodexChangelogURL, url.QueryEscape(topic))
	body, err := fetchHTMLViaCurl(pageURL)
	if err != nil {
		return nil, "", err
	}

	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, "", err
	}

	entries := extractOpenAIChangelogEntries(doc, topic)
	return entries, etag, nil
}

func fetchOpenCodeChangelogEntries(etag string) ([]SourceEntry, string, error) {
	body, err := fetchHTMLViaCurl(openCodeChangelogURL)
	if err != nil {
		return nil, "", err
	}

	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, "", err
	}

	entries := extractOpenCodeChangelogEntries(doc)
	return entries, etag, nil
}

func fetchHTMLViaCurl(pageURL string) ([]byte, error) {
	cmd := exec.Command(
		"curl",
		"--compressed",
		"-fsSL",
		"-A",
		"Mozilla/5.0",
		"-H",
		"Accept: text/html,application/xhtml+xml",
		pageURL,
	)

	body, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("curl failed: %w: %s", err, strings.TrimSpace(string(body)))
	}

	return body, nil
}

func extractOpenAIChangelogEntries(doc *html.Node, topic string) []SourceEntry {
	var entries []SourceEntry

	walkNodes(doc, func(node *html.Node) {
		if node.Type != html.ElementNode || node.Data != "li" {
			return
		}

		if !hasTopic(attrValue(node, "data-codex-topics"), topic) {
			return
		}

		sourceID := attrValue(node, "id")
		if sourceID == "" {
			return
		}

		timeNode := findFirstDescendant(node, "time")
		publishedAt, err := time.Parse("2006-01-02", strings.TrimSpace(nodeText(timeNode)))
		if err != nil {
			return
		}

		headingNode := findFirstDescendant(node, "h3")
		headingText := collapseWhitespace(nodeText(headingNode))
		version := extractVersion(headingNode)
		title := strings.TrimSpace(headingText)
		if version != "" {
			title = strings.TrimSpace(strings.Replace(title, version, "", 1))
		}
		title = collapseWhitespace(title)
		if title == "" {
			title = headingText
		}

		articleNode := findFirstDescendant(node, "article")
		bodyMD := strings.TrimSpace(renderMarkdown(articleNode, openAIBaseURL))

		entries = append(entries, SourceEntry{
			SourceEntryID:   sourceID,
			Version:         version,
			Title:           title,
			URL:             fmt.Sprintf("%s?type=%s#%s", openAICodexChangelogURL, url.QueryEscape(topic), sourceID),
			BodyMD:          bodyMD,
			PublishedAt:     publishedAt,
			SourceUpdatedAt: publishedAt,
			IsPrerelease:    0,
		})
	})

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].PublishedAt.After(entries[j].PublishedAt)
	})

	return entries
}

func extractOpenCodeChangelogEntries(doc *html.Node) []SourceEntry {
	var entries []SourceEntry

	walkNodes(doc, func(node *html.Node) {
		if node.Type != html.ElementNode || node.Data != "article" {
			return
		}

		if attrValue(node, "data-component") != "release" {
			return
		}

		versionNode := findFirstDescendantWithAttr(node, "div", "data-slot", "version")
		linkNode := findFirstDescendant(versionNode, "a")
		version := collapseWhitespace(nodeText(linkNode))
		if version == "" {
			return
		}

		timeNode := findFirstDescendant(node, "time")
		publishedAt, err := time.Parse(time.RFC3339, attrValue(timeNode, "datetime"))
		if err != nil {
			return
		}

		contentNode := findFirstDescendantWithAttr(node, "div", "data-slot", "content")
		bodyMD := strings.TrimSpace(renderMarkdown(contentNode, openCodeBaseURL))

		entries = append(entries, SourceEntry{
			SourceEntryID:   version,
			Version:         version,
			Title:           version,
			URL:             normalizeURL(attrValue(linkNode, "href"), openCodeBaseURL),
			BodyMD:          bodyMD,
			PublishedAt:     publishedAt,
			SourceUpdatedAt: publishedAt,
			IsPrerelease:    0,
		})
	})

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].PublishedAt.After(entries[j].PublishedAt)
	})

	return entries
}

func extractVersion(node *html.Node) string {
	if node == nil {
		return ""
	}

	var version string
	walkNodes(node, func(child *html.Node) {
		if version != "" || child.Type != html.ElementNode || child.Data != "span" {
			return
		}

		if strings.Contains(attrValue(child, "class"), "text-tertiary") {
			version = collapseWhitespace(nodeText(child))
		}
	})

	return version
}

func renderMarkdown(node *html.Node, baseURL string) string {
	if node == nil {
		return ""
	}

	var builder strings.Builder
	renderMarkdownChildren(&builder, node, 0, baseURL)

	output := strings.TrimSpace(builder.String())
	for strings.Contains(output, "\n\n\n") {
		output = strings.ReplaceAll(output, "\n\n\n", "\n\n")
	}

	return output
}

func renderMarkdownChildren(builder *strings.Builder, node *html.Node, indent int, baseURL string) {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		renderMarkdownNode(builder, child, indent, baseURL)
	}
}

func renderMarkdownNode(builder *strings.Builder, node *html.Node, indent int, baseURL string) {
	if node == nil {
		return
	}

	if node.Type == html.TextNode {
		text := collapseWhitespace(htmlpkg.UnescapeString(node.Data))
		if text != "" {
			builder.WriteString(text)
		}
		return
	}

	if node.Type != html.ElementNode {
		return
	}

	switch node.Data {
	case "article", "div", "section":
		renderMarkdownChildren(builder, node, indent, baseURL)
	case "details":
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.ElementNode && child.Data == "summary" {
				continue
			}
			renderMarkdownNode(builder, child, indent, baseURL)
		}
	case "p":
		appendBlock(builder, renderInline(node, baseURL))
	case "h1", "h2", "h3", "h4", "h5", "h6":
		level := int(node.Data[1] - '0')
		appendBlock(builder, strings.Repeat("#", level)+" "+renderInline(node, baseURL))
	case "ul":
		renderList(builder, node, indent, false, baseURL)
	case "ol":
		renderList(builder, node, indent, true, baseURL)
	case "pre":
		code := strings.Trim(nodeCodeText(node), "\n")
		if code == "" {
			return
		}

		language := attrValue(node, "data-language")
		if builder.Len() > 0 && !strings.HasSuffix(builder.String(), "\n\n") {
			builder.WriteString("\n\n")
		}
		builder.WriteString("```")
		builder.WriteString(language)
		builder.WriteString("\n")
		builder.WriteString(code)
		builder.WriteString("\n```")
	case "blockquote":
		quote := strings.TrimSpace(renderMarkdown(node, baseURL))
		if quote == "" {
			return
		}
		lines := strings.Split(quote, "\n")
		for i, line := range lines {
			lines[i] = "> " + line
		}
		appendBlock(builder, strings.Join(lines, "\n"))
	default:
		renderMarkdownChildren(builder, node, indent, baseURL)
	}
}

func renderList(builder *strings.Builder, node *html.Node, indent int, ordered bool, baseURL string) {
	if builder.Len() > 0 && !strings.HasSuffix(builder.String(), "\n\n") {
		builder.WriteString("\n\n")
	}

	index := 1
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type != html.ElementNode || child.Data != "li" {
			continue
		}

		marker := "-"
		if ordered {
			marker = fmt.Sprintf("%d.", index)
			index++
		}

		item := strings.TrimSpace(renderListItem(child, indent+1, baseURL))
		if item == "" {
			continue
		}

		lines := strings.Split(item, "\n")
		for i, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}
			if i == 0 {
				builder.WriteString(strings.Repeat("  ", indent))
				builder.WriteString(marker)
				builder.WriteString(" ")
				builder.WriteString(line)
				builder.WriteString("\n")
				continue
			}

			builder.WriteString(strings.Repeat("  ", indent+1))
			builder.WriteString(line)
			builder.WriteString("\n")
		}
	}
}

func renderListItem(node *html.Node, indent int, baseURL string) string {
	var builder strings.Builder

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && (child.Data == "ul" || child.Data == "ol") {
			if builder.Len() > 0 && !strings.HasSuffix(builder.String(), "\n") {
				builder.WriteString("\n")
			}
			renderList(&builder, child, indent, child.Data == "ol", baseURL)
			continue
		}

		var fragment string
		if child.Type == html.ElementNode && (child.Data == "p" || strings.HasPrefix(child.Data, "h")) {
			fragment = renderInline(child, baseURL)
		} else {
			fragment = strings.TrimSpace(renderInlineNode(child, baseURL))
		}

		if fragment == "" {
			continue
		}

		if builder.Len() > 0 && !strings.HasSuffix(builder.String(), "\n") {
			builder.WriteString(" ")
		}
		builder.WriteString(fragment)
	}

	return strings.TrimSpace(builder.String())
}

func renderInline(node *html.Node, baseURL string) string {
	if node == nil {
		return ""
	}

	var builder strings.Builder
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		builder.WriteString(renderInlineNode(child, baseURL))
	}

	return collapseWhitespace(builder.String())
}

func renderInlineNode(node *html.Node, baseURL string) string {
	if node == nil {
		return ""
	}

	switch node.Type {
	case html.TextNode:
		return htmlpkg.UnescapeString(node.Data)
	case html.ElementNode:
		switch node.Data {
		case "code", "tt":
			text := collapseWhitespace(nodeCodeText(node))
			if text == "" {
				return ""
			}
			return "`" + text + "`"
		case "a":
			href := normalizeURL(attrValue(node, "href"), baseURL)
			text := collapseWhitespace(renderInline(node, baseURL))
			if text == "" {
				text = href
			}
			if href == "" {
				return text
			}
			return "[" + text + "](" + href + ")"
		case "strong", "b":
			text := collapseWhitespace(renderInline(node, baseURL))
			if text == "" {
				return ""
			}
			return "**" + text + "**"
		case "em", "i":
			text := collapseWhitespace(renderInline(node, baseURL))
			if text == "" {
				return ""
			}
			return "*" + text + "*"
		case "img":
			src := normalizeURL(attrValue(node, "src"), baseURL)
			if src == "" {
				return ""
			}
			return "![" + collapseWhitespace(attrValue(node, "alt")) + "](" + src + ")"
		case "br":
			return "\n"
		default:
			return renderInline(node, baseURL)
		}
	default:
		return ""
	}
}

func nodeText(node *html.Node) string {
	if node == nil {
		return ""
	}

	var builder strings.Builder
	walkNodes(node, func(child *html.Node) {
		if child.Type == html.TextNode {
			builder.WriteString(htmlpkg.UnescapeString(child.Data))
		}
	})

	return builder.String()
}

func nodeCodeText(node *html.Node) string {
	if node == nil {
		return ""
	}

	var builder strings.Builder
	var walk func(*html.Node)
	walk = func(current *html.Node) {
		if current == nil {
			return
		}

		if current.Type == html.TextNode {
			builder.WriteString(htmlpkg.UnescapeString(current.Data))
			return
		}

		if current.Type == html.ElementNode && current.Data == "br" {
			builder.WriteString("\n")
		}

		for child := current.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}

	walk(node)
	return builder.String()
}

func appendBlock(builder *strings.Builder, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}

	if builder.Len() > 0 && !strings.HasSuffix(builder.String(), "\n\n") {
		builder.WriteString("\n\n")
	}
	builder.WriteString(text)
}

func normalizeURL(raw string, baseURL string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}
	if strings.HasPrefix(raw, "/") {
		return baseURL + raw
	}
	return raw
}

func walkNodes(node *html.Node, visit func(*html.Node)) {
	if node == nil {
		return
	}

	visit(node)
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		walkNodes(child, visit)
	}
}

func findFirstDescendant(node *html.Node, tag string) *html.Node {
	var found *html.Node
	walkNodes(node, func(child *html.Node) {
		if found != nil {
			return
		}
		if child.Type == html.ElementNode && child.Data == tag {
			found = child
		}
	})
	return found
}

func findFirstDescendantWithAttr(node *html.Node, tag, key, value string) *html.Node {
	var found *html.Node
	walkNodes(node, func(child *html.Node) {
		if found != nil {
			return
		}
		if child.Type == html.ElementNode && child.Data == tag && attrValue(child, key) == value {
			found = child
		}
	})
	return found
}

func attrValue(node *html.Node, key string) string {
	if node == nil {
		return ""
	}

	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}

	return ""
}

func hasTopic(topics string, wanted string) bool {
	for _, topic := range strings.FieldsFunc(topics, func(r rune) bool {
		return r == ',' || r == ' '
	}) {
		if topic == wanted {
			return true
		}
	}
	return false
}

func collapseWhitespace(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func fetchSourceEntries(tool *models.Tool, etag string) ([]SourceEntry, string, error) {
	switch tool.SourceType {
	case "github":
		return fetchGitHubEntries(tool.SourceRepo, os.Getenv("GITHUB_TOKEN"), etag)
	case "openai-changelog":
		return fetchOpenAIChangelogEntries(tool.SourceRepo, etag)
	case "opencode-changelog":
		return fetchOpenCodeChangelogEntries(etag)
	default:
		return nil, "", fmt.Errorf("unsupported source type %s", tool.SourceType)
	}
}

func upsertEntry(tool *models.Tool, item SourceEntry) error {
	contentHash := computeHash(item.BodyMD)

	var entry models.Entry
	result := database.DB.Where("tool_id = ? AND source_entry_id = ?", tool.ID, item.SourceEntryID).Limit(1).Find(&entry)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected > 0 {
		if entry.ContentHash == contentHash &&
			entry.Title == item.Title &&
			entry.Version == item.Version &&
			entry.URL == item.URL &&
			entry.IsPrerelease == item.IsPrerelease &&
			entry.PublishedAt.Equal(item.PublishedAt) &&
			entry.SourceUpdatedAt.Equal(item.SourceUpdatedAt) {
			return nil
		}

		entry.Version = item.Version
		entry.Title = item.Title
		entry.URL = item.URL
		entry.BodyMD = item.BodyMD
		entry.PublishedAt = item.PublishedAt
		entry.SourceUpdatedAt = item.SourceUpdatedAt
		entry.ContentHash = contentHash
		entry.IsPrerelease = item.IsPrerelease
		entry.UpdatedAt = time.Now()
		return database.DB.Save(&entry).Error
	}

	entry = models.Entry{
		ToolID:          tool.ID,
		SourceEntryID:   item.SourceEntryID,
		Version:         item.Version,
		Title:           item.Title,
		URL:             item.URL,
		BodyMD:          item.BodyMD,
		PublishedAt:     item.PublishedAt,
		SourceUpdatedAt: item.SourceUpdatedAt,
		ContentHash:     contentHash,
		IsPrerelease:    item.IsPrerelease,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	return database.DB.Create(&entry).Error
}

func SyncTool(tool *models.Tool, fullSync bool) error {
	var syncState models.SyncState
	database.DB.FirstOrCreate(&syncState, models.SyncState{ToolID: tool.ID})

	now := time.Now()
	syncState.LastAttemptAt = &now

	items, newEtag, err := fetchSourceEntries(tool, syncState.EtagPage1)
	if err != nil {
		syncState.LastError = err.Error()
		database.DB.Save(&syncState)
		return err
	}

	if len(items) == 0 && newEtag != "" {
		syncState.EtagPage1 = newEtag
		database.DB.Save(&syncState)
		return nil
	}

	for _, item := range items {
		if err := upsertEntry(tool, item); err != nil {
			syncState.LastError = err.Error()
			database.DB.Save(&syncState)
			return err
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

func SyncAll(fullSync bool) error {
	var tools []models.Tool
	if err := database.DB.Where("is_active = ?", 1).Find(&tools).Error; err != nil {
		return err
	}

	var syncErrors []error
	for _, tool := range tools {
		if err := SyncTool(&tool, fullSync); err != nil {
			syncErrors = append(syncErrors, fmt.Errorf("%s: %w", tool.Slug, err))
		}
	}

	if len(syncErrors) > 0 {
		return errors.Join(syncErrors...)
	}

	return nil
}

func desiredTools() []models.Tool {
	return []models.Tool{
		{Slug: "claude-code", Name: "Claude Code", SourceType: "github", SourceRepo: "anthropics/claude-code", Homepage: "https://claude.ai/code", IsActive: 1, CreatedAt: time.Now()},
		{Slug: "codex-app", Name: "Codex App", SourceType: "openai-changelog", SourceRepo: "codex-app", Homepage: "https://developers.openai.com/codex/app", IsActive: 1, CreatedAt: time.Now()},
		{Slug: "codex-cli", Name: "Codex CLI", SourceType: "openai-changelog", SourceRepo: "codex-cli", Homepage: "https://developers.openai.com/codex/cli", IsActive: 1, CreatedAt: time.Now()},
		{Slug: "gemini-cli", Name: "Gemini CLI", SourceType: "github", SourceRepo: "google-gemini/gemini-cli", Homepage: "https://github.com/google-gemini/gemini-cli", IsActive: 1, CreatedAt: time.Now()},
		{Slug: "opencode", Name: "OpenCode", SourceType: "opencode-changelog", SourceRepo: "opencode", Homepage: "https://opencode.ai", IsActive: 1, CreatedAt: time.Now()},
	}
}

func InitTools() {
	for _, tool := range desiredTools() {
		var existing models.Tool
		result := database.DB.Where("slug = ?", tool.Slug).Limit(1).Find(&existing)
		if result.Error != nil {
			continue
		}

		if result.RowsAffected == 0 {
			database.DB.Create(&tool)
			continue
		}

		if existing.SourceType != tool.SourceType || existing.SourceRepo != tool.SourceRepo {
			database.DB.Where("tool_id = ?", existing.ID).Delete(&models.Entry{})
			database.DB.Where("tool_id = ?", existing.ID).Delete(&models.SyncState{})
		}

		existing.Name = tool.Name
		existing.SourceType = tool.SourceType
		existing.SourceRepo = tool.SourceRepo
		existing.Homepage = tool.Homepage
		existing.IsActive = 1
		database.DB.Save(&existing)
	}

	var legacyCodex models.Tool
	result := database.DB.Where("slug = ?", "codex").Limit(1).Find(&legacyCodex)
	if result.Error == nil && result.RowsAffected > 0 && legacyCodex.IsActive != 0 {
		legacyCodex.IsActive = 0
		database.DB.Save(&legacyCodex)
	}
}

func EnsureSeeded() error {
	InitTools()

	var activeToolCount int64
	if err := database.DB.Model(&models.Tool{}).Where("is_active = ?", 1).Count(&activeToolCount).Error; err != nil {
		return err
	}

	var seededToolCount int64
	if err := database.DB.Model(&models.Entry{}).
		Joins("JOIN tools ON tools.id = entries.tool_id").
		Where("tools.is_active = ?", 1).
		Distinct("entries.tool_id").
		Count(&seededToolCount).Error; err != nil {
		return err
	}

	if activeToolCount > 0 && seededToolCount >= activeToolCount {
		return nil
	}

	return SyncAll(true)
}
