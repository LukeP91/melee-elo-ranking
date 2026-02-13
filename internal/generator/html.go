package generator

import (
	"fmt"
	"os"
	"time"

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
	html := g.generateHTML(rankings)
	return os.WriteFile(outputPath, []byte(html), 0644)
}

func (g *Generator) GeneratePlayerPage(playerName string, matches []storage.PlayerMatch, playerStats storage.Ranking, outputPath string) error {
	html := g.generatePlayerHTML(playerName, matches, playerStats)
	return os.WriteFile(outputPath, []byte(html), 0644)
}

func (g *Generator) generateHTML(rankings []storage.Ranking) string {
	timestamp := time.Now().Format("January 2, 2006 15:04")

	tableRows := ""
	for _, r := range rankings {
		winRateClass := "neutral"
		if r.WinRate >= 60 {
			winRateClass = "positive"
		} else if r.WinRate < 40 {
			winRateClass = "negative"
		}

		tableRows += fmt.Sprintf(`
				<tr>
					<td class="rank">%d</td>
					<td class="player"><a href="players/%s.html">%s</a></td>
					<td class="elo">%d</td>
					<td class="matches">%d</td>
					<td class="record">%d-%d</td>
					<td class="winrate %s">%.1f%%</td>
				</tr>`,
			r.Rank,
			r.DisplayName,
			r.DisplayName,
			r.CurrentELO,
			r.MatchesPlayed,
			r.Wins,
			r.Losses,
			winRateClass,
			r.WinRate,
		)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #1a1a2e 0%%, #16213e 100%%);
            min-height: 100vh;
            padding: 2rem 1rem;
            color: #eee;
        }
        
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        
        header {
            text-align: center;
            margin-bottom: 2rem;
        }
        
        h1 {
            font-size: 2.5rem;
            margin-bottom: 0.5rem;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        
        .subtitle {
            color: #888;
            font-size: 1.1rem;
        }
        
        .last-updated {
            text-align: center;
            color: #666;
            font-size: 0.9rem;
            margin-bottom: 2rem;
        }
        
        .rankings-table {
            width: 100%%;
            border-collapse: collapse;
            background: rgba(255, 255, 255, 0.05);
            border-radius: 12px;
            overflow: hidden;
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
        }
        
        .rankings-table thead {
            background: rgba(102, 126, 234, 0.2);
        }
        
        .rankings-table th {
            padding: 1rem;
            text-align: left;
            font-weight: 600;
            text-transform: uppercase;
            font-size: 0.85rem;
            letter-spacing: 0.5px;
            color: #a0a0a0;
            cursor: pointer;
            user-select: none;
        }
        
        .rankings-table th:hover {
            color: #fff;
            background: rgba(102, 126, 234, 0.3);
        }
        
        .rankings-table td {
            padding: 1rem;
            border-bottom: 1px solid rgba(255, 255, 255, 0.1);
        }
        
        .rankings-table tbody tr:hover {
            background: rgba(255, 255, 255, 0.05);
        }
        
        .rankings-table tbody tr:last-child td {
            border-bottom: none;
        }
        
        .rank {
            font-weight: 700;
            font-size: 1.2rem;
            color: #667eea;
            width: 60px;
        }
        
        .player {
            font-weight: 500;
        }
        
        .elo {
            font-weight: 700;
            font-size: 1.1rem;
            color: #fff;
        }
        
        .matches, .record {
            color: #aaa;
        }
        
        .winrate {
            font-weight: 600;
        }
        
        .winrate.positive {
            color: #4ade80;
        }
        
        .winrate.negative {
            color: #f87171;
        }
        
        .winrate.neutral {
            color: #fbbf24;
        }
        
        @media (max-width: 768px) {
            h1 {
                font-size: 1.8rem;
            }
            
            .rankings-table {
                font-size: 0.9rem;
            }
            
            .rankings-table th,
            .rankings-table td {
                padding: 0.75rem 0.5rem;
            }
            
            .rank {
                width: 40px;
            }
        }
        
        .footer {
            text-align: center;
            margin-top: 2rem;
            color: #666;
            font-size: 0.85rem;
        }
        
        .footer a {
            color: #667eea;
            text-decoration: none;
        }
        
        .footer a:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>%s</h1>
            <p class="subtitle">%s</p>
        </header>
        
        <p class="last-updated">Last updated: %s</p>
        
        <table class="rankings-table">
            <thead>
                <tr>
                    <th>Rank</th>
                    <th>Player</th>
                    <th>ELO</th>
                    <th>Matches</th>
                    <th>W-L</th>
                    <th>Win %%</th>
                </tr>
            </thead>
            <tbody>
                %s
            </tbody>
        </table>
        
        <div class="footer">
            <p>Powered by <a href="https://github.com/melee-elo-ranking">Melee ELO Rankings</a></p>
        </div>
    </div>
    
    <script>
        // Simple table sorting
        document.querySelectorAll('.rankings-table th').forEach(header => {
            header.addEventListener('click', () => {
                const table = header.closest('table');
                const tbody = table.querySelector('tbody');
                const rows = Array.from(tbody.querySelectorAll('tr'));
                const index = Array.from(header.parentNode.children).indexOf(header);
                const isNumeric = index > 0;
                
                rows.sort((a, b) => {
                    const aVal = a.children[index].textContent.trim();
                    const bVal = b.children[index].textContent.trim();
                    
                    if (isNumeric) {
                        return parseFloat(aVal) - parseFloat(bVal);
                    }
                    return aVal.localeCompare(bVal);
                });
                
                rows.forEach(row => tbody.appendChild(row));
            });
        });
    </script>
</body>
</html>`,
		g.title,
		g.title,
		g.description,
		timestamp,
		tableRows,
	)
}

func (g *Generator) generatePlayerHTML(playerName string, matches []storage.PlayerMatch, playerStats storage.Ranking) string {
	// Generate match history rows
	matchRows := ""
	for _, m := range matches {
		resultClass := "neutral"
		if m.Result == "Win" {
			resultClass = "positive"
		} else if m.Result == "Loss" {
			resultClass = "negative"
		}

		matchRows += fmt.Sprintf(`
				<tr>
					<td>Tournament %d</td>
					<td>Round %d</td>
					<td>%s</td>
					<td>%d-%d</td>
					<td class="%s">%s</td>
					<td>%d</td>
					<td>%d</td>
				</tr>`,
			m.TournamentID,
			m.Round,
			m.OpponentName,
			m.PlayerWins,
			m.OpponentWins,
			resultClass,
			m.Result,
			m.PlayerELOBefore,
			m.PlayerELOAfter,
		)
	}

	// Generate ELO progression data for chart
	eloData := ""
	for i, m := range matches {
		if i > 0 {
			eloData += ", "
		}
		eloData += fmt.Sprintf("%d", m.PlayerELOAfter)
	}

	winRateClass := "neutral"
	if playerStats.WinRate >= 60 {
		winRateClass = "positive"
	} else if playerStats.WinRate < 40 {
		winRateClass = "negative"
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - Melee ELO Profile</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #1a1a2e 0%%, #16213e 100%%);
            min-height: 100vh;
            padding: 2rem 1rem;
            color: #eee;
        }
        
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        
        .back-link {
            margin-bottom: 1rem;
        }
        
        .back-link a {
            color: #667eea;
            text-decoration: none;
            font-size: 0.9rem;
        }
        
        .back-link a:hover {
            text-decoration: underline;
        }
        
        header {
            text-align: center;
            margin-bottom: 2rem;
        }
        
        h1 {
            font-size: 2.5rem;
            margin-bottom: 0.5rem;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
            gap: 1rem;
            margin-bottom: 2rem;
        }
        
        .stat-card {
            background: rgba(255, 255, 255, 0.05);
            border-radius: 12px;
            padding: 1.5rem;
            text-align: center;
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
        }
        
        .stat-value {
            font-size: 2rem;
            font-weight: 700;
            color: #667eea;
            margin-bottom: 0.5rem;
        }
        
        .stat-label {
            color: #888;
            font-size: 0.9rem;
        }
        
        .stat-value.positive {
            color: #4ade80;
        }
        
        .stat-value.negative {
            color: #f87171;
        }
        
        .stat-value.neutral {
            color: #fbbf24;
        }
        
        .section {
            background: rgba(255, 255, 255, 0.05);
            border-radius: 12px;
            padding: 1.5rem;
            margin-bottom: 2rem;
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
        }
        
        .section h2 {
            margin-bottom: 1rem;
            color: #667eea;
        }
        
        .chart-container {
            width: 100%%;
            height: 300px;
            margin: 2rem 0;
        }
        
        .matches-table {
            width: 100%%;
            border-collapse: collapse;
        }
        
        .matches-table th {
            padding: 1rem;
            text-align: left;
            font-weight: 600;
            text-transform: uppercase;
            font-size: 0.85rem;
            letter-spacing: 0.5px;
            color: #a0a0a0;
            border-bottom: 1px solid rgba(255, 255, 255, 0.1);
        }
        
        .matches-table td {
            padding: 1rem;
            border-bottom: 1px solid rgba(255, 255, 255, 0.05);
        }
        
        .matches-table tbody tr:hover {
            background: rgba(255, 255, 255, 0.03);
        }
        
        .positive {
            color: #4ade80;
            font-weight: 600;
        }
        
        .negative {
            color: #f87171;
            font-weight: 600;
        }
        
        .neutral {
            color: #fbbf24;
            font-weight: 600;
        }
        
        @media (max-width: 768px) {
            h1 {
                font-size: 1.8rem;
            }
            
            .matches-table {
                font-size: 0.85rem;
            }
            
            .matches-table th,
            .matches-table td {
                padding: 0.75rem 0.5rem;
            }
            
            .chart-container {
                height: 200px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="back-link">
            <a href="../index.html">&larr; Back to Rankings</a>
        </div>
        
        <header>
            <h1>%s</h1>
        </header>
        
        <div class="stats-grid">
            <div class="stat-card">
                <div class="stat-value">%d</div>
                <div class="stat-label">Current ELO</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">%d</div>
                <div class="stat-label">Rank</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">%d</div>
                <div class="stat-label">Matches</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">%d-%d</div>
                <div class="stat-label">Record</div>
            </div>
            <div class="stat-card">
                <div class="stat-value %s">%.1f%%</div>
                <div class="stat-label">Win Rate</div>
            </div>
        </div>
        
        <div class="section">
            <h2>ELO Progression</h2>
            <div class="chart-container">
                %s
            </div>
        </div>
        
        <div class="section">
            <h2>Match History</h2>
            <table class="matches-table">
                <thead>
                    <tr>
                        <th>Tournament</th>
                        <th>Round</th>
                        <th>Opponent</th>
                        <th>Score</th>
                        <th>Result</th>
                        <th>ELO Before</th>
                        <th>ELO After</th>
                    </tr>
                </thead>
                <tbody>
                    %s
                </tbody>
            </table>
        </div>
    </div>
</body>
</html>`,
		playerName,
		playerName,
		playerStats.CurrentELO,
		playerStats.Rank,
		playerStats.MatchesPlayed,
		playerStats.Wins,
		playerStats.Losses,
		winRateClass,
		playerStats.WinRate,
		g.generateELOChart(matches),
		matchRows,
	)
}

func (g *Generator) generateELOChart(matches []storage.PlayerMatch) string {
	if len(matches) == 0 {
		return "<p>No match data available</p>"
	}

	width := 800
	height := 300
	padding := 50

	chartWidth := width - 2*padding
	chartHeight := height - 2*padding

	// Find min and max ELO values
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

	// Add some padding to the range
	eloRange := maxELO - minELO
	if eloRange == 0 {
		eloRange = 100
	}
	minELO -= eloRange / 10
	maxELO += eloRange / 10
	eloRange = maxELO - minELO

	// Generate points
	points := ""
	for i, m := range matches {
		x := padding + (i * chartWidth / (len(matches) - 1))
		y := height - padding - ((m.PlayerELOAfter - minELO) * chartHeight / eloRange)
		if i > 0 {
			points += " "
		}
		points += fmt.Sprintf("%d,%d", x, y)
	}

	// Generate axis labels
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
