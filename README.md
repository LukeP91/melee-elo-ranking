# Melee ELO Rankings

A Go-based application that calculates ELO rankings from melee.gg tournament match data and generates a static HTML page for GitHub Pages.

## Features

- ELO ranking calculation (K-factor: 32, starting ELO: 1500)
- SQLite database for persistent storage
- Processes multiple tournaments in chronological order
- Generates responsive HTML ranking page
- GitHub Pages ready

## Usage

1. Place tournament JSON files in `data/matches-pending/`
2. Run the application:
   ```bash
   make run
   # or
   go run ./cmd/elo-cli
   ```
3. Generated rankings will be in `docs/index.html`
4. Commit and push the `docs/` folder to GitHub for Pages hosting

## Configuration

Edit `config.json` to customize:
- K-factor for ELO calculations
- Initial ELO rating
- Output file path

## Data Flow

```
data/matches-pending/*.json → Process → SQLite → docs/index.html
                              ↓
                    data/matches-processed/
```

## Project Structure

- `cmd/elo-cli/` - CLI entry point
- `internal/` - Application logic
  - `config/` - Configuration management
  - `parser/` - JSON match parsing
  - `elo/` - ELO calculation engine
  - `storage/` - SQLite operations
  - `output/` - Output interface and implementations
  - `generator/` - HTML generation
- `data/` - Local data storage (not in git)
- `docs/` - Generated HTML for GitHub Pages

## License

MIT
