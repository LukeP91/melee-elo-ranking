package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/melee-elo-ranking/internal/config"
	"github.com/melee-elo-ranking/internal/elo"
	"github.com/melee-elo-ranking/internal/generator"
	"github.com/melee-elo-ranking/internal/melee"
	"github.com/melee-elo-ranking/internal/parser"
	"github.com/melee-elo-ranking/internal/storage"
)

var tournamentDates = flag.String("dates", "", "Tournament dates in format: 170676=2024-08-31,172453=2024-10-17")

func main() {
	flag.Parse()

	// Load configuration
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Ensure directories exist
	ensureDirs(cfg)

	// Initialize storage
	store, err := storage.New(cfg.Paths.Database)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Create ELO calculator
	calculator := elo.New(cfg.ELO.InitialRating)

	// Create parser
	matchParser := parser.New()

	// Create melee client for fetching tournament dates from melee.gg
	meleeClient := melee.NewClient()

	// Parse tournament dates from flag
	datesMap := parseTournamentDates(*tournamentDates)

	// Process pending matches
	processor := NewProcessor(store, calculator, matchParser, meleeClient, datesMap, cfg)
	if err := processor.Process(); err != nil {
		log.Fatalf("Failed to process matches: %v", err)
	}

	// Generate rankings
	rankings, err := store.GetRankings()
	if err != nil {
		log.Fatalf("Failed to get rankings: %v", err)
	}

	// Generate HTML
	gen := generator.New(cfg.Output.Title, cfg.Output.Description)
	if err := gen.Generate(rankings, cfg.Paths.Output); err != nil {
		log.Fatalf("Failed to generate HTML: %v", err)
	}

	// Generate player detail pages
	playersDir := "docs/players"
	if err := os.MkdirAll(playersDir, 0755); err != nil {
		log.Fatalf("Failed to create players directory: %v", err)
	}

	for _, r := range rankings {
		matches, err := store.GetPlayerMatchHistory(r.DisplayName)
		if err != nil {
			log.Printf("Warning: Failed to get match history for %s: %v", r.DisplayName, err)
			continue
		}

		playerPath := playersDir + "/" + r.DisplayName + ".html"
		if err := gen.GeneratePlayerPage(r.DisplayName, matches, r, playerPath); err != nil {
			log.Printf("Warning: Failed to generate player page for %s: %v", r.DisplayName, err)
			continue
		}
	}

	log.Println("Successfully generated rankings at", cfg.Paths.Output)
	log.Println("Generated player pages in", playersDir)

	// Generate matchup matrix
	matchups, err := store.GetMatchups()
	if err != nil {
		log.Printf("Warning: Failed to get matchups: %v", err)
	} else {
		playerNames := make([]string, len(rankings))
		for i, r := range rankings {
			playerNames[i] = r.DisplayName
		}

		matchupPath := "docs/matchups.html"
		if err := gen.GenerateMatchupMatrix(matchups, playerNames, matchupPath); err != nil {
			log.Printf("Warning: Failed to generate matchup matrix: %v", err)
		} else {
			log.Println("Generated matchup matrix at", matchupPath)
		}
	}
}

func ensureDirs(cfg *config.Config) {
	dirs := []string{
		cfg.Paths.PendingDir,
		cfg.Paths.ProcessedDir,
		cfg.Paths.FailedDir,
		"docs",
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}
}

func parseTournamentDates(datesStr string) map[int]string {
	result := make(map[int]string)
	if datesStr == "" {
		return result
	}
	// Format: 170676=2024-08-31,172453=2024-10-17
	// or: 170676=2024-08-31
	for _, part := range splitAndTrim(datesStr, ",") {
		kv := splitAndTrim(part, "=")
		if len(kv) == 2 {
			var id int
			if _, err := fmt.Sscanf(kv[0], "%d", &id); err == nil {
				result[id] = kv[1]
			}
		}
	}
	return result
}

func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, len(parts))
	for i, p := range parts {
		result[i] = strings.TrimSpace(p)
	}
	return result
}
