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
					<td class="player">%s</td>
					<td class="elo">%d</td>
					<td class="matches">%d</td>
					<td class="record">%d-%d</td>
					<td class="winrate %s">%.1f%%</td>
				</tr>`,
			r.Rank,
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
