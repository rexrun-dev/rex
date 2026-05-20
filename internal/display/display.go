package display

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"rexrun.dev/rex/internal/detect"
)

// ANSI color codes
const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	cyan   = "\033[36m"
	orange = "\033[38;5;208m"
)

func Red(s string) string    { return red + s + reset }
func Green(s string) string  { return green + s + reset }
func Yellow(s string) string { return yellow + s + reset }
func Cyan(s string) string   { return cyan + s + reset }
func Dim(s string) string    { return dim + s + reset }
func Bold(s string) string   { return bold + s + reset }
func Orange(s string) string { return orange + s + reset }

func Arrow() string { return Orange("→") }

// Overview prints the project detection summary.
func Overview(root string, d detect.Detection) {
	name := filepath.Base(root)

	fmt.Println()
	fmt.Printf("  %s %s\n", Orange("🦖"), Bold(name))
	fmt.Println()

	if d.Stack == "" {
		fmt.Printf("  %s No project detected in this directory.\n", Yellow("⚠"))
		fmt.Println()
		fmt.Printf("  Supported: Go, Node, Python, Rust, Java, Docker, Make, Just\n")
		fmt.Println()
		return
	}

	// Stack info
	stackLine := d.Stack
	if d.PkgMgr != "" && d.PkgMgr != d.Stack {
		stackLine += " + " + d.PkgMgr
	}
	if len(d.Frameworks) > 0 {
		stackLine += " (" + strings.Join(d.Frameworks, ", ") + ")"
	}
	fmt.Printf("  %s  %s\n", Dim("stack"), stackLine)
	fmt.Println()

	// Commands
	if len(d.Commands) == 0 {
		return
	}

	fmt.Printf("  %s\n", Dim("commands"))

	verbOrder := []detect.Verb{
		detect.VerbTest, detect.VerbRun, detect.VerbBuild,
		detect.VerbDeps, detect.VerbFmt, detect.VerbLint, detect.VerbClean,
	}

	for _, v := range verbOrder {
		if cmd, ok := d.Commands[v]; ok {
			fmt.Printf("    %s %s %s\n",
				Cyan(fmt.Sprintf("rex %-6s", v)),
				Dim("→"),
				cmd,
			)
		}
	}

	fmt.Println()

	// Hint
	if isFirstRun(root) {
		fmt.Printf("  %s try: %s\n", Dim("hint:"), Cyan("rex test"))
		fmt.Println()
	}
}

func isFirstRun(root string) bool {
	_, err := os.Stat(filepath.Join(root, ".rex"))
	return err != nil
}

// MonorepoOverview prints detected sub-projects in a monorepo.
func MonorepoOverview(root string, subs []detect.SubProject) {
	name := filepath.Base(root)

	fmt.Println()
	fmt.Printf("  %s %s %s\n", Orange("🦖"), Bold(name), Dim("(monorepo)"))
	fmt.Println()
	fmt.Printf("  %s  %d sub-projects detected\n", Dim("workspace"), len(subs))
	fmt.Println()

	for _, sub := range subs {
		stack := sub.Detection.Stack
		if sub.Detection.PkgMgr != "" && sub.Detection.PkgMgr != stack {
			stack += " + " + sub.Detection.PkgMgr
		}
		fmt.Printf("    %s %s %s\n",
			Cyan(fmt.Sprintf("%-20s", sub.Path)),
			Dim("→"),
			stack,
		)
	}
	fmt.Println()
	fmt.Printf("  %s cd into a sub-project, then use %s\n", Dim("hint:"), Cyan("rex test"))
	fmt.Println()
}
