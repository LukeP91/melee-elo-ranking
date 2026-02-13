package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	configContent := `{
		"elo": {
			"k_factor": 32,
			"initial_rating": 1500
		},
		"paths": {
			"pending_dir": "data/pending",
			"processed_dir": "data/processed",
			"failed_dir": "data/failed",
			"database": "data/rankings.db",
			"output": "docs/index.html"
		},
		"output": {
			"type": "file",
			"title": "Test Rankings",
			"description": "Test description"
		}
	}`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.ELO.KFactor != 32 {
		t.Errorf("expected k_factor 32, got %d", cfg.ELO.KFactor)
	}
	if cfg.ELO.InitialRating != 1500 {
		t.Errorf("expected initial_rating 1500, got %d", cfg.ELO.InitialRating)
	}
	if cfg.Output.Title != "Test Rankings" {
		t.Errorf("expected title 'Test Rankings', got %s", cfg.Output.Title)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	// Create temp config with minimal content
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	configContent := `{
		"paths": {
			"pending_dir": "data/pending",
			"database": "data/rankings.db",
			"output": "docs/index.html"
		}
	}`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Check default values
	if cfg.ELO.KFactor != 32 {
		t.Errorf("expected default k_factor 32, got %d", cfg.ELO.KFactor)
	}
	if cfg.ELO.InitialRating != 1500 {
		t.Errorf("expected default initial_rating 1500, got %d", cfg.ELO.InitialRating)
	}
}

func TestLoadConfigMissing(t *testing.T) {
	_, err := Load("/nonexistent/path/config.json")
	if err == nil {
		t.Error("expected error for missing config file")
	}
}
