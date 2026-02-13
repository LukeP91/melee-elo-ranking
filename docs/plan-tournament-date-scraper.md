# Plan: Wire Tournament Date Scraper into Processing

## Current state

- **Scraper exists but is unused**: [internal/melee/client.go](internal/melee/client.go) implements `FetchTournamentDate(meleeID int) (time.Time, error)`:
  - GETs `https://melee.gg/Tournament/View/{meleeID}`
  - Parses date from HTML: regex for `data-toggle="datetime" data-value="..."` (e.g. `8/31/2024 7:00:00 AM`)
  - Tries several time formats and returns `time.Time`
- **Processor** ([cmd/elo-cli/processor.go](cmd/elo-cli/processor.go)) resolves tournament date in this order:
  1. `-dates` flag (e.g. `170676=2024-08-31`)
  2. Existing tournament in DB (if it already has a date)
  3. **Interactive prompt** (stdin) – no scraping today
- **Processor** does not receive or use the melee client; [cmd/elo-cli/main.go](cmd/elo-cli/main.go) builds the processor without it.

## Goal

Use the existing scraper when a tournament date is missing, so that:
- For each pending tournament file, if the date is not from the flag and not in the DB, **fetch the date from melee.gg** first.
- Only fall back to the interactive prompt (or skip) when the scrape fails or scraping is disabled.

## Proposed flow

```
For each tournament file:
  1. If date from -dates flag → use it
  2. Else if existing tournament in DB has date → use it
  3. Else if scraper enabled → GET https://melee.gg/Tournament/View/{id}, parse date
     - Success → use scraped date, save tournament
     - Failure → fall back to step 4
  4. Else (or scrape failed) → prompt user for date (current behaviour)
```

## Implementation

### 1. Processor: optional date fetcher

- Add an optional dependency to the processor for “fetch date from URL”:
  - Either a **concrete** `*melee.Client` (simplest), or
  - An **interface** e.g. `TournamentDateFetcher` with `FetchTournamentDate(meleeID int) (time.Time, error)` so tests can mock and the processor stays in `cmd/elo-cli` without importing melee if you later move it.
- **Recommendation**: Add `meleeClient *melee.Client` to `Processor` and `NewProcessor`. If `meleeClient` is nil, skip scraping and keep current behaviour (prompt only).

### 2. Processor: use scraper in date resolution

- In the loop where you resolve `tournamentDate` for each `tf`:
  - After “existing tournament in DB” and before “prompt”:
    - If `p.meleeClient != nil`, call `p.meleeClient.FetchTournamentDate(tf.tournamentID)`.
    - If err == nil and the returned time is not zero, set `tournamentDate = t` and continue (no prompt).
    - If err != nil or zero time, log a short message (e.g. “Could not fetch date for tournament %d: %v”) and fall through to the existing prompt (or skip) behaviour.
- Use the scraped date as-is (or normalise to date-only when persisting if you want; DB already has DATETIME).

### 3. main.go: create and inject client

- In [cmd/elo-cli/main.go](cmd/elo-cli/main.go):
  - Create `meleeClient := melee.NewClient()`.
  - Pass `meleeClient` into `NewProcessor(..., meleeClient, ...)` (add the parameter to the constructor and store it on the processor).

### 4. NewProcessor signature

- Current: `NewProcessor(store, calc, parser, tournamentDates, cfg)`.
- New: `NewProcessor(store, calc, parser, tournamentDates, cfg, meleeClient *melee.Client)` (or equivalent interface). All existing call sites (only main.go) pass the new argument.

### 5. Optional improvements

- **Rate limiting**: When processing many tournament files, add a short delay (e.g. 1–2 seconds) between `FetchTournamentDate` calls so melee.gg is not hammered.
- **User-Agent**: In [internal/melee/client.go](internal/melee/client.go), set a polite `User-Agent` on the request (e.g. `"MeleeEloRanking/1.0 (https://github.com/melee-elo-ranking)"`).
- **Config/flag to disable**: e.g. `-no-scrape` or config `scrape_tournament_dates: false` to skip scraping and go straight to prompt (useful for CI or offline use). If not present, keep behaviour “scraper on when client is non-nil”.

## Files to touch

| File | Change |
|------|--------|
| [cmd/elo-cli/processor.go](cmd/elo-cli/processor.go) | Add `meleeClient *melee.Client` (or interface) to struct and constructor; in date resolution, try `FetchTournamentDate` before prompt when client != nil. |
| [cmd/elo-cli/main.go](cmd/elo-cli/main.go) | Create `melee.NewClient()`, pass to `NewProcessor`. |
| [internal/melee/client.go](internal/melee/client.go) | Optional: set User-Agent; optional: add retries/backoff. |

## Testing

- **Unit**: Processor can be tested with a nil client (prompt path) or a mock that returns a fixed date / error. No change to melee client tests if you don’t change its contract.
- **Manual**: Run with a pending file whose tournament is not in the `-dates` flag and not in the DB; confirm the date is fetched from melee.gg and processing continues without prompt (or that prompt is used when scrape fails / is disabled).

## Summary

- The scraper (URL fetch + date parse) is already implemented in `internal/melee/client.go`.
- This plan only **wires** it into the processor: add an optional melee client to the processor, try `FetchTournamentDate` when the date is missing, and fall back to the existing prompt. Main.go creates the client and passes it into the processor.
