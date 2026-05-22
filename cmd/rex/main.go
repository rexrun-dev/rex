package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"rexrun.dev/rex/internal/completion"
	"rexrun.dev/rex/internal/detect"
	"rexrun.dev/rex/internal/display"
	"rexrun.dev/rex/internal/envfile"
	"rexrun.dev/rex/internal/generate"
	"rexrun.dev/rex/internal/watcher"
)

const version = "0.3.0"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s\n", display.Red("error:"), err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return showOverview()
	}

	switch args[0] {
	case "version", "--version", "-v":
		fmt.Printf("rex %s %s/%s\n", version, runtime.GOOS, runtime.GOARCH)
		return nil
	case "help", "--help", "-h":
		printHelp()
		return nil
	case "--list":
		return showList()
	case "--dry-run":
		if len(args) < 2 {
			return fmt.Errorf("usage: rex --dry-run <verb>")
		}
		return dryRun(args[1])
	case "doctor":
		return runDoctor()
	case "fresh":
		return runFresh()
	case "clone":
		if len(args) < 2 {
			return fmt.Errorf("usage: rex clone <url> [dir]")
		}
		dir := ""
		if len(args) > 2 {
			dir = args[2]
		}
		return runClone(args[1], dir)
	case "init":
		return runInit()
	case "completion":
		shell := "bash"
		if len(args) > 1 {
			shell = args[1]
		}
		return runCompletion(shell)
	case "watch":
		verb := detect.VerbTest
		if len(args) > 1 {
			verb = detect.Verb(args[1])
		}
		return runWatch(verb)
	case "ci":
		return runCI()
	}

	verb := detect.Verb(args[0])
	extra := args[1:]
	return execute(verb, extra)
}

func showOverview() error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	d := detect.Detect(root)

	// If no stack at root, check for monorepo
	if d.Stack == "" {
		subs := detect.DetectMonorepo(root)
		if len(subs) > 0 {
			display.MonorepoOverview(root, subs)
			return nil
		}
	}

	display.Overview(root, d)
	return nil
}

func showList() error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	d := detect.Detect(root)
	if len(d.Commands) == 0 {
		return fmt.Errorf("no commands detected in this directory")
	}
	for verb, cmd := range d.Commands {
		fmt.Printf("  %s %s\n", display.Cyan(fmt.Sprintf("rex %-6s", verb)), display.Dim("→ "+cmd))
	}
	return nil
}

func dryRun(verbStr string) error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	d := detect.Detect(root)
	verb := detect.Verb(verbStr)
	cmd, ok := d.Commands[verb]
	if !ok {
		return fmt.Errorf("no command detected for %q", verbStr)
	}
	fmt.Printf("%s %s\n", display.Dim("would run:"), cmd)
	return nil
}

func execute(verb detect.Verb, extra []string) error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}

	// Load .env file before executing any command
	envfile.Load(root)

	d := detect.Detect(root)
	cmd, ok := d.Commands[verb]
	if !ok {
		return verbNotFound(verb, d)
	}

	// Append extra args
	if len(extra) > 0 {
		if extra[0] == "--" {
			extra = extra[1:]
		}
		cmd = cmd + " " + strings.Join(extra, " ")
	}

	fmt.Printf("%s %s\n", display.Arrow(), cmd)

	return execCommand(cmd, root)
}

func execCommand(cmdStr, dir string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", cmdStr)
	} else {
		cmd = exec.Command("sh", "-c", cmdStr)
	}
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return err
	}
	return nil
}

func verbNotFound(verb detect.Verb, d detect.Detection) error {
	if d.Stack == "" {
		return fmt.Errorf("could not detect project type in this directory")
	}
	available := make([]string, 0, len(d.Commands))
	for v := range d.Commands {
		available = append(available, string(v))
	}
	return fmt.Errorf("no %q command detected for %s project\navailable: %s",
		verb, d.Stack, strings.Join(available, ", "))
}

func runDoctor() error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	checks := detect.Doctor(root)
	fmt.Println()
	fmt.Printf("  %s %s\n", display.Orange("🦖"), display.Bold("rex doctor"))
	fmt.Println()
	allOK := true
	for _, c := range checks {
		icon := display.Green("✓")
		if !c.OK {
			icon = display.Red("✗")
			allOK = false
		}
		line := fmt.Sprintf("  %s %-20s %s", icon, c.Label, display.Dim(c.Detail))
		fmt.Println(line)
		if !c.OK && c.Fix != "" {
			fmt.Printf("    %s %s\n", display.Dim("fix:"), display.Cyan(c.Fix))
		}
	}
	fmt.Println()
	if allOK {
		fmt.Printf("  %s\n\n", display.Green("All checks passed. Ready to go."))
	} else {
		fmt.Printf("  %s\n\n", display.Yellow("Some issues found. Fix them and run rex doctor again."))
	}
	return nil
}

func runFresh() error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	d := detect.Detect(root)

	steps := []detect.Verb{detect.VerbClean, detect.VerbDeps, detect.VerbBuild}
	for _, step := range steps {
		cmd, ok := d.Commands[step]
		if !ok {
			continue
		}
		fmt.Printf("%s %s %s\n", display.Arrow(), display.Bold(string(step)), display.Dim(cmd))
		if err := execCommand(cmd, root); err != nil {
			return fmt.Errorf("%s failed: %w", step, err)
		}
	}
	fmt.Printf("\n%s fresh complete\n", display.Green("✓"))
	return nil
}

func runInit() error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}

	// Check if rex.toml already exists
	if _, err := os.Stat(filepath.Join(root, "rex.toml")); err == nil {
		return fmt.Errorf("rex.toml already exists in this directory")
	}

	d := detect.Detect(root)
	if d.Stack == "" {
		return fmt.Errorf("no project detected — nothing to initialize")
	}

	content := detect.GenerateConfig(d)
	if err := os.WriteFile(filepath.Join(root, "rex.toml"), []byte(content), 0644); err != nil {
		return err
	}

	fmt.Printf("%s created rex.toml (%s project)\n", display.Green("✓"), d.Stack)
	fmt.Printf("%s edit commands to customize, then commit for your team\n", display.Dim("hint:"))
	return nil
}

func runClone(url, dir string) error {
	// Determine target directory
	if dir == "" {
		// Extract repo name from URL
		parts := strings.Split(strings.TrimSuffix(url, ".git"), "/")
		dir = parts[len(parts)-1]
	}

	fmt.Printf("%s cloning %s\n", display.Arrow(), display.Bold(url))
	if err := execCommand("git clone "+url+" "+dir, "."); err != nil {
		return fmt.Errorf("clone failed: %w", err)
	}

	// Resolve absolute path
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	// Detect the project
	d := detect.Detect(absDir)
	if d.Stack == "" {
		fmt.Printf("\n%s cloned to %s (no stack detected)\n", display.Green("✓"), dir)
		return nil
	}

	fmt.Printf("\n%s detected %s project\n", display.Green("✓"), display.Cyan(d.Stack))

	// Auto-install dependencies if detected
	if depsCmd, ok := d.Commands[detect.VerbDeps]; ok {
		fmt.Printf("%s %s %s\n", display.Arrow(), display.Bold("deps"), display.Dim(depsCmd))
		if err := execCommand(depsCmd, absDir); err != nil {
			fmt.Printf("%s deps failed (non-fatal): %v\n", display.Yellow("⚠"), err)
		}
	}

	fmt.Printf("\n%s ready! cd %s && rex run\n", display.Green("✓"), dir)
	return nil
}

func runCI() error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	d := detect.Detect(root)
	if d.Stack == "" {
		return fmt.Errorf("no stack detected — cannot generate CI config")
	}

	yml := generate.CI(&d)
	dir := filepath.Join(root, ".github", "workflows")
	path := filepath.Join(dir, "ci.yml")

	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("ci.yml already exists — remove it first or edit manually")
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(yml), 0o644); err != nil {
		return err
	}

	fmt.Printf("%s created .github/workflows/ci.yml (%s project)\n", display.Green("✓"), d.Stack)
	return nil
}

func runWatch(verb detect.Verb) error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}

	envfile.Load(root)

	d := detect.Detect(root)
	cmd, ok := d.Commands[verb]
	if !ok {
		return fmt.Errorf("no command detected for %q", verb)
	}

	interval := watcher.DetectInterval(root)
	fileCount := watcher.FileCount(root)

	fmt.Println()
	fmt.Printf("  %s watching %s (%s)\n", display.Orange("🦖"), display.Bold(filepath.Base(root)), d.Stack)
	fmt.Printf("  %s  %s\n", display.Dim("files"), fileCount)
	fmt.Printf("  %s  rex %s → %s\n", display.Dim("run"), display.Cyan(string(verb)), cmd)
	fmt.Printf("  %s  %s\n", display.Dim("poll"), interval)
	fmt.Println()
	fmt.Printf("  %s %s\n", display.Dim("waiting for changes..."), display.Dim("(ctrl+c to stop)"))
	fmt.Println()

	// Run once immediately
	fmt.Printf("  %s %s %s\n", display.Dim(watcher.FormatTime(time.Now())), display.Arrow(), cmd)
	_ = execCommand(cmd, root)

	watcher.Watch(root, interval, func() {
		fmt.Println()
		fmt.Printf("  %s %s detected, re-running...\n", display.Dim(watcher.FormatTime(time.Now())), display.Orange("change"))
		fmt.Printf("  %s %s\n", display.Arrow(), cmd)
		_ = execCommand(cmd, root)
	})

	return nil
}

func runCompletion(shell string) error {
	switch shell {
	case "bash":
		fmt.Print(completion.Bash())
	case "zsh":
		fmt.Print(completion.Zsh())
	case "fish":
		fmt.Print(completion.Fish())
	default:
		return fmt.Errorf("unsupported shell: %s (use bash, zsh, or fish)", shell)
	}
	return nil
}

func printHelp() {
	fmt.Print(`rex — run anything, know nothing.

Usage:
  rex              show project overview
  rex test         run tests
  rex run          start the app
  rex build        build the project
  rex deps         install dependencies
  rex clean        remove build artifacts
  rex fresh        clean + deps + build
  rex fmt          format code
  rex lint         lint code
  rex watch [verb] watch files and re-run on change (default: test)
  rex ci           generate GitHub Actions CI for your stack
  rex clone <url>  clone + detect + install deps
  rex init         generate rex.toml from detected commands
  rex doctor       diagnose environment

Flags:
  rex --list       show all detected commands
  rex --dry-run    show what would run without executing
  rex -v           show version

Pass extra args after --:
  rex test -- -v -run TestFoo

Zero config. Detects your stack automatically.
`)
}
