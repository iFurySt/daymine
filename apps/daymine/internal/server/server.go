package server

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/ifuryst/daymine/apps/daymine/internal/webassets"
	"github.com/ifuryst/daymine/packages/agent"
	"github.com/ifuryst/daymine/packages/panels"
	"github.com/ifuryst/daymine/packages/tasks"
	"github.com/ifuryst/daymine/packages/workspace"
)

type Server struct {
	store  *workspace.Store
	panels *panels.Service
	agents *agent.Controller
	tasks  *tasks.Service
	assets fs.FS
	logger *slog.Logger
}

type Options struct {
	Store  *workspace.Store
	Logger *slog.Logger
}

func New(opts Options) (*Server, error) {
	if opts.Store == nil {
		return nil, errors.New("workspace store is required")
	}
	assets, err := fs.Sub(webassets.Files, "dist")
	if err != nil {
		return nil, err
	}
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}
	agents := agent.NewController(opts.Store, agent.LocalCommandProvider{}, agent.CodexCLIProvider{})
	return &Server{
		store:  opts.Store,
		panels: panels.NewService(opts.Store),
		agents: agents,
		tasks:  tasks.NewService(opts.Store, agents),
		assets: assets,
		logger: logger,
	}, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", s.health)
	mux.HandleFunc("GET /api/v1/workspace", s.workspaceInfo)
	mux.HandleFunc("GET /api/v1/dashboard/config", s.dashboardConfig)
	mux.HandleFunc("GET /api/v1/panels", s.panelList)
	mux.HandleFunc("GET /api/v1/panels/", s.panelData)
	mux.HandleFunc("GET /api/v1/agent/runs", s.agentRuns)
	mux.HandleFunc("POST /api/v1/agent/runs", s.startAgentRun)
	mux.HandleFunc("GET /api/v1/tasks", s.taskList)
	mux.HandleFunc("POST /api/v1/tasks/", s.startTaskRun)
	mux.HandleFunc("/", s.static)
	return s.logging(mux)
}

func (s *Server) StartScheduler(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = time.Hour
	}
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		s.tasks.RunDue(ctx, s.logger)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.tasks.RunDue(ctx, s.logger)
			}
		}
	}()
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "message": "Daymine API is running"})
}

func (s *Server) workspaceInfo(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"root": s.store.Root})
}

func (s *Server) dashboardConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := s.store.DashboardConfig()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, cfg)
}

func (s *Server) panelList(w http.ResponseWriter, r *http.Request) {
	items, err := s.panels.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"panels": items})
}

func (s *Server) panelData(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/panels/")
	if id == "" || strings.Contains(id, "/") {
		writeError(w, http.StatusNotFound, errors.New("panel not found"))
		return
	}
	resp, err := s.panels.Get(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) agentRuns(w http.ResponseWriter, r *http.Request) {
	runs, err := s.store.Runs(50)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"runs": runs})
}

func (s *Server) startAgentRun(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Provider string `json:"provider"`
		Query    string `json:"query"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Minute)
	defer cancel()
	record, err := s.agents.Run(ctx, req.Provider, req.Query)
	status := http.StatusCreated
	if err != nil {
		status = http.StatusBadGateway
	}
	writeJSON(w, status, map[string]any{"run": record})
}

func (s *Server) taskList(w http.ResponseWriter, r *http.Request) {
	items, err := s.tasks.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tasks": items})
}

func (s *Server) startTaskRun(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/tasks/")
	id = strings.TrimSuffix(id, "/runs")
	if id == "" || strings.Contains(id, "/") || !strings.HasSuffix(r.URL.Path, "/runs") {
		writeError(w, http.StatusNotFound, errors.New("task not found"))
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()
	record, err := s.tasks.Run(ctx, id)
	status := http.StatusCreated
	if err != nil {
		status = http.StatusBadGateway
	}
	writeJSON(w, status, map[string]any{"run": record})
}

func (s *Server) static(w http.ResponseWriter, r *http.Request) {
	clean := path.Clean(strings.TrimPrefix(r.URL.Path, "/"))
	if clean == "." || clean == "/" {
		clean = "index.html"
	}
	file, err := s.assets.Open(clean)
	if err != nil {
		clean = "index.html"
	} else {
		_ = file.Close()
	}
	http.ServeFileFS(w, r, s.assets, clean)
}

func (s *Server) logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		s.logger.Info("http request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start).String())
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]any{"error": err.Error()})
}
