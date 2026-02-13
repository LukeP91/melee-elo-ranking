package generator

import (
	"os"

	"github.com/melee-elo-ranking/internal/storage"
)

type Generator struct {
	title       string
	description string
}

func New(title, description string) *Generator {
	return &Generator{
		title:       title,
		description: description,
	}
}

func (g *Generator) Generate(rankings []storage.Ranking, outputPath string) error {
	buf, err := g.renderIndex(rankings)
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, buf, 0644)
}

func (g *Generator) GeneratePlayerPage(playerName string, matches []storage.PlayerMatch, playerStats storage.Ranking, outputPath string) error {
	buf, err := g.renderPlayer(playerName, matches, playerStats)
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, buf, 0644)
}

func (g *Generator) GenerateMatchupMatrix(matchups []storage.Matchup, players []string, outputPath string) error {
	buf, err := g.renderMatchup(matchups, players)
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, buf, 0644)
}
