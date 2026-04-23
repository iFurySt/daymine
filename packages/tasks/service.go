package tasks

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/ifuryst/daymine/packages/agent"
	"github.com/ifuryst/daymine/packages/workspace"
)

type Task struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Provider    string   `json:"provider"`
	Schedule    string   `json:"schedule,omitempty"`
	Prompt      string   `json:"-"`
	Artifacts   []string `json:"artifacts"`
	PanelIDs    []string `json:"panel_ids,omitempty"`
}

type TaskView struct {
	Task
	LastRun   *workspace.RunRecord `json:"last_run,omitempty"`
	NextRunAt string               `json:"next_run_at,omitempty"`
}

type Service struct {
	Store  *workspace.Store
	Agents *agent.Controller
	Now    func() time.Time
}

func NewService(store *workspace.Store, agents *agent.Controller) *Service {
	return &Service{Store: store, Agents: agents, Now: time.Now}
}

func (s *Service) List() ([]TaskView, error) {
	runs, err := s.Store.Runs(0)
	if err != nil {
		return nil, err
	}
	var views []TaskView
	for _, task := range DefaultTasks() {
		view := TaskView{Task: task}
		if last := lastRunForTask(runs, task.ID); last != nil {
			copied := *last
			view.LastRun = &copied
		}
		if due := nextRunAt(task, view.LastRun, s.Now()); !due.IsZero() {
			view.NextRunAt = due.Format(time.RFC3339)
		}
		views = append(views, view)
	}
	return views, nil
}

func (s *Service) Get(id string) (Task, error) {
	for _, task := range DefaultTasks() {
		if task.ID == id {
			return task, nil
		}
	}
	return Task{}, fmt.Errorf("task %q not found", id)
}

func (s *Service) Run(ctx context.Context, id string) (workspace.RunRecord, error) {
	task, err := s.Get(id)
	if err != nil {
		return workspace.RunRecord{}, err
	}
	return s.Agents.Run(ctx, task.Provider, task.Prompt, agent.WithTaskID(task.ID), agent.WithTimeout(10*time.Minute))
}

func (s *Service) RunDue(ctx context.Context, logger *slog.Logger) {
	tasks := DefaultTasks()
	runs, err := s.Store.Runs(0)
	if err != nil {
		logger.Error("read task runs", "error", err)
		return
	}
	now := s.Now()
	for _, task := range tasks {
		if !isDue(task, lastRunForTask(runs, task.ID), now) {
			continue
		}
		logger.Info("running scheduled task", "task", task.ID, "provider", task.Provider)
		if _, err := s.Run(ctx, task.ID); err != nil {
			logger.Error("scheduled task failed", "task", task.ID, "error", err)
		}
	}
}

func DefaultTasks() []Task {
	return []Task{{
		ID:          "hacker-news-daily-top10",
		Title:       "Hacker News daily top 10",
		Description: "Collect the top Hacker News stories from the last 24 hours and persist a daily digest.",
		Provider:    "codex-cli",
		Schedule:    "daily",
		Prompt:      hackerNewsDailyPrompt,
		Artifacts: []string{
			"index/hacker-news/top10-latest.json",
			"index/hacker-news/YYYY-MM-DD-top10.json",
			"notes/sources/hacker-news/YYYY-MM-DD-top10.md",
		},
		PanelIDs: []string{"hacker-news-top", "agent-runs"},
	}}
}

func lastRunForTask(runs []workspace.RunRecord, taskID string) *workspace.RunRecord {
	for i := range runs {
		if runs[i].TaskID == taskID {
			return &runs[i]
		}
	}
	return nil
}

func isDue(task Task, last *workspace.RunRecord, now time.Time) bool {
	if task.Schedule != "daily" {
		return false
	}
	if last == nil {
		return true
	}
	lastDay := last.StartedAt.Local().Format("2006-01-02")
	return lastDay != now.Local().Format("2006-01-02")
}

func nextRunAt(task Task, last *workspace.RunRecord, now time.Time) time.Time {
	if task.Schedule != "daily" {
		return time.Time{}
	}
	if last == nil {
		return now
	}
	next := time.Date(last.StartedAt.Local().Year(), last.StartedAt.Local().Month(), last.StartedAt.Local().Day()+1, 0, 0, 0, 0, now.Local().Location())
	if next.Before(now) {
		return now
	}
	return next
}

var hackerNewsDailyPrompt = strings.TrimSpace(`
You are running as a Daymine Codex task inside the Daymine workspace.

Goal:
- Fetch Hacker News top stories from the last 24 hours.
- Keep the top 10 stories by score.
- Persist the result as workspace knowledge that Daymine can render later.

Data source:
- Use the public Hacker News Firebase API:
  - https://hacker-news.firebaseio.com/v0/topstories.json
  - https://hacker-news.firebaseio.com/v0/item/<id>.json

Rules:
- Only write under these paths:
  - index/hacker-news/
  - notes/sources/hacker-news/
- Do not modify source code, configs, or unrelated workspace files.
- Use the current clock when deciding the 24 hour window.
- Filter stories whose Unix "time" is within the last 24 hours.
- Sort filtered stories by "score" descending, then keep the first 10.
- If fewer than 10 stories are in the window, include all matching stories.
- For each story, create a one sentence neutral summary based on title, URL, and HN metadata only. Do not invent details not present in the fetched data.

Write:
1. index/hacker-news/top10-latest.json
2. index/hacker-news/YYYY-MM-DD-top10.json
3. notes/sources/hacker-news/YYYY-MM-DD-top10.md

JSON schema:
{
  "generated_at": "RFC3339 timestamp",
  "window_hours": 24,
  "source": "hacker-news-firebase",
  "items": [
    {
      "rank": 1,
      "id": 123,
      "title": "Story title",
      "url": "https://...",
      "hn_url": "https://news.ycombinator.com/item?id=123",
      "score": 100,
      "by": "author",
      "time": "RFC3339 timestamp",
      "comments": 20,
      "summary": "One neutral sentence."
    }
  ]
}

Markdown:
- Start with "# Hacker News Top 10 - YYYY-MM-DD".
- Include generated time and the 24 hour window.
- Add one numbered item per story with title, score, comments, URL, HN URL, and summary.

After writing the files, print the paths you wrote.
`)
