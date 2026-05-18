package detect

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDetectGo(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/app\n\ngo 1.23\n")
	os.MkdirAll(filepath.Join(root, "cmd", "server"), 0o755)

	d := Detect(root)

	if d.Stack != "go" {
		t.Fatalf("expected stack=go, got %q", d.Stack)
	}
	if d.PkgMgr != "go" {
		t.Fatalf("expected pkgmgr=go, got %q", d.PkgMgr)
	}
	if d.Commands[VerbTest] != "go test ./..." {
		t.Fatalf("test cmd: %q", d.Commands[VerbTest])
	}
	if d.Commands[VerbRun] != "go run ./cmd/server" {
		t.Fatalf("run cmd: %q", d.Commands[VerbRun])
	}
	if d.Commands[VerbDeps] != "go mod download" {
		t.Fatalf("deps cmd: %q", d.Commands[VerbDeps])
	}
	if d.Commands[VerbFmt] != "gofmt -w ." {
		t.Fatalf("fmt cmd: %q", d.Commands[VerbFmt])
	}
}

func TestDetectNodePnpm(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "package.json"), `{"scripts":{"test":"vitest","dev":"vite","build":"tsc"}}`)
	writeFile(t, filepath.Join(root, "pnpm-lock.yaml"), "lockfileVersion: 9\n")

	d := Detect(root)

	if d.Stack != "node" {
		t.Fatalf("expected stack=node, got %q", d.Stack)
	}
	if d.PkgMgr != "pnpm" {
		t.Fatalf("expected pkgmgr=pnpm, got %q", d.PkgMgr)
	}
	if d.Commands[VerbTest] != "pnpm test" {
		t.Fatalf("test cmd: %q", d.Commands[VerbTest])
	}
	if d.Commands[VerbDeps] != "pnpm install" {
		t.Fatalf("deps cmd: %q", d.Commands[VerbDeps])
	}
}

func TestDetectNodeBun(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "package.json"), `{"scripts":{"test":"bun test"}}`)
	writeFile(t, filepath.Join(root, "bun.lockb"), "")

	d := Detect(root)

	if d.PkgMgr != "bun" {
		t.Fatalf("expected pkgmgr=bun, got %q", d.PkgMgr)
	}
}

func TestDetectPythonUv(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "pyproject.toml"), "[project]\nname = \"myapp\"\n")
	writeFile(t, filepath.Join(root, "uv.lock"), "")

	d := Detect(root)

	if d.Stack != "python" {
		t.Fatalf("expected stack=python, got %q", d.Stack)
	}
	if d.PkgMgr != "uv" {
		t.Fatalf("expected pkgmgr=uv, got %q", d.PkgMgr)
	}
	if d.Commands[VerbDeps] != "uv sync" {
		t.Fatalf("deps cmd: %q", d.Commands[VerbDeps])
	}
}

func TestDetectRust(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "Cargo.toml"), "[package]\nname = \"myapp\"\n")

	d := Detect(root)

	if d.Stack != "rust" {
		t.Fatalf("expected stack=rust, got %q", d.Stack)
	}
	if d.Commands[VerbTest] != "cargo test" {
		t.Fatalf("test cmd: %q", d.Commands[VerbTest])
	}
	if d.Commands[VerbFmt] != "cargo fmt" {
		t.Fatalf("fmt cmd: %q", d.Commands[VerbFmt])
	}
	if d.Commands[VerbClean] != "cargo clean" {
		t.Fatalf("clean cmd: %q", d.Commands[VerbClean])
	}
}

func TestDetectMakefile(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/app\n\ngo 1.23\n")
	writeFile(t, filepath.Join(root, "Makefile"), "test:\n\tgo test -race ./...\n\nlint:\n\tgolangci-lint run\n\nclean:\n\trm -rf bin/\n")

	d := Detect(root)

	// Makefile takes priority via fallback chain
	if d.Commands[VerbTest] != "make test" {
		t.Fatalf("expected Makefile test override, got %q", d.Commands[VerbTest])
	}
	if d.Commands[VerbLint] != "make lint" {
		t.Fatalf("expected Makefile lint override, got %q", d.Commands[VerbLint])
	}
	if d.Commands[VerbClean] != "make clean" {
		t.Fatalf("expected Makefile clean override, got %q", d.Commands[VerbClean])
	}
	// Non-overridden verbs fall through to Go detection
	if d.Commands[VerbDeps] != "go mod download" {
		t.Fatalf("deps should fallback to go: %q", d.Commands[VerbDeps])
	}
}

func TestDetectJustfile(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "Cargo.toml"), "[package]\nname = \"app\"\n")
	writeFile(t, filepath.Join(root, "justfile"), "test:\n  cargo test --all\n\nfmt:\n  cargo fmt --all\n")

	d := Detect(root)

	if d.Commands[VerbTest] != "just test" {
		t.Fatalf("expected justfile test, got %q", d.Commands[VerbTest])
	}
	if d.Commands[VerbFmt] != "just fmt" {
		t.Fatalf("expected justfile fmt, got %q", d.Commands[VerbFmt])
	}
	// Fallback to Rust for non-specified verbs
	if d.Commands[VerbBuild] != "cargo build" {
		t.Fatalf("build should fallback to cargo: %q", d.Commands[VerbBuild])
	}
}

func TestDetectEmpty(t *testing.T) {
	root := t.TempDir()

	d := Detect(root)

	if d.Stack != "" {
		t.Fatalf("expected empty stack, got %q", d.Stack)
	}
	if len(d.Commands) != 0 {
		t.Fatalf("expected no commands, got %v", d.Commands)
	}
}
