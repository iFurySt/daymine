package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/ifuryst/daymine/apps/daymine/internal/server"
	"github.com/ifuryst/daymine/packages/workspace"
)

func main() {
	addr := flag.String("addr", ":6345", "HTTP listen address")
	workspaceRoot := flag.String("workspace", defaultWorkspaceRoot(), "Daymine workspace directory")
	scheduler := flag.Bool("scheduler", false, "run scheduled Daymine tasks")
	flag.Parse()

	store, err := workspace.Open(*workspaceRoot)
	if err != nil {
		log.Fatal(err)
	}
	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	srv, err := server.New(server.Options{Store: store, Logger: logger})
	if err != nil {
		log.Fatal(err)
	}

	if *scheduler {
		srv.StartScheduler(context.Background(), 1*time.Hour)
	}

	logger.Info("starting daymine", "addr", *addr, "workspace", store.Root)
	if err := http.ListenAndServe(*addr, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}

func defaultWorkspaceRoot() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ".daymine"
	}
	return home + string(os.PathSeparator) + ".daymine"
}
