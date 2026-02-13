package generator

import (
	"fmt"
	"html/template"
	"time"

	"github.com/melee-elo-ranking/internal/storage"
)

// MatchupData is the data passed to the matchup template.
type MatchupData struct {
	Timestamp         string
	MatrixHeaderHTML  template.HTML
	MatrixBodyHTML    template.HTML
}

func (g *Generator) buildMatchupData(matchups []storage.Matchup, players []string) MatchupData {
	matchupMap := make(map[string]map[string]storage.Matchup)
	for _, m := range matchups {
		if matchupMap[m.Player1] == nil {
			matchupMap[m.Player1] = make(map[string]storage.Matchup)
		}
		matchupMap[m.Player1][m.Player2] = m
	}
	return MatchupData{
		Timestamp:        time.Now().Format("January 2, 2006 15:04"),
		MatrixHeaderHTML: template.HTML(generateMatrixHeader(players)),
		MatrixBodyHTML:   template.HTML(generateMatrixBody(players, matchupMap)),
	}
}

func (g *Generator) renderMatchup(matchups []storage.Matchup, players []string) ([]byte, error) {
	data := g.buildMatchupData(matchups, players)
	return executeTemplate("templates/matchup.tmpl", data)
}

func generateMatrixHeader(players []string) string {
	header := ""
	for _, player := range players {
		shortName := player
		if len(shortName) > 8 {
			shortName = shortName[:8] + "."
		}
		header += fmt.Sprintf(`<th title="%s">%s</th>`, template.HTMLEscapeString(player), template.HTMLEscapeString(shortName))
	}
	return header
}

func generateMatrixBody(players []string, matchupMap map[string]map[string]storage.Matchup) string {
	body := ""
	for _, player1 := range players {
		body += "<tr>"
		body += fmt.Sprintf(`<th class="row-header" title="%s">%s</th>`, template.HTMLEscapeString(player1), template.HTMLEscapeString(player1))
		for _, player2 := range players {
			if player1 == player2 {
				body += `<td class="empty">-</td>`
			} else {
				m, found := matchupMap[player1][player2]
				if !found || m.GamesPlayed == 0 {
					body += `<td class="empty">-</td>`
				} else {
					cellClass := "cell-neutral"
					if m.Player1WinRate >= 60 {
						cellClass = "cell-positive"
					} else if m.Player1WinRate < 40 {
						cellClass = "cell-negative"
					}
					body += fmt.Sprintf(`<td class="cell %s" data-games="%d" data-detail="%s vs %s: %d-%d">%.0f%%</td>`,
						cellClass,
						m.GamesPlayed,
						template.HTMLEscapeString(player1), template.HTMLEscapeString(player2),
						m.Player1Wins, m.Player2Wins,
						m.Player1WinRate)
				}
			}
		}
		body += "</tr>"
	}
	return body
}
