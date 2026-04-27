package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
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

	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}
	url := localURL(listener.Addr().String())
	logger.Info("starting daymine", "addr", *addr, "url", url, "workspace", store.Root)
	httpServer := &http.Server{Handler: srv.Handler()}
	if err := httpServer.Serve(listener); err != nil {
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

func localURL(addr string) string {
	host, port := splitListenAddr(addr)
	if host == "" || host == "0.0.0.0" || host == "::" || host == "[::]" {
		host = "localhost"
	}
	if strings.Contains(host, ":") && !strings.HasPrefix(host, "[") {
		host = "[" + host + "]"
	}
	return "http://" + host + ":" + port
}

func splitListenAddr(addr string) (string, string) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return "localhost", "6345"
	}
	if strings.HasPrefix(addr, ":") {
		return "", addr[1:]
	}
	host, port, err := net.SplitHostPort(addr)
	if err == nil {
		return strings.Trim(host, "[]"), port
	}
	if _, err := strconv.Atoi(addr); err == nil {
		return "", addr
	}
	return "localhost", addr
}
