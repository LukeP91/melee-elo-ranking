package generator

import (
	"fmt"
	"html/template"
	"time"

	"github.com/melee-elo-ranking/internal/storage"
)

// PlayerData is the data passed to the player template.
type PlayerData struct {
	PlayerName    string
	Timestamp     string
	CurrentELO    int
	Rank          int
	MatchesPlayed int
	Wins          int
	Losses        int
	WinRate       float64
	WinRateClass  string
	ELOChartHTML  template.HTML
	Matches       []PlayerMatchRow
}

// PlayerMatchRow is one row in the match history table.
type PlayerMatchRow struct {
	Date           string
	Round          int
	OpponentName   string
	PlayerWins     int
	OpponentWins   int
	Result         string
	ResultClass    string
	PlayerELOBefore int
	PlayerELOAfter  int
}

func (g *Generator) buildPlayerData(playerName string, matches []storage.PlayerMatch, playerStats storage.Ranking) PlayerData {
	rows := make([]PlayerMatchRow, 0, len(matches))
	for _, m := range matches {
		resultClass := "neutral"
		if m.Result == "Win" {
			resultClass = "positive"
		} else if m.Result == "Loss" {
			resultClass = "negative"
		}
		rows = append(rows, PlayerMatchRow{
			Date:            m.DatePlayed.Format("Jan 2, 2006"),
			Round:           m.Round,
			OpponentName:    m.OpponentName,
			PlayerWins:     m.PlayerWins,
			OpponentWins:   m.OpponentWins,
			Result:         m.Result,
			ResultClass:    resultClass,
			PlayerELOBefore: m.PlayerELOBefore,
			PlayerELOAfter:  m.PlayerELOAfter,
		})
	}

	winRateClass := "neutral"
	if playerStats.WinRate >= 60 {
		winRateClass = "positive"
	} else if playerStats.WinRate < 40 {
		winRateClass = "negative"
	}

	return PlayerData{
		PlayerName:    playerName,
		Timestamp:     time.Now().Format("January 2, 2006 15:04"),
		CurrentELO:    playerStats.CurrentELO,
		Rank:          playerStats.Rank,
		MatchesPlayed: playerStats.MatchesPlayed,
		Wins:          playerStats.Wins,
		Losses:        playerStats.Losses,
		WinRate:       playerStats.WinRate,
		WinRateClass:  winRateClass,
		ELOChartHTML:  template.HTML(generateELOChart(matches)),
		Matches:       rows,
	}
}

func (g *Generator) renderPlayer(playerName string, matches []storage.PlayerMatch, playerStats storage.Ranking) ([]byte, error) {
	data := g.buildPlayerData(playerName, matches, playerStats)
	return executeTemplate("templates/player.tmpl", data)
}

func generateELOChart(matches []storage.PlayerMatch) string {
	if len(matches) == 0 {
		return "<p>No match data available</p>"
	}

	width := 800
	height := 300
	padding := 50

	chartWidth := width - 2*padding
	chartHeight := height - 2*padding

	minELO := matches[0].PlayerELOBefore
	maxELO := matches[0].PlayerELOBefore
	for _, m := range matches {
		if m.PlayerELOBefore < minELO {
			minELO = m.PlayerELOBefore
		}
		if m.PlayerELOAfter < minELO {
			minELO = m.PlayerELOAfter
		}
		if m.PlayerELOBefore > maxELO {
			maxELO = m.PlayerELOBefore
		}
		if m.PlayerELOAfter > maxELO {
			maxELO = m.PlayerELOAfter
		}
	}

	eloRange := maxELO - minELO
	if eloRange == 0 {
		eloRange = 100
	}
	minELO -= eloRange / 10
	maxELO += eloRange / 10
	eloRange = maxELO - minELO

	points := ""
	for i, m := range matches {
		x := padding + (i * chartWidth / (len(matches) - 1))
		y := height - padding - ((m.PlayerELOAfter - minELO) * chartHeight / eloRange)
		if i > 0 {
			points += " "
		}
		points += fmt.Sprintf("%d,%d", x, y)
	}

	yAxisLabels := ""
	for i := 0; i <= 5; i++ {
		value := minELO + (eloRange * i / 5)
		y := height - padding - (chartHeight * i / 5)
		yAxisLabels += fmt.Sprintf(`<text x="%d" y="%d" text-anchor="end" fill="#888" font-size="12">%d</text>`, padding-10, int(y)+4, value)
	}

	return fmt.Sprintf(`<svg viewBox="0 0 %d %d" style="width:100%%;height:100%%;">
		<defs>
			<linearGradient id="lineGradient" x1="0%%" y1="0%%" x2="100%%" y2="0%%">
				<stop offset="0%%" style="stop-color:#667eea" />
				<stop offset="100%%" style="stop-color:#764ba2" />
			</linearGradient>
		</defs>
		<rect x="0" y="0" width="%d" height="%d" fill="rgba(255,255,255,0.02)" rx="8" />
		<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="rgba(255,255,255,0.1)" stroke-width="1" />
		<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="rgba(255,255,255,0.1)" stroke-width="1" />
		%s
		<polyline points="%s" fill="none" stroke="url(#lineGradient)" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" />
		<text x="%d" y="%d" text-anchor="middle" fill="#667eea" font-size="14" font-weight="600">ELO Over Time (by Tournament)</text>
	</svg>`,
		width, height,
		width, height,
		padding, padding, padding, height-padding,
		padding, height-padding, width-padding, height-padding,
		yAxisLabels,
		points,
		width/2, padding-15,
	)
}
