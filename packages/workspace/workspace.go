package workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Store struct {
	Root string
}

type DashboardConfig struct {
	Pages []Page `json:"pages"`
}

type Page struct {
	Name           string     `json:"name"`
	Title          string     `json:"title"`
	ColumnWidths   []int      `json:"column_widths"`
	LayoutByColumn [][]string `json:"layout_by_column"`
	Panels         []Panel    `json:"panels"`
}

type Panel struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Title    string         `json:"title"`
	Refresh  string         `json:"refresh,omitempty"`
	Source   string         `json:"source,omitempty"`
	Renderer *Renderer      `json:"renderer,omitempty"`
	Data     *DataSource    `json:"data,omitempty"`
	Config   map[string]any `json:"config,omitempty"`
}

type Renderer struct {
	Type         string            `json:"type"`
	Variant      string            `json:"variant,omitempty"`
	Template     string            `json:"template,omitempty"`
	TemplatePath string            `json:"template_path,omitempty"`
	Style        string            `json:"style,omitempty"`
	StylePath    string            `json:"style_path,omitempty"`
	Fields       map[string]string `json:"fields,omitempty"`
	Config       map[string]any    `json:"config,omitempty"`
}

type DataSource struct {
	Type     string `json:"type"`
	Path     string `json:"path,omitempty"`
	Selector string `json:"selector,omitempty"`
	As       string `json:"as,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

type RunRecord struct {
	ID          string    `json:"id"`
	Provider    string    `json:"provider"`
	Query       string    `json:"query"`
	Status      string    `json:"status"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
	ExitCode    int       `json:"exit_code,omitempty"`
	Output      string    `json:"output,omitempty"`
	Error       string    `json:"error,omitempty"`
	Artifacts   []string  `json:"artifacts,omitempty"`
}

func Open(root string) (*Store, error) {
	if root == "" {
		return nil, errors.New("workspace root is required")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve workspace root: %w", err)
	}
	return &Store{Root: abs}, nil
}

func (s *Store) Init() error {
	dirs := []string{
		"config/panels",
		"inbox/rss",
		"inbox/web",
		"inbox/social",
		"inbox/manual",
		"notes/daily",
		"notes/topics",
		"notes/sources",
		"artifacts/runs",
		"artifacts/scripts",
		"artifacts/attachments",
		"index",
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(s.Path(dir), 0o755); err != nil {
			return fmt.Errorf("create workspace dir %s: %w", dir, err)
		}
	}

	writes := map[string][]byte{
		"config/daymine.json": mustJSON(DefaultDashboardConfig()),
		"index/panels.json":   mustJSON(DefaultPanelData()),
		"config/panels/external-signal.template.html": []byte(`<dm-list>
  <dm-item data-for="item in items">
    <dm-link href="{{ item.url }}">{{ item.title }}</dm-link>
    <dm-text tone="muted" max-lines="3">{{ item.summary }}</dm-text>
    <dm-meta>{{ item.source }} · {{ item.published_at }}</dm-meta>
  </dm-item>
</dm-list>
`),
		"config/panels/external-signal.panel.css": []byte(`dm-item {
  border-color: hsl(43 50% 24%);
}

dm-link {
  color: hsl(43 50% 72%);
}
`),
		"notes/daily/welcome.md": []byte(`# Welcome to Daymine

Daymine keeps local, long-lived notes and Agent artifacts in this workspace.

- Edit Markdown directly when that is faster.
- Let Agents write scripts and summaries into this tree.
- Use the dashboard to scan what changed today.
`),
	}
	for rel, data := range writes {
		if err := writeFileIfMissing(s.Path(rel), data); err != nil {
			return err
		}
	}
	if err := s.ensureDefaultHTMLTemplatePanel(); err != nil {
		return err
	}
	if err := s.ensureDefaultPanelData(); err != nil {
		return err
	}
	runs := s.Path("index/runs.jsonl")
	if _, err := os.Stat(runs); errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile(runs, []byte{}, 0o644); err != nil {
			return fmt.Errorf("create run log: %w", err)
		}
	}
	return nil
}

func (s *Store) ensureDefaultPanelData() error {
	index, err := s.PanelIndex()
	if err != nil {
		return err
	}
	changed := false
	for key, value := range DefaultPanelData() {
		if _, ok := index[key]; !ok {
			index[key] = value
			changed = true
		}
	}
	if !changed {
		return nil
	}
	return os.WriteFile(s.Path("index/panels.json"), mustJSON(index), 0o644)
}

func (s *Store) ensureDefaultHTMLTemplatePanel() error {
	cfg, err := s.DashboardConfig()
	if err != nil {
		return err
	}
	if len(cfg.Pages) == 0 {
		return nil
	}
	page := &cfg.Pages[0]
	for _, panel := range page.Panels {
		if panel.ID == "external-signal" {
			return nil
		}
	}
	page.Panels = append(page.Panels, defaultExternalSignalPanel())
	if len(page.LayoutByColumn) == 0 {
		page.LayoutByColumn = [][]string{{"external-signal"}}
	} else {
		last := len(page.LayoutByColumn) - 1
		page.LayoutByColumn[last] = append([]string{"external-signal"}, page.LayoutByColumn[last]...)
	}
	return os.WriteFile(s.Path("config/daymine.json"), mustJSON(cfg), 0o644)
}

func (s *Store) Path(rel string) string {
	clean := filepath.Clean(rel)
	if clean == "." {
		return s.Root
	}
	return filepath.Join(s.Root, clean)
}

func (s *Store) Relative(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(s.Root, abs)
	if err != nil {
		return "", err
	}
	if rel == "." {
		return ".", nil
	}
	if strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." {
		return "", fmt.Errorf("path %s is outside workspace", path)
	}
	return filepath.ToSlash(rel), nil
}

func (s *Store) DashboardConfig() (DashboardConfig, error) {
	var cfg DashboardConfig
	if err := readJSON(s.Path("config/daymine.json"), &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func (s *Store) PanelIndex() (map[string]any, error) {
	var data map[string]any
	if err := readJSON(s.Path("index/panels.json"), &data); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *Store) AppendRun(record RunRecord) error {
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(s.Path("index/runs.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open run log: %w", err)
	}
	defer f.Close()
	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("append run log: %w", err)
	}
	return nil
}

func (s *Store) Runs(limit int) ([]RunRecord, error) {
	data, err := os.ReadFile(s.Path("index/runs.jsonl"))
	if err != nil {
		return nil, fmt.Errorf("read run log: %w", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	records := make([]RunRecord, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var record RunRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return nil, fmt.Errorf("decode run log: %w", err)
		}
		records = append(records, record)
	}
	for i, j := 0, len(records)-1; i < j; i, j = i+1, j-1 {
		records[i], records[j] = records[j], records[i]
	}
	if limit > 0 && len(records) > limit {
		records = records[:limit]
	}
	return records, nil
}

func (s *Store) Snapshot() (map[string]struct{}, error) {
	files := map[string]struct{}{}
	err := filepath.WalkDir(s.Root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := s.Relative(path)
		if err != nil {
			return err
		}
		files[rel] = struct{}{}
		return nil
	})
	return files, err
}

func (s *Store) NewFiles(before map[string]struct{}) ([]string, error) {
	after, err := s.Snapshot()
	if err != nil {
		return nil, err
	}
	var files []string
	for rel := range after {
		if _, ok := before[rel]; !ok {
			files = append(files, rel)
		}
	}
	return files, nil
}

func DefaultDashboardConfig() DashboardConfig {
	return DashboardConfig{Pages: []Page{{
		Name:           "home",
		Title:          "Home",
		ColumnWidths:   []int{1, 1, 1},
		LayoutByColumn: [][]string{{"calendar", "feed"}, {"article-list", "github-list"}, {"external-signal", "agent-runs", "markdown-view"}},
		Panels: []Panel{
			{ID: "calendar", Type: "calendar", Title: "Calendar", Refresh: "1h"},
			{ID: "feed", Type: "feed", Title: "RSS Feed", Refresh: "15m", Source: "index/panels.json"},
			{ID: "article-list", Type: "article-list", Title: "Articles", Source: "index/panels.json"},
			{ID: "github-list", Type: "github-list", Title: "GitHub", Source: "index/panels.json"},
			defaultExternalSignalPanel(),
			{ID: "agent-runs", Type: "agent-runs", Title: "Agent Runs", Refresh: "10s", Source: "index/runs.jsonl"},
			{ID: "markdown-view", Type: "markdown-view", Title: "Welcome", Source: "notes/daily/welcome.md"},
		},
	}}}
}

func DefaultPanelData() map[string]any {
	return map[string]any{
		"feed": []map[string]any{
			{"title": "Daymine workspace initialized", "source": "local", "url": "", "summary": "The local filesystem workspace is ready for Agent-maintained knowledge.", "published_at": time.Now().Format(time.RFC3339)},
			{"title": "Panel manifests are file-backed", "source": "system", "url": "", "summary": "Panels read structured data from local JSON and Markdown files.", "published_at": time.Now().Format(time.RFC3339)},
		},
		"article-list": []map[string]any{
			{"title": "Why Daymine is FS-first", "path": "notes/daily/welcome.md", "status": "draft", "tags": []string{"architecture", "local-first"}},
		},
		"github-list": []map[string]any{
			{"name": "daymine", "full_name": "ifuryst/daymine", "description": "Self-hosted Agent-maintained information dashboard.", "stars": 0, "language": "Go"},
		},
		"external-signal": []map[string]any{
			{"title": "HTML fragment panel is live", "source": "workspace template", "url": "", "summary": "This panel is rendered from config/panels/external-signal.template.html and data-bound from index/panels.json.", "published_at": time.Now().Format(time.RFC3339)},
			{"title": "AI can edit this UI at runtime", "source": "Daymine DSL", "url": "", "summary": "Change the template, style, or data in the workspace and refresh the panel without rebuilding the binary.", "published_at": time.Now().Format(time.RFC3339)},
		},
	}
}

func defaultExternalSignalPanel() Panel {
	return Panel{
		ID:      "external-signal",
		Type:    "html-template",
		Title:   "External Panel",
		Refresh: "15m",
		Renderer: &Renderer{
			Type:         "html-template",
			TemplatePath: "config/panels/external-signal.template.html",
			StylePath:    "config/panels/external-signal.panel.css",
		},
		Data: &DataSource{
			Type:     "json",
			Path:     "index/panels.json",
			Selector: "$.external-signal",
			As:       "items",
			Limit:    5,
		},
	}
}

func readJSON(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}

func writeFileIfMissing(path string, data []byte) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func mustJSON(value any) []byte {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		panic(err)
	}
	return append(data, '\n')
}
