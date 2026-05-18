package detect

import (
	"os"
	"path/filepath"
	"strings"
)

// LoadConfig reads an optional rex.toml and returns command overrides.
// Format:
//
//	[commands]
//	test = "go test -race ./..."
//	run = "air"
func LoadConfig(root string) map[Verb]string {
	path := filepath.Join(root, "rex.toml")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return parseRexToml(string(b))
}

func parseRexToml(content string) map[Verb]string {
	cmds := make(map[Verb]string)
	inCommands := false

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if line == "[commands]" {
			inCommands = true
			continue
		}
		if strings.HasPrefix(line, "[") {
			inCommands = false
			continue
		}
		if !inCommands {
			continue
		}
		idx := strings.Index(line, "=")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		val = strings.Trim(val, `"'`)
		if key != "" && val != "" {
			cmds[Verb(key)] = val
		}
	}
	return cmds
}
