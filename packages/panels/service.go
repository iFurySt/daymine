package panels

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ifuryst/daymine/packages/workspace"
)

type Service struct {
	Store *workspace.Store
}

type Response struct {
	ID        string              `json:"id"`
	Type      string              `json:"type"`
	Title     string              `json:"title"`
	UpdatedAt string              `json:"updated_at"`
	Renderer  *workspace.Renderer `json:"renderer,omitempty"`
	Data      any                 `json:"data"`
}

func NewService(store *workspace.Store) *Service {
	return &Service{Store: store}
}

func (s *Service) List() ([]workspace.Panel, error) {
	cfg, err := s.Store.DashboardConfig()
	if err != nil {
		return nil, err
	}
	var panels []workspace.Panel
	for _, page := range cfg.Pages {
		panels = append(panels, page.Panels...)
	}
	return panels, nil
}

func (s *Service) Get(id string) (Response, error) {
	panel, err := s.find(id)
	if err != nil {
		return Response{}, err
	}

	renderer, err := s.renderer(panel)
	if err != nil {
		return Response{}, err
	}
	data, err := s.data(panel)
	if err != nil {
		return Response{}, err
	}
	return Response{
		ID:        panel.ID,
		Type:      panel.Type,
		Title:     panel.Title,
		UpdatedAt: time.Now().Format(time.RFC3339),
		Renderer:  renderer,
		Data:      data,
	}, nil
}

func (s *Service) find(id string) (workspace.Panel, error) {
	panels, err := s.List()
	if err != nil {
		return workspace.Panel{}, err
	}
	for _, panel := range panels {
		if panel.ID == id {
			return panel, nil
		}
	}
	return workspace.Panel{}, fmt.Errorf("panel %q not found", id)
}

func (s *Service) data(panel workspace.Panel) (any, error) {
	if panel.Renderer != nil && panel.Renderer.Type == "html-template" {
		return s.dataSourceContext(panel)
	}
	switch panel.Type {
	case "calendar":
		return map[string]any{
			"today": time.Now().Format("2006-01-02"),
			"events": []map[string]any{
				{"id": "daily-review", "title": "Daily review", "date": time.Now().Format("2006-01-02"), "time": "09:00"},
			},
		}, nil
	case "agent-runs":
		runs, err := s.Store.Runs(20)
		if err != nil {
			return nil, err
		}
		return map[string]any{"runs": runs}, nil
	case "markdown-view":
		if panel.Source == "" {
			return map[string]any{"markdown": ""}, nil
		}
		data, err := os.ReadFile(s.Store.Path(panel.Source))
		if err != nil {
			return nil, err
		}
		return map[string]any{"path": panel.Source, "markdown": string(data)}, nil
	default:
		index, err := s.Store.PanelIndex()
		if err != nil {
			return nil, err
		}
		value, ok := index[panel.ID]
		if !ok {
			value = []any{}
		}
		return map[string]any{"items": value}, nil
	}
}

func (s *Service) renderer(panel workspace.Panel) (*workspace.Renderer, error) {
	if panel.Renderer == nil {
		return nil, nil
	}
	renderer := *panel.Renderer
	if renderer.TemplatePath != "" {
		data, err := os.ReadFile(s.Store.Path(renderer.TemplatePath))
		if err != nil {
			return nil, err
		}
		renderer.Template = string(data)
	}
	if renderer.StylePath != "" {
		data, err := os.ReadFile(s.Store.Path(renderer.StylePath))
		if err != nil {
			return nil, err
		}
		renderer.Style = string(data)
	}
	return &renderer, nil
}

func (s *Service) dataSourceContext(panel workspace.Panel) (map[string]any, error) {
	context := map[string]any{
		"panel": map[string]any{
			"id":    panel.ID,
			"title": panel.Title,
			"type":  panel.Type,
		},
	}
	if panel.Data == nil {
		context["items"] = []any{}
		return context, nil
	}

	switch panel.Data.Type {
	case "json":
		var value any
		data, err := os.ReadFile(s.Store.Path(panel.Data.Path))
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, &value); err != nil {
			return nil, err
		}
		selected := selectValue(value, panel.Data.Selector)
		selected = applyLimit(selected, panel.Data.Limit)
		as := panel.Data.As
		if as == "" {
			as = "items"
		}
		context[as] = selected
		if as != "items" {
			context["items"] = selected
		}
		return context, nil
	default:
		return nil, fmt.Errorf("unsupported data source type %q", panel.Data.Type)
	}
}

func selectValue(value any, selector string) any {
	if selector == "" || selector == "$" {
		return value
	}
	if !strings.HasPrefix(selector, "$.") {
		return value
	}
	current := value
	for _, part := range strings.Split(strings.TrimPrefix(selector, "$."), ".") {
		object, ok := current.(map[string]any)
		if !ok {
			return []any{}
		}
		current = object[part]
	}
	if current == nil {
		return []any{}
	}
	return current
}

func applyLimit(value any, limit int) any {
	if limit <= 0 {
		return value
	}
	items, ok := value.([]any)
	if !ok || len(items) <= limit {
		return value
	}
	return items[:limit]
}
