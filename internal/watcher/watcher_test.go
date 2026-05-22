package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTakeSnapshot(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(dir, "util.go"), []byte("package util"), 0644)

	snap := Take(dir)
	if len(snap) != 2 {
		t.Errorf("expected 2 files in snapshot, got %d", len(snap))
	}
}

func TestSnapshotIgnoresGit(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git", "objects"), 0755)
	os.WriteFile(filepath.Join(dir, ".git", "objects", "abc"), []byte("data"), 0644)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)

	snap := Take(dir)
	if len(snap) != 1 {
		t.Errorf("expected 1 file (ignoring .git), got %d", len(snap))
	}
}

func TestSnapshotIgnoresNodeModules(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "node_modules", "pkg"), 0755)
	os.WriteFile(filepath.Join(dir, "node_modules", "pkg", "index.js"), []byte("x"), 0644)
	os.WriteFile(filepath.Join(dir, "index.js"), []byte("console.log()"), 0644)

	snap := Take(dir)
	if len(snap) != 1 {
		t.Errorf("expected 1 file (ignoring node_modules), got %d", len(snap))
	}
}

func TestSnapshotIgnoresBinaryExts(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(dir, "app.exe"), []byte("binary"), 0644)
	os.WriteFile(filepath.Join(dir, "lib.so"), []byte("binary"), 0644)

	snap := Take(dir)
	if len(snap) != 1 {
		t.Errorf("expected 1 file (ignoring binaries), got %d", len(snap))
	}
}

func TestChangedDetectsNew(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	a := Take(dir)

	os.WriteFile(filepath.Join(dir, "new.go"), []byte("package new"), 0644)
	b := Take(dir)

	if !Changed(a, b) {
		t.Error("expected change after adding file")
	}
}

func TestChangedDetectsModification(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("v1"), 0644)
	a := Take(dir)

	time.Sleep(50 * time.Millisecond)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("v2-changed"), 0644)
	b := Take(dir)

	if !Changed(a, b) {
		t.Error("expected change after modifying file")
	}
}

func TestChangedNoChange(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	a := Take(dir)
	b := Take(dir)

	if Changed(a, b) {
		t.Error("expected no change")
	}
}

func TestCountFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.go"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(dir, "b.go"), []byte("b"), 0644)
	os.WriteFile(filepath.Join(dir, "c.exe"), []byte("c"), 0644)

	n := CountFiles(dir)
	if n != 2 {
		t.Errorf("expected 2 watched files, got %d", n)
	}
}

func TestFileCount(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.go"), []byte("a"), 0644)
	s := FileCount(dir)
	if s != "1 file" {
		t.Errorf("expected '1 file', got %q", s)
	}

	os.WriteFile(filepath.Join(dir, "b.go"), []byte("b"), 0644)
	s = FileCount(dir)
	if s != "2 files" {
		t.Errorf("expected '2 files', got %q", s)
	}
}

func TestDetectInterval(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.go"), []byte("a"), 0644)
	interval := DetectInterval(dir)
	if interval != 500*time.Millisecond {
		t.Errorf("expected 500ms for small project, got %v", interval)
	}
}

func TestIgnoreFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"src/main.go", false},
		{"node_modules/pkg/index.js", true},
		{".git/objects/abc", true},
		{"vendor/lib/file.go", true},
		{"app.exe", true},
		{"lib.so", true},
		{"main.py", false},
	}
	for _, tt := range tests {
		got := IgnoreFile(tt.path)
		if got != tt.want {
			t.Errorf("IgnoreFile(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestWatchedExtensions(t *testing.T) {
	exts := WatchedExtensions("go")
	if len(exts) == 0 {
		t.Error("expected Go extensions")
	}
	found := false
	for _, e := range exts {
		if e == ".go" {
			found = true
		}
	}
	if !found {
		t.Error("expected .go in Go extensions")
	}

	unknown := WatchedExtensions("unknown-stack")
	if unknown != nil {
		t.Errorf("expected nil for unknown stack, got %v", unknown)
	}
}
