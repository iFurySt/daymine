package agent

import (
	"context"
	"os"
	"testing"

	"github.com/ifuryst/daymine/packages/workspace"
)

func TestLocalCommandRunRecordsArtifact(t *testing.T) {
	store, err := workspace.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Init(); err != nil {
		t.Fatal(err)
	}

	controller := NewController(store, LocalCommandProvider{})
	record, err := controller.Run(context.Background(), "local-command", "printf hello > artifacts/runs/hello.txt")
	if err != nil {
		t.Fatal(err)
	}
	if record.Status != "completed" {
		t.Fatalf("expected completed run, got %+v", record)
	}
	if _, err := os.Stat(store.Path("artifacts/runs/hello.txt")); err != nil {
		t.Fatal(err)
	}
	if len(record.Artifacts) == 0 {
		t.Fatalf("expected artifact discovery, got %+v", record)
	}
}
