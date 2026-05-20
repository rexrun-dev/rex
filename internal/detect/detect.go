package detect

import (
	"os"
	"path/filepath"
	"strings"
)

// Verb represents a standard project action.
type Verb string

const (
	VerbTest  Verb = "test"
	VerbRun   Verb = "run"
	VerbBuild Verb = "build"
	VerbDeps  Verb = "deps"
	VerbClean Verb = "clean"
	VerbFmt   Verb = "fmt"
	VerbLint  Verb = "lint"
)

// Detection holds the result of analyzing a project root.
type Detection struct {
	Stack      string            // e.g. "go", "node", "python", "rust"
	PkgMgr    string            // e.g. "pnpm", "npm", "yarn", "bun", "pip", "uv", "cargo"
	Commands  map[Verb]string   // resolved commands per verb
	Frameworks []string
}

// Detect analyzes the project at root and returns the best detection.
func Detect(root string) Detection {
	d := Detection{Commands: make(map[Verb]string)}

	// Priority 1: task runners (Justfile, Makefile) — override language defaults
	if resolved := detectTaskRunner(root); resolved != nil {
		mergeCommands(d.Commands, resolved)
	}

	switch {
	case exists(root, "go.mod"):
		d.Stack = "go"
		d.PkgMgr = "go"
		setIfMissing(d.Commands, VerbTest, "go test ./...")
		setIfMissing(d.Commands, VerbBuild, "go build ./...")
		setIfMissing(d.Commands, VerbRun, detectGoRun(root))
		setIfMissing(d.Commands, VerbDeps, "go mod download")
		setIfMissing(d.Commands, VerbFmt, "gofmt -w .")
		setIfMissing(d.Commands, VerbLint, "go vet ./...")
		setIfMissing(d.Commands, VerbClean, "go clean -cache")

	case exists(root, "package.json"):
		d.Stack = "node"
		d.PkgMgr = detectNodePkgMgr(root)
		run := d.PkgMgr + " run"
		if d.PkgMgr == "bun" {
			run = "bun run"
		}
		setIfMissing(d.Commands, VerbTest, d.PkgMgr+" test")
		setIfMissing(d.Commands, VerbRun, run+" dev")
		setIfMissing(d.Commands, VerbBuild, run+" build")
		setIfMissing(d.Commands, VerbDeps, d.PkgMgr+" install")
		setIfMissing(d.Commands, VerbFmt, run+" format || "+run+" fmt || npx prettier --write .")
		setIfMissing(d.Commands, VerbLint, run+" lint || npx eslint .")
		setIfMissing(d.Commands, VerbClean, "rm -rf node_modules dist .next .nuxt build out")

	case exists(root, "pyproject.toml"), exists(root, "requirements.txt"), exists(root, "setup.py"):
		d.Stack = "python"
		d.PkgMgr = detectPythonPkgMgr(root)
		setIfMissing(d.Commands, VerbTest, "pytest")
		setIfMissing(d.Commands, VerbRun, "python -m "+filepath.Base(root))
		setIfMissing(d.Commands, VerbBuild, "python -m build")
		setIfMissing(d.Commands, VerbDeps, pythonDepsCmd(d.PkgMgr))
		setIfMissing(d.Commands, VerbFmt, "black . || ruff format .")
		setIfMissing(d.Commands, VerbLint, "ruff check . || pylint .")
		setIfMissing(d.Commands, VerbClean, "rm -rf __pycache__ .pytest_cache dist build *.egg-info .venv")

	case exists(root, "Cargo.toml"):
		d.Stack = "rust"
		d.PkgMgr = "cargo"
		setIfMissing(d.Commands, VerbTest, "cargo test")
		setIfMissing(d.Commands, VerbRun, "cargo run")
		setIfMissing(d.Commands, VerbBuild, "cargo build")
		setIfMissing(d.Commands, VerbDeps, "cargo fetch")
		setIfMissing(d.Commands, VerbFmt, "cargo fmt")
		setIfMissing(d.Commands, VerbLint, "cargo clippy")
		setIfMissing(d.Commands, VerbClean, "cargo clean")

	case exists(root, "composer.json"):
		d.Stack = "php"
		d.PkgMgr = "composer"
		if exists(root, "artisan") {
			d.Frameworks = append(d.Frameworks, "laravel")
			setIfMissing(d.Commands, VerbTest, "php artisan test")
			setIfMissing(d.Commands, VerbRun, "php artisan serve")
			setIfMissing(d.Commands, VerbBuild, "php artisan optimize")
		} else {
			setIfMissing(d.Commands, VerbTest, "./vendor/bin/phpunit")
			setIfMissing(d.Commands, VerbRun, "php -S localhost:8000")
			setIfMissing(d.Commands, VerbBuild, "composer dump-autoload --optimize")
		}
		setIfMissing(d.Commands, VerbDeps, "composer install")
		setIfMissing(d.Commands, VerbFmt, "./vendor/bin/pint || ./vendor/bin/php-cs-fixer fix .")
		setIfMissing(d.Commands, VerbLint, "./vendor/bin/phpstan analyse")
		setIfMissing(d.Commands, VerbClean, "php artisan cache:clear 2>/dev/null; rm -rf vendor")

	case exists(root, "build.zig"):
		d.Stack = "zig"
		d.PkgMgr = "zig"
		setIfMissing(d.Commands, VerbTest, "zig build test")
		setIfMissing(d.Commands, VerbRun, "zig build run")
		setIfMissing(d.Commands, VerbBuild, "zig build")
		setIfMissing(d.Commands, VerbClean, "zig build --help 2>/dev/null; rm -rf zig-out zig-cache .zig-cache")

	case exists(root, "mix.exs"):
		d.Stack = "elixir"
		d.PkgMgr = "mix"
		if exists(root, "config") && exists(root, "lib") {
			if exists(root, "assets") {
				d.Frameworks = append(d.Frameworks, "phoenix")
				setIfMissing(d.Commands, VerbRun, "mix phx.server")
			}
		}
		setIfMissing(d.Commands, VerbTest, "mix test")
		setIfMissing(d.Commands, VerbRun, "iex -S mix")
		setIfMissing(d.Commands, VerbBuild, "mix compile")
		setIfMissing(d.Commands, VerbDeps, "mix deps.get")
		setIfMissing(d.Commands, VerbFmt, "mix format")
		setIfMissing(d.Commands, VerbLint, "mix credo")
		setIfMissing(d.Commands, VerbClean, "mix clean")
	}

	// Priority 0 (applied last = wins): rex.toml explicit overrides
	if cfg := LoadConfig(root); cfg != nil {
		mergeCommands(d.Commands, cfg)
	}

	return d
}

func detectTaskRunner(root string) map[Verb]string {
	// Justfile
	if exists(root, "Justfile") || exists(root, "justfile") {
		return parseJustfile(root)
	}
	// Makefile
	if exists(root, "Makefile") || exists(root, "makefile") {
		return parseMakefile(root)
	}
	return nil
}

func detectNodePkgMgr(root string) string {
	switch {
	case exists(root, "bun.lockb"), exists(root, "bun.lock"):
		return "bun"
	case exists(root, "pnpm-lock.yaml"):
		return "pnpm"
	case exists(root, "yarn.lock"):
		return "yarn"
	default:
		return "npm"
	}
}

func detectPythonPkgMgr(root string) string {
	if exists(root, "uv.lock") {
		return "uv"
	}
	if exists(root, "poetry.lock") {
		return "poetry"
	}
	if exists(root, "Pipfile.lock") {
		return "pipenv"
	}
	return "pip"
}

func pythonDepsCmd(mgr string) string {
	switch mgr {
	case "uv":
		return "uv sync"
	case "poetry":
		return "poetry install"
	case "pipenv":
		return "pipenv install"
	default:
		return "pip install -r requirements.txt"
	}
}

func detectGoRun(root string) string {
	// Look for cmd/ directory with a main package
	cmd := filepath.Join(root, "cmd")
	entries, err := os.ReadDir(cmd)
	if err == nil && len(entries) > 0 {
		for _, e := range entries {
			if e.IsDir() {
				return "go run ./cmd/" + e.Name()
			}
		}
	}
	return "go run ."
}

func parseMakefile(root string) map[Verb]string {
	b, err := os.ReadFile(filepath.Join(root, "Makefile"))
	if err != nil {
		b, err = os.ReadFile(filepath.Join(root, "makefile"))
		if err != nil {
			return nil
		}
	}
	cmds := make(map[Verb]string)
	verbMap := map[string]Verb{
		"test": VerbTest, "run": VerbRun, "build": VerbBuild,
		"dev": VerbRun, "start": VerbRun, "serve": VerbRun,
		"install": VerbDeps, "deps": VerbDeps,
		"fmt": VerbFmt, "format": VerbFmt,
		"lint": VerbLint, "check": VerbLint,
		"clean": VerbClean,
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasSuffix(line, ":") && !strings.HasPrefix(line, ".") && !strings.HasPrefix(line, "#") {
			target := strings.TrimSuffix(line, ":")
			target = strings.TrimSpace(target)
			if v, ok := verbMap[strings.ToLower(target)]; ok {
				cmds[v] = "make " + target
			}
		}
	}
	return cmds
}

func parseJustfile(root string) map[Verb]string {
	path := filepath.Join(root, "Justfile")
	b, err := os.ReadFile(path)
	if err != nil {
		b, err = os.ReadFile(filepath.Join(root, "justfile"))
		if err != nil {
			return nil
		}
	}
	cmds := make(map[Verb]string)
	verbMap := map[string]Verb{
		"test": VerbTest, "run": VerbRun, "build": VerbBuild,
		"dev": VerbRun, "start": VerbRun, "serve": VerbRun,
		"install": VerbDeps, "deps": VerbDeps,
		"fmt": VerbFmt, "format": VerbFmt,
		"lint": VerbLint, "check": VerbLint,
		"clean": VerbClean,
	}
	for _, line := range strings.Split(string(b), "\n") {
		if len(line) == 0 || line[0] == ' ' || line[0] == '\t' || line[0] == '#' {
			continue
		}
		if idx := strings.Index(line, ":"); idx > 0 {
			recipe := strings.TrimSpace(line[:idx])
			if v, ok := verbMap[strings.ToLower(recipe)]; ok {
				cmds[v] = "just " + recipe
			}
		}
	}
	return cmds
}

func mergeCommands(dst, src map[Verb]string) {
	for k, v := range src {
		dst[k] = v
	}
}

func setIfMissing(m map[Verb]string, k Verb, v string) {
	if _, ok := m[k]; !ok && v != "" {
		m[k] = v
	}
}

func exists(root, name string) bool {
	_, err := os.Stat(filepath.Join(root, name))
	return err == nil
}
