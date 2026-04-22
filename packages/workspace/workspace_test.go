package workspace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestInitCreatesDefaultWorkspace(t *testing.T) {
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Init(); err != nil {
		t.Fatal(err)
	}

	for _, rel := range []string{"config/daymine.json", "index/panels.json", "index/runs.jsonl", "notes/daily/welcome.md"} {
		if _, err := os.Stat(store.Path(rel)); err != nil {
			t.Fatalf("expected %s: %v", rel, err)
		}
	}

	cfg, err := store.DashboardConfig()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Pages) != 1 || cfg.Pages[0].Panels[0].ID != "calendar" {
		t.Fatalf("unexpected dashboard config: %+v", cfg)
	}
}

func TestInitAddsDefaultHTMLTemplatePanelToExistingConfig(t *testing.T) {
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(store.Path("config"), 0o755); err != nil {
		t.Fatal(err)
	}
	legacy := DashboardConfig{Pages: []Page{{
		Name:           "home",
		Title:          "Home",
		ColumnWidths:   []int{1},
		LayoutByColumn: [][]string{{"calendar"}},
		Panels:         []Panel{{ID: "calendar", Type: "calendar", Title: "Calendar"}},
	}}}
	data, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(store.Path("config/daymine.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := store.Init(); err != nil {
		t.Fatal(err)
	}
	cfg, err := store.DashboardConfig()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, panel := range cfg.Pages[0].Panels {
		if panel.ID == "external-signal" && panel.Renderer != nil && panel.Renderer.Type == "html-template" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected migrated external-signal panel: %+v", cfg.Pages[0].Panels)
	}
	if cfg.Pages[0].LayoutByColumn[0][0] != "external-signal" {
		t.Fatalf("expected external panel inserted into layout: %+v", cfg.Pages[0].LayoutByColumn)
	}
}

func TestAppendAndReadRunsNewestFirst(t *testing.T) {
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Init(); err != nil {
		t.Fatal(err)
	}

	old := RunRecord{ID: "old", Provider: "local-command", Query: "one", Status: "completed", StartedAt: time.Now()}
	newer := RunRecord{ID: "new", Provider: "local-command", Query: "two", Status: "completed", StartedAt: time.Now()}
	if err := store.AppendRun(old); err != nil {
		t.Fatal(err)
	}
	if err := store.AppendRun(newer); err != nil {
		t.Fatal(err)
	}

	runs, err := store.Runs(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 1 || runs[0].ID != "new" {
		t.Fatalf("expected newest run first, got %+v", runs)
	}
}

func TestRelativeRejectsOutsidePath(t *testing.T) {
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.Relative(filepath.Dir(store.Root))
	if err == nil {
		t.Fatal("expected outside path error")
	}
}
