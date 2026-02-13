package parser

import (
	"encoding/json"
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

func (p *Parser) ParseFile(filepath string) ([]Match, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var rawMatches []RawMatch
	if err := json.Unmarshal(data, &rawMatches); err != nil {
		return nil, err
	}

	var matches []Match
	for _, raw := range rawMatches {
		match := p.convertRawMatch(raw)
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

func (p *Parser) convertRawMatch(raw RawMatch) Match {
	match := Match{
		ID:           raw.Guid,
		TournamentID: raw.TournamentId,
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
