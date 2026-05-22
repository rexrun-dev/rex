package watcher

import (
	"crypto/md5"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// IgnoredDirs are directories that should never be watched.
var IgnoredDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
	".next":        true,
	".nuxt":        true,
	"dist":         true,
	"build":        true,
	"target":       true,
	"__pycache__":  true,
	".venv":        true,
	"_build":       true,
	"deps":         true,
	"zig-cache":    true,
	"zig-out":      true,
}

// IgnoredExts are file extensions that should be ignored.
var IgnoredExts = map[string]bool{
	".exe": true,
	".o":   true,
	".a":   true,
	".so":  true,
	".dll": true,
	".dylib": true,
	".pyc": true,
	".class": true,
}

// Snapshot represents the state of all watched files at a point in time.
type Snapshot map[string][16]byte

// Take creates a snapshot of all relevant files under root.
func Take(root string) Snapshot {
	snap := make(Snapshot)
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if IgnoredDirs[name] || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if IgnoredExts[ext] {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		// Use size + modtime as a fast hash proxy
		key := fmt.Sprintf("%d-%d", info.Size(), info.ModTime().UnixNano())
		snap[path] = md5.Sum([]byte(key))
		return nil
	})
	return snap
}

// Changed returns true if two snapshots differ.
func Changed(a, b Snapshot) bool {
	if len(a) != len(b) {
		return true
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok || bv != v {
			return true
		}
	}
	return false
}

// Watch polls the filesystem and calls onChange when files change.
// It blocks until ctx is done (use a cancel context or signal handler).
func Watch(root string, interval time.Duration, onChange func()) {
	prev := Take(root)
	for {
		time.Sleep(interval)
		curr := Take(root)
		if Changed(prev, curr) {
			onChange()
			prev = Take(root) // re-snapshot after build
		} else {
			prev = curr
		}
	}
}

// CountFiles returns the number of watched files.
func CountFiles(root string) int {
	count := 0
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if IgnoredDirs[name] || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if IgnoredExts[ext] {
			return nil
		}
		count++
		return nil
	})
	return count
}

// FileCount returns a human-friendly file count string.
func FileCount(root string) string {
	n := CountFiles(root)
	if n == 1 {
		return "1 file"
	}
	return fmt.Sprintf("%d files", n)
}

// DetectInterval picks polling frequency based on project size.
func DetectInterval(root string) time.Duration {
	n := CountFiles(root)
	switch {
	case n > 5000:
		return 2 * time.Second
	case n > 1000:
		return time.Second
	default:
		return 500 * time.Millisecond
	}
}

// ClearScreen prints ANSI clear sequence.
func ClearScreen() {
	fmt.Print("\033[2J\033[H")
}

// FormatTime formats a time for display.
func FormatTime(t time.Time) string {
	return t.Format("15:04:05")
}

// Debounce returns the last-sent value after a quiet period.
func Debounce(interval time.Duration) func(func()) {
	var timer *time.Timer
	return func(fn func()) {
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(interval, fn)
	}
}

// IgnoreFile checks if a specific file path should be ignored.
func IgnoreFile(path string) bool {
	parts := strings.Split(filepath.ToSlash(path), "/")
	for _, p := range parts {
		if IgnoredDirs[p] {
			return true
		}
	}
	ext := strings.ToLower(filepath.Ext(path))
	return IgnoredExts[ext]
}

// WatchedExtensions returns common source file extensions for a stack.
func WatchedExtensions(stack string) []string {
	switch stack {
	case "go":
		return []string{".go", ".mod", ".sum"}
	case "node":
		return []string{".js", ".ts", ".jsx", ".tsx", ".json", ".css", ".html", ".vue", ".svelte"}
	case "python":
		return []string{".py", ".pyi", ".toml", ".cfg", ".ini"}
	case "rust":
		return []string{".rs", ".toml"}
	case "php":
		return []string{".php", ".blade.php", ".env"}
	case "ruby":
		return []string{".rb", ".erb", ".rake", ".gemspec"}
	case "java":
		return []string{".java", ".kt", ".gradle", ".xml", ".properties"}
	case "zig":
		return []string{".zig"}
	case "elixir":
		return []string{".ex", ".exs", ".eex", ".leex", ".heex"}
	default:
		return nil
	}
}

// Pid writes the current process PID to .rex/watch.pid for external tools.
func Pid(root string) error {
	dir := filepath.Join(root, ".rex")
	os.MkdirAll(dir, 0755)
	return os.WriteFile(filepath.Join(dir, "watch.pid"), []byte(fmt.Sprintf("%d", os.Getpid())), 0644)
}
