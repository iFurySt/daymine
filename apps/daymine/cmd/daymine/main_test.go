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
