package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/melee-elo-ranking/internal/config"
	"github.com/melee-elo-ranking/internal/elo"
	"github.com/melee-elo-ranking/internal/parser"
	"github.com/melee-elo-ranking/internal/storage"
)

type Processor struct {
	store      *storage.Storage
	calculator *elo.Calculator
	parser     *parser.Parser
	config     *config.Config
}

func NewProcessor(store *storage.Storage, calc *elo.Calculator, parser *parser.Parser, cfg *config.Config) *Processor {
	return &Processor{
		store:      store,
		calculator: calc,
		parser:     parser,
		config:     cfg,
	}
}

func (p *Processor) Process() error {
	// Get list of pending files
	files, err := os.ReadDir(p.config.Paths.PendingDir)
	if err != nil {
		return fmt.Errorf("failed to read pending directory: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("No pending files to process")
		return nil
	}

	// Collect all matches from all files
	var allMatches []parser.Match
	fileMatches := make(map[string][]parser.Match)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filepath := filepath.Join(p.config.Paths.PendingDir, file.Name())
		matches, err := p.parser.ParseFile(filepath)
		if err != nil {
			fmt.Printf("Warning: failed to parse %s: %v\n", file.Name(), err)
			p.moveToFailed(file.Name())
			continue
		}

		fileMatches[file.Name()] = matches
		allMatches = append(allMatches, matches...)
	}

	if len(allMatches) == 0 {
		fmt.Println("No valid matches found in pending files")
		return nil
	}

	// Sort matches by tournament ID, then by round number (chronological order)
	sort.Slice(allMatches, func(i, j int) bool {
		if allMatches[i].TournamentID != allMatches[j].TournamentID {
			return allMatches[i].TournamentID < allMatches[j].TournamentID
		}
		return allMatches[i].RoundNumber < allMatches[j].RoundNumber
	})

	fmt.Printf("Processing %d matches from %d files...\n", len(allMatches), len(fileMatches))

	// Process each match
	processedCount := 0
	skippedCount := 0
	for _, match := range allMatches {
		// Check if already processed
		exists, err := p.store.MatchExists(match.ID)
		if err != nil {
			return fmt.Errorf("failed to check match existence: %w", err)
		}
		if exists {
			skippedCount++
			continue
		}

		// Process the match
		if err := p.processMatch(match); err != nil {
			fmt.Printf("Warning: failed to process match %s: %v\n", match.ID, err)
			continue
		}
		processedCount++
	}

	// Move processed files
	for filename := range fileMatches {
		if err := p.moveToProcessed(filename); err != nil {
			fmt.Printf("Warning: failed to move file %s: %v\n", filename, err)
		}
	}

	fmt.Printf("Processed %d new matches, skipped %d duplicates\n", processedCount, skippedCount)
	return nil
}

func (p *Processor) processMatch(match parser.Match) error {
	// Skip byes (only one competitor)
	if len(match.Competitors) != 2 {
		return nil
	}

	c1 := match.Competitors[0]
	c2 := match.Competitors[1]

	// Get or create players
	player1, err := p.store.GetOrCreatePlayer(c1.Player.ID, c1.Player.DisplayName, c1.Player.Username)
	if err != nil {
		return fmt.Errorf("failed to get/create player 1: %w", err)
	}

	player2, err := p.store.GetOrCreatePlayer(c2.Player.ID, c2.Player.DisplayName, c2.Player.Username)
	if err != nil {
		return fmt.Errorf("failed to get/create player 2: %w", err)
	}

	// Determine winner (higher game wins wins)
	var winnerID *int64
	if c1.GameWins > c2.GameWins {
		winnerID = &player1.ID
	} else if c2.GameWins > c1.GameWins {
		winnerID = &player2.ID
	}

	// Calculate new ELOs
	newELO1, newELO2 := p.calculator.Calculate(
		player1.CurrentELO,
		player2.CurrentELO,
		winnerID,
		&player1.ID,
		&player2.ID,
		player1.MatchesPlayed,
		player2.MatchesPlayed,
	)

	// Update players
	if err := p.store.UpdatePlayerELO(player1.ID, newELO1, winnerID != nil && *winnerID == player1.ID); err != nil {
		return fmt.Errorf("failed to update player 1: %w", err)
	}
	if err := p.store.UpdatePlayerELO(player2.ID, newELO2, winnerID != nil && *winnerID == player2.ID); err != nil {
		return fmt.Errorf("failed to update player 2: %w", err)
	}

	// Store match record
	matchRecord := storage.Match{
		ID:               match.ID,
		TournamentID:     match.TournamentID,
		Round:            match.RoundNumber,
		Player1ID:        player1.ID,
		Player2ID:        player2.ID,
		Player1Wins:      c1.GameWins,
		Player2Wins:      c2.GameWins,
		DatePlayed:       match.DateCreated,
		Player1ELOBefore: player1.CurrentELO,
		Player2ELOBefore: player2.CurrentELO,
		Player1ELOAfter:  newELO1,
		Player2ELOAfter:  newELO2,
	}

	if err := p.store.SaveMatch(matchRecord); err != nil {
		return fmt.Errorf("failed to save match: %w", err)
	}

	return nil
}

func (p *Processor) moveToProcessed(filename string) error {
	src := filepath.Join(p.config.Paths.PendingDir, filename)
	dst := filepath.Join(p.config.Paths.ProcessedDir, filename)
	return moveFile(src, dst)
}

func (p *Processor) moveToFailed(filename string) error {
	src := filepath.Join(p.config.Paths.PendingDir, filename)
	dst := filepath.Join(p.config.Paths.FailedDir, filename)
	return moveFile(src, dst)
}

func moveFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return os.Remove(src)
}
