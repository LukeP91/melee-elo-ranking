package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	ELO    ELOConfig    `json:"elo"`
	Paths  PathsConfig  `json:"paths"`
	Output OutputConfig `json:"output"`
}

type ELOConfig struct {
	KFactor       int `json:"k_factor"`
	InitialRating int `json:"initial_rating"`
}

type PathsConfig struct {
	PendingDir   string `json:"pending_dir"`
	ProcessedDir string `json:"processed_dir"`
	FailedDir    string `json:"failed_dir"`
	Database     string `json:"database"`
	Output       string `json:"output"`
}

type OutputConfig struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Set defaults if not specified
	if cfg.ELO.KFactor == 0 {
		cfg.ELO.KFactor = 32
	}
	if cfg.ELO.InitialRating == 0 {
		cfg.ELO.InitialRating = 1500
	}

	return &cfg, nil
}
