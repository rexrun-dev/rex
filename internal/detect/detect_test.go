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

func TestDetectRuby(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "Gemfile"), "source 'https://rubygems.org'\n")

	d := Detect(root)

	if d.Stack != "ruby" {
		t.Fatalf("expected stack=ruby, got %q", d.Stack)
	}
	if d.PkgMgr != "bundler" {
		t.Fatalf("expected pkgmgr=bundler, got %q", d.PkgMgr)
	}
	if d.Commands[VerbTest] != "bundle exec rspec" {
		t.Fatalf("test cmd: %q", d.Commands[VerbTest])
	}
	if d.Commands[VerbDeps] != "bundle install" {
		t.Fatalf("deps cmd: %q", d.Commands[VerbDeps])
	}
}

func TestDetectRails(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "Gemfile"), "source 'https://rubygems.org'\ngem 'rails'\n")
	os.MkdirAll(filepath.Join(root, "app"), 0o755)
	os.MkdirAll(filepath.Join(root, "config"), 0o755)
	writeFile(t, filepath.Join(root, "Rakefile"), "")

	d := Detect(root)

	if d.Stack != "ruby" {
		t.Fatalf("expected stack=ruby, got %q", d.Stack)
	}
	if len(d.Frameworks) == 0 || d.Frameworks[0] != "rails" {
		t.Fatalf("expected rails framework, got %v", d.Frameworks)
	}
	if d.Commands[VerbTest] != "bundle exec rails test" {
		t.Fatalf("test cmd: %q", d.Commands[VerbTest])
	}
	if d.Commands[VerbRun] != "bundle exec rails server" {
		t.Fatalf("run cmd: %q", d.Commands[VerbRun])
	}
}

func TestDetectJavaMaven(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "pom.xml"), "<project></project>")

	d := Detect(root)

	if d.Stack != "java" {
		t.Fatalf("expected stack=java, got %q", d.Stack)
	}
	if d.PkgMgr != "maven" {
		t.Fatalf("expected pkgmgr=maven, got %q", d.PkgMgr)
	}
	if d.Commands[VerbTest] != "mvn test" {
		t.Fatalf("test cmd: %q", d.Commands[VerbTest])
	}
	if d.Commands[VerbClean] != "mvn clean" {
		t.Fatalf("clean cmd: %q", d.Commands[VerbClean])
	}
}

func TestDetectJavaMavenWrapper(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "pom.xml"), "<project></project>")
	writeFile(t, filepath.Join(root, "mvnw"), "#!/bin/sh\n")

	d := Detect(root)

	if d.Commands[VerbTest] != "./mvnw test" {
		t.Fatalf("expected mvnw, got %q", d.Commands[VerbTest])
	}
}

func TestDetectJavaGradle(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "build.gradle"), "apply plugin: 'java'")

	d := Detect(root)

	if d.Stack != "java" {
		t.Fatalf("expected stack=java, got %q", d.Stack)
	}
	if d.PkgMgr != "gradle" {
		t.Fatalf("expected pkgmgr=gradle, got %q", d.PkgMgr)
	}
	if d.Commands[VerbTest] != "gradle test" {
		t.Fatalf("test cmd: %q", d.Commands[VerbTest])
	}
}

func TestDetectJavaGradleWrapper(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "build.gradle"), "apply plugin: 'java'")
	writeFile(t, filepath.Join(root, "gradlew"), "#!/bin/sh\n")

	d := Detect(root)

	if d.Commands[VerbTest] != "./gradlew test" {
		t.Fatalf("expected gradlew, got %q", d.Commands[VerbTest])
	}
}

func TestDetectPHP(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "composer.json"), "{}")

	d := Detect(root)

	if d.Stack != "php" {
		t.Fatalf("expected stack=php, got %q", d.Stack)
	}
	if d.PkgMgr != "composer" {
		t.Fatalf("expected pkgmgr=composer, got %q", d.PkgMgr)
	}
	if d.Commands[VerbDeps] != "composer install" {
		t.Fatalf("deps cmd: %q", d.Commands[VerbDeps])
	}
}

func TestDetectLaravel(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "composer.json"), "{}")
	writeFile(t, filepath.Join(root, "artisan"), "#!/usr/bin/env php\n")

	d := Detect(root)

	if d.Stack != "php" {
		t.Fatalf("expected stack=php, got %q", d.Stack)
	}
	if len(d.Frameworks) == 0 || d.Frameworks[0] != "laravel" {
		t.Fatalf("expected laravel framework, got %v", d.Frameworks)
	}
	if d.Commands[VerbTest] != "php artisan test" {
		t.Fatalf("test cmd: %q", d.Commands[VerbTest])
	}
}

func TestDetectZig(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "build.zig"), "const std = @import(\"std\");\n")

	d := Detect(root)

	if d.Stack != "zig" {
		t.Fatalf("expected stack=zig, got %q", d.Stack)
	}
	if d.Commands[VerbTest] != "zig build test" {
		t.Fatalf("test cmd: %q", d.Commands[VerbTest])
	}
	if d.Commands[VerbBuild] != "zig build" {
		t.Fatalf("build cmd: %q", d.Commands[VerbBuild])
	}
}

func TestDetectElixir(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "mix.exs"), "defmodule MyApp do\nend\n")

	d := Detect(root)

	if d.Stack != "elixir" {
		t.Fatalf("expected stack=elixir, got %q", d.Stack)
	}
	if d.PkgMgr != "mix" {
		t.Fatalf("expected pkgmgr=mix, got %q", d.PkgMgr)
	}
	if d.Commands[VerbTest] != "mix test" {
		t.Fatalf("test cmd: %q", d.Commands[VerbTest])
	}
	if d.Commands[VerbDeps] != "mix deps.get" {
		t.Fatalf("deps cmd: %q", d.Commands[VerbDeps])
	}
}

func TestDetectElixirPhoenix(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "mix.exs"), "defmodule MyApp do\nend\n")
	os.MkdirAll(filepath.Join(root, "assets"), 0o755)
	os.MkdirAll(filepath.Join(root, "config"), 0o755)
	os.MkdirAll(filepath.Join(root, "lib"), 0o755)

	d := Detect(root)

	if len(d.Frameworks) == 0 || d.Frameworks[0] != "phoenix" {
		t.Fatalf("expected phoenix framework, got %v", d.Frameworks)
	}
	if d.Commands[VerbRun] != "mix phx.server" {
		t.Fatalf("run cmd: %q", d.Commands[VerbRun])
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
