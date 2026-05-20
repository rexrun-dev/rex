package detect

import (
	"os"
	"path/filepath"
)

// SubProject represents a detected sub-project in a monorepo.
type SubProject struct {
	Path      string    // relative path from root
	Detection Detection // detection result for this sub-project
}

// DetectMonorepo scans for sub-projects within a root directory.
// Returns nil if the directory is not a monorepo (only one project at root).
func DetectMonorepo(root string) []SubProject {
	var subs []SubProject

	// Common monorepo patterns
	patterns := []string{
		"packages/*",
		"apps/*",
		"services/*",
		"modules/*",
		"libs/*",
		"projects/*",
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(root, pattern))
		if err != nil {
			continue
		}
		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil || !info.IsDir() {
				continue
			}
			d := Detect(match)
			if d.Stack != "" {
				rel, _ := filepath.Rel(root, match)
				subs = append(subs, SubProject{
					Path:      rel,
					Detection: d,
				})
			}
		}
	}

	// Also check if current directory has a workspace file (pnpm-workspace.yaml, etc.)
	if len(subs) == 0 {
		subs = detectWorkspaces(root)
	}

	return subs
}

// IsMonorepo returns true if the directory contains multiple sub-projects.
func IsMonorepo(root string) bool {
	return len(DetectMonorepo(root)) > 1
}

func detectWorkspaces(root string) []SubProject {
	var subs []SubProject

	// pnpm workspaces
	if exists(root, "pnpm-workspace.yaml") {
		return scanDir(root, "packages")
	}

	// Lerna
	if exists(root, "lerna.json") {
		return scanDir(root, "packages")
	}

	// Go workspace
	if exists(root, "go.work") {
		return scanGoWork(root)
	}

	return subs
}

func scanDir(root, subdir string) []SubProject {
	var subs []SubProject
	dir := filepath.Join(root, subdir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(dir, e.Name())
		d := Detect(path)
		if d.Stack != "" {
			rel, _ := filepath.Rel(root, path)
			subs = append(subs, SubProject{Path: rel, Detection: d})
		}
	}
	return subs
}

func scanGoWork(root string) []SubProject {
	var subs []SubProject
	b, err := os.ReadFile(filepath.Join(root, "go.work"))
	if err != nil {
		return nil
	}

	// Simple parser: look for "use" directives
	lines := splitLines(string(b))
	inUse := false
	for _, line := range lines {
		trimmed := trimSpace(line)
		if trimmed == "use (" {
			inUse = true
			continue
		}
		if trimmed == ")" {
			inUse = false
			continue
		}
		if inUse && trimmed != "" {
			modPath := filepath.Join(root, trimmed)
			if info, err := os.Stat(modPath); err == nil && info.IsDir() {
				d := Detect(modPath)
				if d.Stack != "" {
					subs = append(subs, SubProject{Path: trimmed, Detection: d})
				}
			}
		}
	}
	return subs
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			line := s[start:i]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			lines = append(lines, line)
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func trimSpace(s string) string {
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	j := len(s)
	for j > i && (s[j-1] == ' ' || s[j-1] == '\t') {
		j--
	}
	return s[i:j]
}
