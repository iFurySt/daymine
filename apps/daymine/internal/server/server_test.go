package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ifuryst/daymine/packages/workspace"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	store, err := workspace.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Init(); err != nil {
		t.Fatal(err)
	}
	srv, err := New(Options{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	return srv
}

func TestHealth(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Daymine API") {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestStartAgentRun(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/agent/runs", strings.NewReader(`{"provider":"local-command","query":"printf ok > artifacts/runs/api.txt"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "completed") {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestTaskList(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "hacker-news-daily-top10") {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}
