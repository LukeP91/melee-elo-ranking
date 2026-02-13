package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Match struct {
	ID           string
	TournamentID int
	RoundNumber  int
	DateCreated  time.Time
	Competitors  []Competitor
}

type Competitor struct {
	Player   Player
	GameWins int
}

type Player struct {
	ID          int64
	DisplayName string
	Username    string
}

type Parser struct{}

func New() *Parser {
	return &Parser{}
}

func (p *Parser) ParseFile(filepath string, tournamentID int) ([]Match, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	// Try new format first (V2)
	var rawMatchesV2 []RawMatchV2
	if err := json.Unmarshal(data, &rawMatchesV2); err == nil && len(rawMatchesV2) > 0 {
		var matches []Match
		for _, raw := range rawMatchesV2 {
			match := p.convertRawMatchV2(raw, tournamentID)
			if match.ID != "" {
				matches = append(matches, match)
			}
		}
		return matches, nil
	}

	// Fall back to old format
	var rawMatches []RawMatch
	if err := json.Unmarshal(data, &rawMatches); err != nil {
		return nil, err
	}

	var matches []Match
	for _, raw := range rawMatches {
		match := p.convertRawMatch(raw, tournamentID)
		matches = append(matches, match)
	}

	return matches, nil
}

type RawMatch struct {
	Guid         string    `json:"Guid"`
	TournamentId int       `json:"TournamentId"`
	RoundNumber  int       `json:"RoundNumber"`
	DateCreated  time.Time `json:"DateCreated"`
	Competitors  []struct {
		Team struct {
			Players []struct {
				ID          int64  `json:"ID"`
				DisplayName string `json:"DisplayName"`
				Username    string `json:"Username"`
			} `json:"Players"`
		} `json:"Team"`
		GameWins *int `json:"GameWins"`
	} `json:"Competitors"`
}

type RawMatchV2 struct {
	RoundNumber          int    `json:"RoundNumber"`
	PhaseId              int    `json:"PhaseId"`
	TableNumber          *int   `json:"TableNumber"`
	Team1Id              int64  `json:"Team1Id"`
	Team1                string `json:"Team1"`
	Player1NameLastFirst string `json:"Player1NameLastFirst"`
	Team1WinsAndByes     int    `json:"Team1WinsAndByes"`
	Team2Id              int64  `json:"Team2Id"`
	Team2                string `json:"Team2"`
	Player2NameLastFirst string `json:"Player2NameLastFirst"`
	Team2WinsAndByes     int    `json:"Team2WinsAndByes"`
	GameDraws            *int   `json:"GameDraws"`
	MatchesPublished     bool   `json:"MatchesPublished"`
	HasResult            bool   `json:"HasResult"`
	ByeReason            *int   `json:"ByeReason"`
	AdminResultString    string `json:"AdminResultString"`
	TimeExtensionMinutes *int   `json:"TimeExtensionMinutes"`
}

func (p *Parser) convertRawMatch(raw RawMatch, tournamentID int) Match {
	match := Match{
		ID:           raw.Guid,
		TournamentID: tournamentID,
		RoundNumber:  raw.RoundNumber,
		DateCreated:  raw.DateCreated,
	}

	for _, rawComp := range raw.Competitors {
		if len(rawComp.Team.Players) == 0 {
			continue
		}
		player := rawComp.Team.Players[0]

		gameWins := 0
		if rawComp.GameWins != nil {
			gameWins = *rawComp.GameWins
		}

		comp := Competitor{
			Player: Player{
				ID:          player.ID,
				DisplayName: player.DisplayName,
				Username:    player.Username,
			},
			GameWins: gameWins,
		}
		match.Competitors = append(match.Competitors, comp)
	}

	return match
}

func (p *Parser) convertRawMatchV2(raw RawMatchV2, tournamentID int) Match {
	// Generate a unique ID from the match data
	matchID := fmt.Sprintf("%d-%d-%d-%d", raw.PhaseId, raw.RoundNumber, raw.Team1Id, raw.Team2Id)

	match := Match{
		ID:           matchID,
		TournamentID: tournamentID,
		RoundNumber:  raw.RoundNumber,
		DateCreated:  time.Now(), // No date in V2 format
	}

	// Skip bye matches
	if raw.ByeReason != nil {
		return Match{}
	}

	// Use nickname hash as consistent player ID across tournaments
	player1ID := hashString(raw.Team1)
	player2ID := hashString(raw.Team2)

	// Player 1 - use nickname (Team1) as display name for GDPR
	comp1 := Competitor{
		Player: Player{
			ID:          player1ID,
			DisplayName: raw.Team1,
			Username:    raw.Team1,
		},
		GameWins: raw.Team1WinsAndByes,
	}
	match.Competitors = append(match.Competitors, comp1)

	// Player 2 - use nickname (Team2) as display name for GDPR
	comp2 := Competitor{
		Player: Player{
			ID:          player2ID,
			DisplayName: raw.Team2,
			Username:    raw.Team2,
		},
		GameWins: raw.Team2WinsAndByes,
	}
	match.Competitors = append(match.Competitors, comp2)

	return match
}

// hashString creates a consistent int64 hash from a string
func hashString(s string) int64 {
	h := int64(0)
	for _, c := range s {
		h = 31*h + int64(c)
	}
	return h
}
