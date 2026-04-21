package panels

import (
	"fmt"
	"os"
	"time"

	"github.com/ifuryst/daymine/packages/workspace"
)

type Service struct {
	Store *workspace.Store
}

type Response struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	UpdatedAt string `json:"updated_at"`
	Data      any    `json:"data"`
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

	data, err := s.data(panel)
	if err != nil {
		return Response{}, err
	}
	return Response{
		ID:        panel.ID,
		Type:      panel.Type,
		Title:     panel.Title,
		UpdatedAt: time.Now().Format(time.RFC3339),
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
