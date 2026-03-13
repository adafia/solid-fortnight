package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadWithEnvExpansion(t *testing.T) {
	os.Setenv("TEST_DB_USER", "postgres_user")
	defer os.Unsetenv("TEST_DB_USER")

	configContent := `
storage:
  type: "postgres"
  postgres:
    user: "${TEST_DB_USER}"
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Storage.Postgres.User != "postgres_user" {
		t.Errorf("expected user 'postgres_user', got '%s'", cfg.Storage.Postgres.User)
	}
}
