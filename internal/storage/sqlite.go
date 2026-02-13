package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

type Player struct {
	ID            int64
	ExternalID    int64
	DisplayName   string
	Username      string
	CurrentELO    int
	MatchesPlayed int
	Wins          int
	Losses        int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Match struct {
	ID                string
	TournamentID      int
	TournamentMeleeID int
	Round             int
	Player1ID         int64
	Player2ID         int64
	Player1Wins       int
	Player2Wins       int
	DatePlayed        time.Time
	Player1ELOBefore  int
	Player2ELOBefore  int
	Player1ELOAfter   int
	Player2ELOAfter   int
}

type Tournament struct {
	ID      int64
	MeleeID int
	Date    time.Time
}

type Ranking struct {
	Rank          int
	DisplayName   string
	Username      string
	CurrentELO    int
	MatchesPlayed int
	Wins          int
	Losses        int
	WinRate       float64
}

type Matchup struct {
	Player1        string
	Player2        string
	Player1Wins    int
	Player2Wins    int
	MatchesPlayed  int
	Player1WinRate float64
}

func New(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	storage := &Storage{db: db}
	if err := storage.createTables(); err != nil {
		return nil, err
	}

	// Migrate existing data
	if err := storage.migrateExistingData(); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS players (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			external_id INTEGER UNIQUE NOT NULL,
			display_name TEXT NOT NULL,
			username TEXT,
			current_elo INTEGER DEFAULT 1500,
			matches_played INTEGER DEFAULT 0,
			wins INTEGER DEFAULT 0,
			losses INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tournaments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			melee_id INTEGER UNIQUE NOT NULL,
			date DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS matches (
			id TEXT PRIMARY KEY,
			tournament_id INTEGER,
			round INTEGER,
			player1_id INTEGER,
			player2_id INTEGER,
			player1_wins INTEGER,
			player2_wins INTEGER,
			date_played DATETIME,
			player1_elo_before INTEGER,
			player2_elo_before INTEGER,
			player1_elo_after INTEGER,
			player2_elo_after INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (tournament_id) REFERENCES tournaments(id),
			FOREIGN KEY (player1_id) REFERENCES players(id),
			FOREIGN KEY (player2_id) REFERENCES players(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_matches_tournament ON matches(tournament_id)`,
		`CREATE INDEX IF NOT EXISTS idx_matches_date ON matches(date_played)`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}

func (s *Storage) migrateExistingData() error {
	// Check if we already have tournaments
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM tournaments").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		// Already migrated
		return nil
	}

	// Get distinct tournament_ids from matches
	rows, err := s.db.Query("SELECT DISTINCT tournament_id FROM matches")
	if err != nil {
		return err
	}
	defer rows.Close()

	// Create a tournament for each unique tournament_id
	for rows.Next() {
		var tournamentID int
		if err := rows.Scan(&tournamentID); err != nil {
			return err
		}

		// Use tournament_id as melee_id with a default date
		// We'll use a placeholder date - in the future this could be fetched from melee.gg
		_, err = s.db.Exec(
			"INSERT OR IGNORE INTO tournaments (melee_id, date) VALUES (?, datetime('2024-01-01'))",
			tournamentID,
		)
		if err != nil {
			return err
		}
	}

	return rows.Err()
}

func (s *Storage) GetOrCreatePlayer(externalID int64, displayName, username string) (*Player, error) {
	// Try to get existing player
	var player Player
	err := s.db.QueryRow(
		"SELECT id, external_id, display_name, username, current_elo, matches_played, wins, losses, created_at, updated_at FROM players WHERE external_id = ?",
		externalID,
	).Scan(&player.ID, &player.ExternalID, &player.DisplayName, &player.Username, &player.CurrentELO, &player.MatchesPlayed, &player.Wins, &player.Losses, &player.CreatedAt, &player.UpdatedAt)

	if err == nil {
		return &player, nil
	}

	if err != sql.ErrNoRows {
		return nil, err
	}

	// Create new player
	res, err := s.db.Exec(
		"INSERT INTO players (external_id, display_name, username) VALUES (?, ?, ?)",
		externalID, displayName, username,
	)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &Player{
		ID:          id,
		ExternalID:  externalID,
		DisplayName: displayName,
		Username:    username,
		CurrentELO:  1500,
	}, nil
}

func (s *Storage) GetPlayerByID(id int64) (*Player, error) {
	var player Player
	err := s.db.QueryRow(
		"SELECT id, external_id, display_name, username, current_elo, matches_played, wins, losses, created_at, updated_at FROM players WHERE id = ?",
		id,
	).Scan(&player.ID, &player.ExternalID, &player.DisplayName, &player.Username, &player.CurrentELO, &player.MatchesPlayed, &player.Wins, &player.Losses, &player.CreatedAt, &player.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &player, nil
}

func (s *Storage) UpdatePlayerELO(playerID int64, newELO int, won bool) error {
	query := `UPDATE players 
			  SET current_elo = ?, 
			      matches_played = matches_played + 1,
			      wins = wins + ?,
			      losses = losses + ?,
			      updated_at = CURRENT_TIMESTAMP
			  WHERE id = ?`

	winInc := 0
	lossInc := 0
	if won {
		winInc = 1
	} else {
		lossInc = 1
	}

	_, err := s.db.Exec(query, newELO, winInc, lossInc, playerID)
	return err
}

func (s *Storage) GetOrCreateTournament(meleeID int, date time.Time) (*Tournament, error) {
	var t Tournament
	var datePtr *time.Time
	err := s.db.QueryRow(
		"SELECT id, melee_id, date FROM tournaments WHERE melee_id = ?",
		meleeID,
	).Scan(&t.ID, &t.MeleeID, &datePtr)

	if err == nil {
		if datePtr != nil {
			t.Date = *datePtr
		}
		return &t, nil
	}

	if err != sql.ErrNoRows {
		return nil, err
	}

	// Handle zero time as null
	if !date.IsZero() {
		datePtr = &date
	}

	res, err := s.db.Exec(
		"INSERT INTO tournaments (melee_id, date) VALUES (?, ?)",
		meleeID, datePtr,
	)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &Tournament{
		ID:      id,
		MeleeID: meleeID,
		Date:    date,
	}, nil
}

func (s *Storage) GetTournamentByMeleeID(meleeID int) (*Tournament, error) {
	var t Tournament
	var datePtr *time.Time
	err := s.db.QueryRow(
		"SELECT id, melee_id, date FROM tournaments WHERE melee_id = ?",
		meleeID,
	).Scan(&t.ID, &t.MeleeID, &datePtr)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if datePtr != nil {
		t.Date = *datePtr
	}
	return &t, nil
}

func (s *Storage) GetTournamentsWithMissingDates() ([]Tournament, error) {
	rows, err := s.db.Query("SELECT id, melee_id, date FROM tournaments WHERE date IS NULL")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tournaments []Tournament
	for rows.Next() {
		var t Tournament
		var datePtr *time.Time
		if err := rows.Scan(&t.ID, &t.MeleeID, &datePtr); err != nil {
			return nil, err
		}
		if datePtr != nil {
			t.Date = *datePtr
		}
		tournaments = append(tournaments, t)
	}
	return tournaments, rows.Err()
}

func (s *Storage) UpdateTournamentDate(meleeID int, date time.Time) error {
	_, err := s.db.Exec("UPDATE tournaments SET date = ? WHERE melee_id = ?", date, meleeID)
	return err
}

func (s *Storage) ResetAllPlayersELO() error {
	_, err := s.db.Exec("UPDATE players SET current_elo = 1500, matches_played = 0, wins = 0, losses = 0")
	return err
}

func (s *Storage) GetAllMatchesSorted() ([]Match, error) {
	query := `
		SELECT m.id, m.tournament_id, m.round, m.player1_id, m.player2_id, 
		       m.player1_wins, m.player2_wins, m.date_played, t.date as tournament_date
		FROM matches m
		JOIN tournaments t ON m.tournament_id = t.melee_id
		ORDER BY COALESCE(t.date, '1970-01-01') ASC, m.tournament_id ASC, m.round ASC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []Match
	for rows.Next() {
		var m Match
		var tournamentDatePtr *time.Time
		err := rows.Scan(&m.ID, &m.TournamentID, &m.Round, &m.Player1ID, &m.Player2ID,
			&m.Player1Wins, &m.Player2Wins, &m.DatePlayed, &tournamentDatePtr)
		if err != nil {
			return nil, err
		}
		if tournamentDatePtr != nil {
			m.DatePlayed = *tournamentDatePtr
		}
		matches = append(matches, m)
	}

	return matches, rows.Err()
}

func (s *Storage) MatchExists(matchID string) (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM matches WHERE id = ?", matchID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *Storage) SaveMatch(match Match) error {
	_, err := s.db.Exec(
		`INSERT INTO matches (id, tournament_id, round, player1_id, player2_id, player1_wins, player2_wins, 
		date_played, player1_elo_before, player2_elo_before, player1_elo_after, player2_elo_after)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		match.ID, match.TournamentID, match.Round, match.Player1ID, match.Player2ID,
		match.Player1Wins, match.Player2Wins, match.DatePlayed,
		match.Player1ELOBefore, match.Player2ELOBefore, match.Player1ELOAfter, match.Player2ELOAfter,
	)
	return err
}

func (s *Storage) UpdateMatchELO(matchID string, player1ELOBefore, player2ELOBefore, player1ELOAfter, player2ELOAfter int) error {
	_, err := s.db.Exec(
		`UPDATE matches SET player1_elo_before = ?, player2_elo_before = ?, player1_elo_after = ?, player2_elo_after = ? WHERE id = ?`,
		player1ELOBefore, player2ELOBefore, player1ELOAfter, player2ELOAfter, matchID,
	)
	return err
}

func (s *Storage) GetRankings() ([]Ranking, error) {
	query := `SELECT 
		display_name, username, current_elo, matches_played, wins, losses
	  FROM players 
	  WHERE matches_played >= 10
	  ORDER BY current_elo DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rankings []Ranking
	rank := 1
	for rows.Next() {
		var r Ranking
		var username sql.NullString
		err := rows.Scan(&r.DisplayName, &username, &r.CurrentELO, &r.MatchesPlayed, &r.Wins, &r.Losses)
		if err != nil {
			return nil, err
		}

		r.Rank = rank
		rank++

		if username.Valid {
			r.Username = username.String
		}

		if r.MatchesPlayed > 0 {
			r.WinRate = float64(r.Wins) / float64(r.MatchesPlayed) * 100
		}

		rankings = append(rankings, r)
	}

	return rankings, rows.Err()
}

type PlayerMatch struct {
	DatePlayed      time.Time
	TournamentID    int
	Round           int
	OpponentName    string
	PlayerWins      int
	OpponentWins    int
	PlayerELOBefore int
	PlayerELOAfter  int
	Result          string
}

func (s *Storage) GetPlayerMatchHistory(displayName string) ([]PlayerMatch, error) {
	query := `
		SELECT 
			t.date as tournament_date,
			m.round,
			CASE 
				WHEN p1.display_name = ? THEN p2.display_name
				ELSE p1.display_name
			END as opponent_name,
			CASE 
				WHEN p1.display_name = ? THEN m.player1_wins
				ELSE m.player2_wins
			END as player_wins,
			CASE 
				WHEN p1.display_name = ? THEN m.player2_wins
				ELSE m.player1_wins
			END as opponent_wins,
			CASE 
				WHEN p1.display_name = ? THEN m.player1_elo_before
				ELSE m.player2_elo_before
			END as player_elo_before,
			CASE 
				WHEN p1.display_name = ? THEN m.player1_elo_after
				ELSE m.player2_elo_after
			END as player_elo_after
		FROM matches m
		JOIN players p1 ON m.player1_id = p1.id
		JOIN players p2 ON m.player2_id = p2.id
		JOIN tournaments t ON m.tournament_id = t.melee_id
		WHERE (p1.display_name = ? OR p2.display_name = ?)
		  AND t.date IS NOT NULL
		ORDER BY t.date ASC, m.round ASC
	`

	rows, err := s.db.Query(query, displayName, displayName, displayName, displayName, displayName, displayName, displayName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []PlayerMatch
	for rows.Next() {
		var m PlayerMatch
		err := rows.Scan(
			&m.DatePlayed,
			&m.Round,
			&m.OpponentName,
			&m.PlayerWins,
			&m.OpponentWins,
			&m.PlayerELOBefore,
			&m.PlayerELOAfter,
		)
		if err != nil {
			return nil, err
		}

		// Determine result
		if m.PlayerWins > m.OpponentWins {
			m.Result = "Win"
		} else if m.PlayerWins < m.OpponentWins {
			m.Result = "Loss"
		} else {
			m.Result = "Draw"
		}

		matches = append(matches, m)
	}

	return matches, rows.Err()
}

func (s *Storage) GetMatchups() ([]Matchup, error) {
	query := `
		SELECT 
			p1.display_name as player1,
			p2.display_name as player2,
			SUM(CASE WHEN m.player1_wins > m.player2_wins THEN 1 ELSE 0 END) as player1_wins,
			SUM(CASE WHEN m.player2_wins > m.player1_wins THEN 1 ELSE 0 END) as player2_wins,
			COUNT(*) as match_count
		FROM matches m
		JOIN players p1 ON m.player1_id = p1.id
		JOIN players p2 ON m.player2_id = p2.id
		GROUP BY p1.display_name, p2.display_name
		HAVING COUNT(*) >= 2
		ORDER BY p1.display_name, p2.display_name
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matchups []Matchup
	for rows.Next() {
		var m Matchup
		var matchCount int
		err := rows.Scan(&m.Player1, &m.Player2, &m.Player1Wins, &m.Player2Wins, &matchCount)
		if err != nil {
			return nil, err
		}
		m.MatchesPlayed = matchCount

		if m.MatchesPlayed > 0 {
			m.Player1WinRate = float64(m.Player1Wins) / float64(m.MatchesPlayed) * 100
		}

		matchups = append(matchups, m)
	}

	return matchups, rows.Err()
}
