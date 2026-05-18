package detect

import (
	"path/filepath"
	"testing"
)

func TestRexTomlOverridesEverything(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/app\n\ngo 1.23\n")
	writeFile(t, filepath.Join(root, "Makefile"), "test:\n\tmake test-custom\n")
	writeFile(t, filepath.Join(root, "rex.toml"), `
[commands]
test = "go test -race -count=1 ./..."
run = "air"
`)

	d := Detect(root)

	// rex.toml should win over Makefile and Go defaults
	if d.Commands[VerbTest] != "go test -race -count=1 ./..." {
		t.Fatalf("expected rex.toml override for test, got %q", d.Commands[VerbTest])
	}
	if d.Commands[VerbRun] != "air" {
		t.Fatalf("expected rex.toml override for run, got %q", d.Commands[VerbRun])
	}
	// Non-overridden verbs should still work from Go detection
	if d.Commands[VerbDeps] != "go mod download" {
		t.Fatalf("deps should fallback: %q", d.Commands[VerbDeps])
	}
}

func TestRexTomlEmpty(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "Cargo.toml"), "[package]\nname = \"app\"\n")
	// No rex.toml

	d := Detect(root)

	if d.Commands[VerbTest] != "cargo test" {
		t.Fatalf("should fallback to cargo: %q", d.Commands[VerbTest])
	}
}

func TestParseRexToml(t *testing.T) {
	input := `
# Rex configuration
[commands]
test = "pytest -x --tb=short"
run = "uvicorn main:app --reload"
build = "docker build -t app ."

[meta]
stack = "python"
`
	cmds := parseRexToml(input)
	if cmds[VerbTest] != "pytest -x --tb=short" {
		t.Fatalf("test: %q", cmds[VerbTest])
	}
	if cmds[VerbRun] != "uvicorn main:app --reload" {
		t.Fatalf("run: %q", cmds[VerbRun])
	}
	if cmds[VerbBuild] != "docker build -t app ." {
		t.Fatalf("build: %q", cmds[VerbBuild])
	}
	// [meta] section should not bleed into commands
	if _, ok := cmds[Verb("stack")]; ok {
		t.Fatal("meta section should not be parsed as commands")
	}
}
