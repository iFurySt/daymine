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
