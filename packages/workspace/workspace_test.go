package workspace

import (
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
