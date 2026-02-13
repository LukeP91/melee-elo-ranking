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
	ID               string
	TournamentID     int
	Round            int
	Player1ID        int64
	Player2ID        int64
	Player1Wins      int
	Player2Wins      int
	DatePlayed       time.Time
	Player1ELOBefore int
	Player2ELOBefore int
	Player1ELOAfter  int
	Player2ELOAfter  int
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
			m.date_played,
			m.tournament_id,
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
		WHERE p1.display_name = ? OR p2.display_name = ?
		ORDER BY m.date_played, m.tournament_id, m.round
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
			&m.TournamentID,
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
			COUNT(*) as matches_played
		FROM matches m
		JOIN players p1 ON m.player1_id = p1.id
		JOIN players p2 ON m.player2_id = p2.id
		GROUP BY p1.display_name, p2.display_name
		HAVING matches_played >= 2
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
		err := rows.Scan(&m.Player1, &m.Player2, &m.Player1Wins, &m.Player2Wins, &m.MatchesPlayed)
		if err != nil {
			return nil, err
		}

		if m.MatchesPlayed > 0 {
			m.Player1WinRate = float64(m.Player1Wins) / float64(m.MatchesPlayed) * 100
		}

		matchups = append(matchups, m)
	}

	return matchups, rows.Err()
}
