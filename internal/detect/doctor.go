package detect

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Check represents a single doctor diagnostic.
type Check struct {
	OK      bool
	Label   string
	Detail  string
	Fix     string
}

// Doctor runs diagnostics on the project environment.
func Doctor(root string) []Check {
	d := Detect(root)
	var checks []Check

	if d.Stack == "" {
		checks = append(checks, Check{
			OK: false, Label: "project detection",
			Detail: "no supported project found",
			Fix:    "ensure you are in a project root with go.mod, package.json, Cargo.toml, etc.",
		})
		return checks
	}

	checks = append(checks, Check{OK: true, Label: "project detected", Detail: d.Stack})

	switch d.Stack {
	case "go":
		checks = append(checks, checkBinary("go", "go version"))
	case "node":
		checks = append(checks, checkBinary("node", "node --version"))
		checks = append(checks, checkBinary(d.PkgMgr, d.PkgMgr+" --version"))
	case "python":
		checks = append(checks, checkBinary("python", "python --version"))
		if d.PkgMgr == "uv" {
			checks = append(checks, checkBinary("uv", "uv --version"))
		}
	case "rust":
		checks = append(checks, checkBinary("cargo", "cargo --version"))
	}

	// Check for .env
	envExample := filepath.Join(root, ".env.example")
	envFile := filepath.Join(root, ".env")
	if fileExists(envExample) {
		if !fileExists(envFile) {
			checks = append(checks, Check{
				OK: false, Label: ".env file",
				Detail: ".env.example exists but .env is missing",
				Fix:    "cp .env.example .env",
			})
		} else {
			missing := findMissingEnvVars(envExample, envFile)
			if len(missing) > 0 {
				checks = append(checks, Check{
					OK: false, Label: ".env vars",
					Detail: fmt.Sprintf("missing: %s", strings.Join(missing, ", ")),
					Fix:    "add missing variables to .env",
				})
			} else {
				checks = append(checks, Check{OK: true, Label: ".env", Detail: "all vars present"})
			}
		}
	}

	// Check deps installed
	switch d.Stack {
	case "node":
		if !dirExists(filepath.Join(root, "node_modules")) {
			checks = append(checks, Check{
				OK: false, Label: "dependencies",
				Detail: "node_modules not found",
				Fix:    "rex deps",
			})
		} else {
			checks = append(checks, Check{OK: true, Label: "dependencies", Detail: "node_modules present"})
		}
	case "go":
		// go modules are always downloaded on demand, less critical
		checks = append(checks, Check{OK: true, Label: "dependencies", Detail: "go modules (downloaded on demand)"})
	}

	return checks
}

func checkBinary(name, versionCmd string) Check {
	_, err := exec.LookPath(name)
	if err != nil {
		return Check{
			OK: false, Label: name,
			Detail: "not found in PATH",
			Fix:    fmt.Sprintf("install %s", name),
		}
	}
	parts := strings.Fields(versionCmd)
	out, err := exec.Command(parts[0], parts[1:]...).Output()
	if err != nil {
		return Check{OK: true, Label: name, Detail: "found (version unknown)"}
	}
	ver := strings.TrimSpace(strings.Split(string(out), "\n")[0])
	return Check{OK: true, Label: name, Detail: ver}
}

func findMissingEnvVars(examplePath, envPath string) []string {
	exampleVars := parseEnvKeys(examplePath)
	envVars := parseEnvKeys(envPath)

	var missing []string
	for _, v := range exampleVars {
		found := false
		for _, ev := range envVars {
			if ev == v {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, v)
		}
	}
	return missing
}

func parseEnvKeys(path string) []string {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var keys []string
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if idx := strings.Index(line, "="); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			if key != "" {
				keys = append(keys, key)
			}
		}
	}
	return keys
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
