package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"rexrun.dev/rex/internal/detect"
	"rexrun.dev/rex/internal/display"
)

const version = "0.1.0"

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

Flags:
  rex --list       show all detected commands
  rex --dry-run    show what would run without executing
  rex -v           show version

Pass extra args after --:
  rex test -- -v -run TestFoo

Zero config. Detects your stack automatically.
`)
}
