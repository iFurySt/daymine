package tasks

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ifuryst/daymine/packages/agent"
	"github.com/ifuryst/daymine/packages/workspace"
)

type fakeProvider struct{}

func (fakeProvider) Name() string {
	return "codex-cli"
}

func (fakeProvider) Run(_ context.Context, req agent.Request) (agent.Result, error) {
	target := filepath.Join(req.Workspace, "index", "hacker-news", "top10-latest.json")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return agent.Result{}, err
	}
	if err := os.WriteFile(target, []byte(`{"items":[]}`), 0o644); err != nil {
		return agent.Result{}, err
	}
	return agent.Result{Output: "wrote HN digest"}, nil
}

func TestDefaultTasksIncludesHackerNewsDailyTask(t *testing.T) {
	tasks := DefaultTasks()
	if len(tasks) != 1 {
		t.Fatalf("expected one default task, got %d", len(tasks))
	}
	task := tasks[0]
	if task.ID != "hacker-news-daily-top10" {
		t.Fatalf("unexpected task: %+v", task)
	}
	if task.Provider != "codex-cli" || task.Schedule != "daily" {
		t.Fatalf("unexpected provider or schedule: %+v", task)
	}
	if !strings.Contains(task.Prompt, "Hacker News Firebase API") {
		t.Fatalf("prompt does not describe the HN source: %s", task.Prompt)
	}
}

func TestRunTaskRecordsTaskIDAndArtifacts(t *testing.T) {
	store, err := workspace.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Init(); err != nil {
		t.Fatal(err)
	}

	agents := agent.NewController(store, fakeProvider{})
	service := NewService(store, agents)
	record, err := service.Run(context.Background(), "hacker-news-daily-top10")
	if err != nil {
		t.Fatal(err)
	}
	if record.TaskID != "hacker-news-daily-top10" {
		t.Fatalf("expected task id on run record, got %+v", record)
	}
	if record.Provider != "codex-cli" || record.Status != "completed" {
		t.Fatalf("unexpected run record: %+v", record)
	}
	if len(record.Artifacts) == 0 {
		t.Fatalf("expected artifact discovery, got %+v", record)
	}
}

func TestListIncludesLastRunAndNextRun(t *testing.T) {
	store, err := workspace.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Init(); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 4, 22, 12, 0, 0, 0, time.Local)
	last := workspace.RunRecord{
		ID:        "run-1",
		TaskID:    "hacker-news-daily-top10",
		Provider:  "codex-cli",
		Query:     "collect hn",
		Status:    "completed",
		StartedAt: now.Add(-2 * time.Hour),
	}
	if err := store.AppendRun(last); err != nil {
		t.Fatal(err)
	}
	service := NewService(store, agent.NewController(store, fakeProvider{}))
	service.Now = func() time.Time { return now }

	views, err := service.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(views) != 1 || views[0].LastRun == nil || views[0].LastRun.ID != "run-1" {
		t.Fatalf("expected last run in task view, got %+v", views)
	}
	if views[0].NextRunAt == "" {
		t.Fatalf("expected next run time, got %+v", views[0])
	}
}
