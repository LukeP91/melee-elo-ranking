package generator

import (
	"time"

	"github.com/melee-elo-ranking/internal/storage"
)

// IndexData is the data passed to the index template.
type IndexData struct {
	Title     string
	Subtitle  string
	Timestamp string
	Rankings  []IndexRankingRow
}

// IndexRankingRow is one row in the rankings table.
type IndexRankingRow struct {
	Rank          int
	DisplayName   string
	CurrentELO    int
	MatchesPlayed int
	Wins          int
	Losses        int
	WinRate       float64
	WinRateClass  string
}

func (g *Generator) buildIndexData(rankings []storage.Ranking) IndexData {
	rows := make([]IndexRankingRow, 0, len(rankings))
	for _, r := range rankings {
		winRateClass := "neutral"
		if r.WinRate >= 60 {
			winRateClass = "positive"
		} else if r.WinRate < 40 {
			winRateClass = "negative"
		}
		rows = append(rows, IndexRankingRow{
			Rank:          r.Rank,
			DisplayName:   r.DisplayName,
			CurrentELO:    r.CurrentELO,
			MatchesPlayed: r.MatchesPlayed,
			Wins:          r.Wins,
			Losses:        r.Losses,
			WinRate:       r.WinRate,
			WinRateClass:  winRateClass,
		})
	}
	return IndexData{
		Title:     g.title,
		Subtitle:  g.description,
		Timestamp: time.Now().Format("January 2, 2006 15:04"),
		Rankings:  rows,
	}
}

func (g *Generator) renderIndex(rankings []storage.Ranking) ([]byte, error) {
	data := g.buildIndexData(rankings)
	return executeTemplate("templates/index.tmpl", data)
}
