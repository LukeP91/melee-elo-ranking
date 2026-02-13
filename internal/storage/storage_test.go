package storage

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func createTestDB(t *testing.T) *Storage {
	tmpFile := filepath.Join(t.TempDir(), "test.db")
	db, err := sql.Open("sqlite3", tmpFile)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	storage := &Storage{db: db}
	if err := storage.createTables(); err != nil {
		t.Fatalf("failed to create tables: %v", err)
	}

	return storage
}

func TestPlayerCRUD(t *testing.T) {
	store := createTestDB(t)
	defer store.Close()

	// Create player
	player, err := store.GetOrCreatePlayer(1, "TestPlayer", "testuser")
	if err != nil {
		t.Fatalf("failed to create player: %v", err)
	}

	if player.DisplayName != "TestPlayer" {
		t.Errorf("expected display name TestPlayer, got %s", player.DisplayName)
	}
	if player.CurrentELO != 1500 {
		t.Errorf("expected initial ELO 1500, got %d", player.CurrentELO)
	}

	// Get existing player
	existing, err := store.GetOrCreatePlayer(1, "TestPlayer", "testuser")
	if err != nil {
		t.Fatalf("failed to get player: %v", err)
	}

	if existing.ID != player.ID {
		t.Errorf("expected same ID %d, got %d", player.ID, existing.ID)
	}
}

func TestTournamentCRUD(t *testing.T) {
	store := createTestDB(t)
	defer store.Close()

	// Create tournament
	tournamentDate := time.Date(2024, 10, 17, 0, 0, 0, 0, time.UTC)
	tournament, err := store.GetOrCreateTournament(170676, tournamentDate)
	if err != nil {
		t.Fatalf("failed to create tournament: %v", err)
	}

	if tournament.MeleeID != 170676 {
		t.Errorf("expected melee ID 170676, got %d", tournament.MeleeID)
	}

	// Get existing tournament
	existing, err := store.GetTournamentByMeleeID(170676)
	if err != nil {
		t.Fatalf("failed to get tournament: %v", err)
	}

	if existing.ID != tournament.ID {
		t.Errorf("expected same ID %d, got %d", tournament.ID, existing.ID)
	}
}

func TestMatchCRUD(t *testing.T) {
	store := createTestDB(t)
	defer store.Close()

	// Create players
	player1, _ := store.GetOrCreatePlayer(1, "Player1", "p1")
	player2, _ := store.GetOrCreatePlayer(2, "Player2", "p2")

	// Create tournament
	tournamentDate := time.Date(2024, 10, 17, 0, 0, 0, 0, time.UTC)
	store.GetOrCreateTournament(1, tournamentDate)

	// Save match
	match := Match{
		ID:           "test-match-1",
		TournamentID: 1,
		Round:        1,
		Player1ID:    player1.ID,
		Player2ID:    player2.ID,
		Player1Wins:  2,
		Player2Wins:  1,
		DatePlayed:   tournamentDate,
	}

	err := store.SaveMatch(match)
	if err != nil {
		t.Fatalf("failed to save match: %v", err)
	}

	// Check exists
	exists, err := store.MatchExists("test-match-1")
	if err != nil {
		t.Fatalf("failed to check match: %v", err)
	}
	if !exists {
		t.Error("expected match to exist")
	}
}

func TestRankingsOrder(t *testing.T) {
	store := createTestDB(t)
	defer store.Close()

	// Create players with different ELOs
	store.GetOrCreatePlayer(1, "Alice", "alice")
	store.GetOrCreatePlayer(2, "Bob", "bob")
	store.GetOrCreatePlayer(3, "Charlie", "charlie")

	// Update ELOs directly
	store.UpdatePlayerELO(1, 1600, true)
	store.UpdatePlayerELO(1, 1600, true) // +2 wins
	store.UpdatePlayerELO(2, 1500, true)
	store.UpdatePlayerELO(2, 1500, true) // +2 wins
	store.UpdatePlayerELO(3, 1400, true)
	store.UpdatePlayerELO(3, 1400, true) // +2 wins

	// Add more wins to meet threshold
	store.UpdatePlayerELO(1, 1610, true)
	store.UpdatePlayerELO(1, 1610, true)
	store.UpdatePlayerELO(1, 1610, true)
	store.UpdatePlayerELO(1, 1610, true)
	store.UpdatePlayerELO(1, 1610, true)
	store.UpdatePlayerELO(1, 1610, true)
	store.UpdatePlayerELO(1, 1610, true)
	store.UpdatePlayerELO(1, 1610, true)
	store.UpdatePlayerELO(2, 1510, true)
	store.UpdatePlayerELO(2, 1510, true)
	store.UpdatePlayerELO(2, 1510, true)
	store.UpdatePlayerELO(2, 1510, true)
	store.UpdatePlayerELO(2, 1510, true)
	store.UpdatePlayerELO(2, 1510, true)
	store.UpdatePlayerELO(2, 1510, true)
	store.UpdatePlayerELO(2, 1510, true)
	store.UpdatePlayerELO(3, 1410, true)
	store.UpdatePlayerELO(3, 1410, true)
	store.UpdatePlayerELO(3, 1410, true)
	store.UpdatePlayerELO(3, 1410, true)
	store.UpdatePlayerELO(3, 1410, true)
	store.UpdatePlayerELO(3, 1410, true)
	store.UpdatePlayerELO(3, 1410, true)
	store.UpdatePlayerELO(3, 1410, true)

	rankings, err := store.GetRankings()
	if err != nil {
		t.Fatalf("failed to get rankings: %v", err)
	}

	// Should be ordered by ELO desc
	if len(rankings) != 3 {
		t.Errorf("expected 3 rankings, got %d", len(rankings))
	}

	if len(rankings) > 0 && rankings[0].DisplayName != "Alice" {
		t.Errorf("expected Alice first, got %s", rankings[0].DisplayName)
	}
	if len(rankings) > 1 && rankings[1].DisplayName != "Bob" {
		t.Errorf("expected Bob second, got %s", rankings[1].DisplayName)
	}
	if len(rankings) > 2 && rankings[2].DisplayName != "Charlie" {
		t.Errorf("expected Charlie third, got %s", rankings[2].DisplayName)
	}
}

func TestPlayerMatchHistory(t *testing.T) {
	store := createTestDB(t)
	defer store.Close()

	// Setup: create players and tournament
	player1, _ := store.GetOrCreatePlayer(1, "Alice", "alice")
	player2, _ := store.GetOrCreatePlayer(2, "Bob", "bob")
	player3, _ := store.GetOrCreatePlayer(3, "Charlie", "charlie")

	tournamentDate := time.Date(2024, 10, 17, 0, 0, 0, 0, time.UTC)
	store.GetOrCreateTournament(1, tournamentDate)

	// Alice vs Bob (Alice wins)
	store.SaveMatch(Match{
		ID: "match-1", TournamentID: 1, Round: 1,
		Player1ID: player1.ID, Player2ID: player2.ID,
		Player1Wins: 2, Player2Wins: 1,
		Player1ELOBefore: 1500, Player2ELOBefore: 1500,
		Player1ELOAfter: 1520, Player2ELOAfter: 1480,
	})

	// Alice vs Charlie (Alice wins)
	store.SaveMatch(Match{
		ID: "match-2", TournamentID: 1, Round: 2,
		Player1ID: player1.ID, Player2ID: player3.ID,
		Player1Wins: 2, Player2Wins: 0,
		Player1ELOBefore: 1520, Player2ELOBefore: 1500,
		Player1ELOAfter: 1540, Player2ELOAfter: 1480,
	})

	matches, err := store.GetPlayerMatchHistory("Alice")
	if err != nil {
		t.Fatalf("failed to get match history: %v", err)
	}

	if len(matches) != 2 {
		t.Errorf("expected 2 matches, got %d", len(matches))
	}
}

func TestMatchups(t *testing.T) {
	store := createTestDB(t)
	defer store.Close()

	// Setup: create players and tournament
	player1, _ := store.GetOrCreatePlayer(1, "Alice", "alice")
	player2, _ := store.GetOrCreatePlayer(2, "Bob", "bob")

	tournamentDate := time.Date(2024, 10, 17, 0, 0, 0, 0, time.UTC)
	store.GetOrCreateTournament(1, tournamentDate)

	// Alice vs Bob: 2-1 (Alice wins)
	store.SaveMatch(Match{
		ID: "match-1", TournamentID: 1, Round: 1,
		Player1ID: player1.ID, Player2ID: player2.ID,
		Player1Wins: 2, Player2Wins: 1,
	})

	// Alice vs Bob: 0-2 (Bob wins) - need second match
	store.SaveMatch(Match{
		ID: "match-2", TournamentID: 1, Round: 2,
		Player1ID: player1.ID, Player2ID: player2.ID,
		Player1Wins: 0, Player2Wins: 2,
	})

	matchups, err := store.GetMatchups()
	if err != nil {
		t.Fatalf("failed to get matchups: %v", err)
	}

	// Should have one matchup (Alice vs Bob)
	if len(matchups) == 0 {
		t.Error("expected at least one matchup")
	}
}

func TestFullRebuild(t *testing.T) {
	store := createTestDB(t)
	defer store.Close()

	// Setup: create players and tournament
	player1, _ := store.GetOrCreatePlayer(1, "Alice", "alice")
	player2, _ := store.GetOrCreatePlayer(2, "Bob", "bob")

	tournamentDate := time.Date(2024, 10, 17, 0, 0, 0, 0, time.UTC)
	store.GetOrCreateTournament(1, tournamentDate)

	// Save matches (without ELO updates, simulating imported data)
	store.SaveMatch(Match{
		ID: "match-1", TournamentID: 1, Round: 1,
		Player1ID: player1.ID, Player2ID: player2.ID,
		Player1Wins: 2, Player2Wins: 1,
	})

	// Reset and rebuild
	err := store.ResetAllPlayersELO()
	if err != nil {
		t.Fatalf("failed to reset ELO: %v", err)
	}

	// Verify players were reset
	playerAfterReset, _ := store.GetPlayerByID(player1.ID)
	if playerAfterReset.CurrentELO != 1500 {
		t.Errorf("expected ELO 1500 after reset, got %d", playerAfterReset.CurrentELO)
	}
	if playerAfterReset.MatchesPlayed != 0 {
		t.Errorf("expected 0 matches after reset, got %d", playerAfterReset.MatchesPlayed)
	}
}
