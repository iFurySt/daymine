package panels

import (
	"testing"

	"github.com/ifuryst/daymine/packages/workspace"
)

func TestGetDefaultPanels(t *testing.T) {
	store, err := workspace.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Init(); err != nil {
		t.Fatal(err)
	}

	service := NewService(store)
	resp, err := service.Get("feed")
	if err != nil {
		t.Fatal(err)
	}
	if resp.Type != "feed" {
		t.Fatalf("expected feed response, got %+v", resp)
	}
	data := resp.Data.(map[string]any)
	if len(data["items"].([]any)) == 0 {
		t.Fatal("expected feed items")
	}
}

func TestGetHTMLTemplatePanel(t *testing.T) {
	store, err := workspace.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Init(); err != nil {
		t.Fatal(err)
	}

	service := NewService(store)
	resp, err := service.Get("external-signal")
	if err != nil {
		t.Fatal(err)
	}
	if resp.Renderer == nil || resp.Renderer.Type != "html-template" {
		t.Fatalf("expected html-template renderer, got %+v", resp.Renderer)
	}
	if resp.Renderer.Template == "" {
		t.Fatal("expected loaded template")
	}
	data := resp.Data.(map[string]any)
	items := data["items"].([]any)
	if len(items) == 0 {
		t.Fatal("expected bound items")
	}
}

func TestGetHackerNewsPanelWithoutDigest(t *testing.T) {
	store, err := workspace.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Init(); err != nil {
		t.Fatal(err)
	}

	service := NewService(store)
	resp, err := service.Get("hacker-news-top")
	if err != nil {
		t.Fatal(err)
	}
	if resp.Type != "hacker-news-top" {
		t.Fatalf("expected HN response, got %+v", resp)
	}
	data := resp.Data.(map[string]any)
	if len(data["items"].([]any)) != 0 {
		t.Fatalf("expected empty HN items before first digest, got %+v", data)
	}
}
