package envfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	envContent := `# Database config
DB_HOST=localhost
DB_PORT=5432
DB_NAME="myapp_dev"
SECRET_KEY='super-secret'

# Empty lines are skipped
API_URL=https://api.example.com
`
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Clear any existing values
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("SECRET_KEY")
	os.Unsetenv("API_URL")

	count := Load(dir)
	if count != 5 {
		t.Errorf("expected 5 vars loaded, got %d", count)
	}

	tests := []struct {
		key, want string
	}{
		{"DB_HOST", "localhost"},
		{"DB_PORT", "5432"},
		{"DB_NAME", "myapp_dev"},
		{"SECRET_KEY", "super-secret"},
		{"API_URL", "https://api.example.com"},
	}
	for _, tt := range tests {
		got := os.Getenv(tt.key)
		if got != tt.want {
			t.Errorf("%s = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestLoadNoOverwrite(t *testing.T) {
	dir := t.TempDir()
	envContent := `MY_VAR=from-file`
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644); err != nil {
		t.Fatal(err)
	}

	os.Setenv("MY_VAR", "from-env")
	defer os.Unsetenv("MY_VAR")

	Load(dir)
	got := os.Getenv("MY_VAR")
	if got != "from-env" {
		t.Errorf("MY_VAR = %q, want %q (should not overwrite)", got, "from-env")
	}
}

func TestLoadNoFile(t *testing.T) {
	dir := t.TempDir()
	count := Load(dir)
	if count != 0 {
		t.Errorf("expected 0 vars loaded from empty dir, got %d", count)
	}
}
