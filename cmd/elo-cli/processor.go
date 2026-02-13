package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/melee-elo-ranking/internal/config"
	"github.com/melee-elo-ranking/internal/elo"
	"github.com/melee-elo-ranking/internal/melee"
	"github.com/melee-elo-ranking/internal/parser"
	"github.com/melee-elo-ranking/internal/storage"
)

type Processor struct {
	store           *storage.Storage
	calculator      *elo.Calculator
	parser          *parser.Parser
	meleeClient     *melee.Client
	tournamentDates map[int]string
	config          *config.Config
}

func NewProcessor(store *storage.Storage, calc *elo.Calculator, parser *parser.Parser, meleeClient *melee.Client, tournamentDates map[int]string, cfg *config.Config) *Processor {
	return &Processor{
		store:           store,
		calculator:      calc,
		parser:          parser,
		meleeClient:     meleeClient,
		tournamentDates: tournamentDates,
		config:          cfg,
	}
}

func (p *Processor) Process() error {
	files, err := os.ReadDir(p.config.Paths.PendingDir)
	if err != nil {
		return fmt.Errorf("failed to read pending directory: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("No pending files to process")
		return p.fullRebuild()
	}

	type tournamentFile struct {
		tournamentID int
		matches      []parser.Match
		filename     string
	}

	var tournamentFiles []tournamentFile

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		tournamentID, err := extractTournamentID(file.Name())
		if err != nil {
			fmt.Printf("Warning: failed to extract tournament ID from %s: %v\n", file.Name(), err)
			p.moveToFailed(file.Name())
			continue
		}

		filepath := filepath.Join(p.config.Paths.PendingDir, file.Name())
		matches, err := p.parser.ParseFile(filepath, tournamentID)
		if err != nil {
			fmt.Printf("Warning: failed to parse %s: %v\n", file.Name(), err)
			p.moveToFailed(file.Name())
			continue
		}

		if len(matches) > 0 {
			tournamentFiles = append(tournamentFiles, tournamentFile{
				tournamentID: tournamentID,
				matches:      matches,
				filename:     file.Name(),
			})
		}
	}

	if len(tournamentFiles) == 0 {
		fmt.Println("No valid matches found in pending files")
		return p.fullRebuild()
	}

	newTournaments := 0
	for _, tf := range tournamentFiles {
		var tournamentDate time.Time
		var err error

		// Check if date was provided via flag
		if dateStr, ok := p.tournamentDates[tf.tournamentID]; ok {
			tournamentDate, err = time.Parse("2006-01-02", dateStr)
			if err != nil {
				fmt.Printf("Warning: invalid date format for tournament %d: %v\n", tf.tournamentID, err)
			}
		} else {
			// Try to get from existing tournament
			existing, _ := p.store.GetTournamentByMeleeID(tf.tournamentID)
			if existing != nil && !existing.Date.IsZero() {
				tournamentDate = existing.Date
			} else {
				// Try to fetch date from melee.gg page
				if p.meleeClient != nil {
					fetched, fetchErr := p.meleeClient.FetchTournamentDate(tf.tournamentID)
					if fetchErr == nil && !fetched.IsZero() {
						tournamentDate = fetched
						fmt.Printf("Fetched date for tournament %d from melee.gg: %s\n", tf.tournamentID, tournamentDate.Format("2006-01-02"))
					} else if fetchErr != nil {
						fmt.Printf("Could not fetch date for tournament %d: %v\n", tf.tournamentID, fetchErr)
					}
				}
				// If still no date, prompt user
				if tournamentDate.IsZero() {
					tournamentDate, err = promptForTournamentDate(tf.tournamentID)
					if err != nil {
						fmt.Printf("Warning: failed to get tournament date for %d: %v\n", tf.tournamentID, err)
						p.moveToFailed(tf.filename)
						continue
					}
				}
			}
		}

		_, err = p.store.GetOrCreateTournament(tf.tournamentID, tournamentDate)
		if err != nil {
			fmt.Printf("Warning: failed to create tournament %d: %v\n", tf.tournamentID, err)
			continue
		}
		newTournaments++

		for _, match := range tf.matches {
			exists, err := p.store.MatchExists(match.ID)
			if err != nil {
				fmt.Printf("Warning: failed to check match %s: %v\n", match.ID, err)
				continue
			}
			if exists {
				continue
			}

			c1 := match.Competitors[0]
			c2 := match.Competitors[1]

			player1, err := p.store.GetOrCreatePlayer(c1.Player.ID, c1.Player.DisplayName, c1.Player.Username)
			if err != nil {
				fmt.Printf("Warning: failed to get/create player %s: %v\n", c1.Player.DisplayName, err)
				continue
			}
			player2, err := p.store.GetOrCreatePlayer(c2.Player.ID, c2.Player.DisplayName, c2.Player.Username)
			if err != nil {
				fmt.Printf("Warning: failed to get/create player %s: %v\n", c2.Player.DisplayName, err)
				continue
			}

			storageMatch := storage.Match{
				ID:           match.ID,
				TournamentID: tf.tournamentID,
				Round:        match.RoundNumber,
				Player1ID:    player1.ID,
				Player2ID:    player2.ID,
				Player1Wins:  c1.GameWins,
				Player2Wins:  c2.GameWins,
				DatePlayed:   match.DateCreated,
			}

			if err := p.store.SaveMatch(storageMatch); err != nil {
				fmt.Printf("Warning: failed to save match %s: %v\n", match.ID, err)
			}
		}

		if err := p.moveToProcessed(tf.filename); err != nil {
			fmt.Printf("Warning: failed to move file %s: %v\n", tf.filename, err)
		}
	}

	fmt.Printf("Processed %d new tournaments\n", newTournaments)

	return p.fullRebuild()
}

func promptForTournamentDate(tournamentID int) (time.Time, error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("========================================")
		fmt.Printf("Tournament ID: %d\n", tournamentID)
		fmt.Printf("URL: https://melee.gg/Tournament/View/%d\n", tournamentID)
		fmt.Print("Please enter date (YYYY-MM-DD) or press Enter to skip: ")

		input, err := reader.ReadString('\n')
		if err != nil {
			return time.Time{}, err
		}

		input = strings.TrimSpace(input)
		if input == "" {
			return time.Time{}, nil
		}

		parsed, err := time.Parse("2006-01-02", input)
		if err != nil {
			fmt.Printf("Invalid date format. Please use YYYY-MM-DD (e.g., 2024-08-31)\n\n")
			continue
		}

		return parsed, nil
	}
}

func extractTournamentID(filename string) (int, error) {
	re := regexp.MustCompile(`Matches-tournament-(\d+)\.json`)
	matches := re.FindStringSubmatch(filename)
	if len(matches) < 2 {
		return 0, fmt.Errorf("filename does not match expected format")
	}
	return strconv.Atoi(matches[1])
}

func (p *Processor) fullRebuild() error {
	fmt.Println("Performing full ELO rebuild...")

	if err := p.store.ResetAllPlayersELO(); err != nil {
		return fmt.Errorf("failed to reset player ELOs: %w", err)
	}

	allMatches, err := p.store.GetAllMatchesSorted()
	if err != nil {
		return fmt.Errorf("failed to get matches: %w", err)
	}

	fmt.Printf("Processing %d matches in chronological order...\n", len(allMatches))

	for _, match := range allMatches {
		if err := p.processMatchFromDB(match); err != nil {
			fmt.Printf("Warning: failed to process match %s: %v\n", match.ID, err)
		}
	}

	fmt.Println("Full rebuild complete")
	return nil
}

func (p *Processor) processMatchFromDB(match storage.Match) error {
	player1, err := p.store.GetPlayerByID(match.Player1ID)
	if err != nil {
		return fmt.Errorf("failed to get player 1: %w", err)
	}
	player2, err := p.store.GetPlayerByID(match.Player2ID)
	if err != nil {
		return fmt.Errorf("failed to get player 2: %w", err)
	}

	var winnerID *int64
	if match.Player1Wins > match.Player2Wins {
		winnerID = &match.Player1ID
	} else if match.Player2Wins > match.Player1Wins {
		winnerID = &match.Player2ID
	}

	newELO1, newELO2 := p.calculator.Calculate(
		player1.CurrentELO,
		player2.CurrentELO,
		winnerID,
		&match.Player1ID,
		&match.Player2ID,
		player1.MatchesPlayed,
		player2.MatchesPlayed,
	)

	if err := p.store.UpdatePlayerELO(match.Player1ID, newELO1, match.Player1Wins > match.Player2Wins); err != nil {
		return fmt.Errorf("failed to update player 1 ELO: %w", err)
	}
	if err := p.store.UpdatePlayerELO(match.Player2ID, newELO2, match.Player2Wins > match.Player1Wins); err != nil {
		return fmt.Errorf("failed to update player 2 ELO: %w", err)
	}

	// Update match with ELO values
	if err := p.store.UpdateMatchELO(match.ID, player1.CurrentELO, player2.CurrentELO, newELO1, newELO2); err != nil {
		return fmt.Errorf("failed to update match ELO: %w", err)
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

	return sourceFile.Close()
}
