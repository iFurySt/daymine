package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultWorkspaceRootUsesUserHome(t *testing.T) {
	root := defaultWorkspaceRoot()
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("user home is unavailable: %v", err)
	}

	want := filepath.Join(home, ".daymine")
	if root != want {
		t.Fatalf("expected %q, got %q", want, root)
	}
}

func TestLocalURL(t *testing.T) {
	tests := []struct {
		name string
		addr string
		want string
	}{
		{name: "port only", addr: ":6345", want: "http://localhost:6345"},
		{name: "localhost", addr: "localhost:7345", want: "http://localhost:7345"},
		{name: "loopback", addr: "127.0.0.1:7345", want: "http://127.0.0.1:7345"},
		{name: "all interfaces", addr: "0.0.0.0:7345", want: "http://localhost:7345"},
		{name: "ipv6 all interfaces", addr: "[::]:7345", want: "http://localhost:7345"},
		{name: "ipv6 loopback", addr: "[::1]:7345", want: "http://[::1]:7345"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := localURL(tt.addr); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
