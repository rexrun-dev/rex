package envfile

import (
	"os"
	"path/filepath"
	"strings"
)

// Load reads a .env file from the given directory and sets environment variables.
// Variables already set in the environment are NOT overwritten.
// Returns the number of variables loaded, or 0 if no .env file exists.
func Load(root string) int {
	path := filepath.Join(root, ".env")
	b, err := os.ReadFile(path)
	if err != nil {
		return 0
	}

	count := 0
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || line[0] == '#' {
			continue
		}

		// Find key=value
		idx := strings.IndexByte(line, '=')
		if idx <= 0 {
			continue
		}

		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])

		// Remove surrounding quotes
		value = unquote(value)

		// Don't overwrite existing env vars
		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		os.Setenv(key, value)
		count++
	}

	return count
}

func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
