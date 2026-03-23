# aptscout — Apartment Availability Tracker

![aptscout](./assets/og-image.svg)

[![CI](https://github.com/dotbrains/aptscout/actions/workflows/ci.yml/badge.svg)](https://github.com/dotbrains/aptscout/actions/workflows/ci.yml)
[![Release](https://github.com/dotbrains/aptscout/actions/workflows/release.yml/badge.svg)](https://github.com/dotbrains/aptscout/actions/workflows/release.yml)
[![License: PolyForm Shield 1.0.0](https://img.shields.io/badge/License-PolyForm%20Shield%201.0.0-blue.svg)](https://polyformproject.org/licenses/shield/1.0.0/)

![Go](https://img.shields.io/badge/-Go-00ADD8?style=flat-square&logo=go&logoColor=white)
![Cobra](https://img.shields.io/badge/-Cobra-00ADD8?style=flat-square&logo=go&logoColor=white)
![SQLite](https://img.shields.io/badge/-SQLite-003B57?style=flat-square&logo=sqlite&logoColor=white)
![macOS](https://img.shields.io/badge/-macOS-000000?style=flat-square&logo=apple&logoColor=white)
![Linux](https://img.shields.io/badge/-Linux-FCC624?style=flat-square&logo=linux&logoColor=black)

Scrape apartment availability across multiple properties, store results in SQLite, and browse with a filterable web UI. Track prices, catch deals, never miss a unit.

## Quick Start

```sh
# Install
go install github.com/dotbrains/aptscout@latest

# Scrape all properties
aptscout scrape

# Scrape a single property
aptscout scrape --property desert-club

# List 2-bed apartments under $2,500
aptscout list --beds 2 --max-price 2500

# Filter by property
aptscout list --property hideaway --beds 2

# Check price history for a unit
aptscout history 2146

# View summary stats
aptscout stats

# Browse in the web UI
aptscout serve --open

# List registered properties
aptscout properties
```

## How It Works

1. `aptscout scrape` runs each registered **provider** — a self-contained scraper for a specific apartment complex.
2. Each provider fetches its property's website, parses floor plans and available units, and returns standardized data.
3. Results are stored in a local **SQLite** database with composite keys `(property, unit_number)` so multiple properties coexist.
4. Every price change is recorded in a **price history** table for tracking trends over time.
5. Units that disappear between scrapes are marked unavailable — you can see when listings come and go.

Currently supported properties:

| Property | Provider | Site |
|---|---|---|
| `desert-club` | Desert Club Apartments (Weidner) | [arizona.weidner.com](https://arizona.weidner.com/apartments/az/phoenix/desert-club0/floorplans) |
| `hideaway` | Hideaway North Scottsdale (Mark-Taylor) | [hideawaynorthscottsdale.com](https://www.hideawaynorthscottsdale.com/floorplans) |

## Installation

### Via `go install`

```sh
go install github.com/dotbrains/aptscout@latest
```

### Via Homebrew

```sh
brew tap dotbrains/tap
brew install --cask aptscout
```

### Via GitHub Release

```sh
gh release download --repo dotbrains/aptscout --pattern 'aptscout_darwin_arm64.tar.gz' --dir /tmp
tar -xzf /tmp/aptscout_darwin_arm64.tar.gz -C /usr/local/bin
```

### From source

```sh
git clone https://github.com/dotbrains/aptscout.git
cd aptscout
make install
```

## Commands

| Command | Description |
|---|---|
| `aptscout scrape` | Scrape all properties (or `--property <id>` for one) |
| `aptscout list` | List available apartments with filters |
| `aptscout history <unit>` | Show price history for a unit |
| `aptscout stats` | Show summary statistics |
| `aptscout serve` | Start a local web UI to browse apartments |
| `aptscout properties` | List registered apartment properties |
| `aptscout clean` | Remove stale apartment data |

### Global Flags

| Flag | Description |
|---|---|
| `--property <id>` | Filter by property (e.g. `desert-club`, `hideaway`) |
| `--db <path>` | Override database path |
| `--version` | Print version |

### List Flags

| Flag | Description |
|---|---|
| `--beds <n>` | Filter by bedroom count |
| `--baths <n>` | Filter by bathroom count |
| `--max-price <n>` | Maximum monthly rent |
| `--min-price <n>` | Minimum monthly rent |
| `--plan <code>` | Filter by floor plan code |
| `--renovated` | Only renovated units |
| `--available-by <date>` | Available by date (YYYY-MM-DD) |
| `--sort <field>` | Sort by: `price`, `date`, `sqft`, `unit` |
| `--json` | Output as JSON |

## Web UI

Browse apartments in a local web interface:

```sh
# Start the apartment browser
aptscout serve

# Custom port + auto-open browser
aptscout serve --port 9000 --open
```

All assets are embedded in the binary — no external dependencies. The UI uses the same desert theme as the marketing site.

Features:
- **Property picker** — home page shows each property with live stats; click to browse its units
- **Filter sidebar** — bedrooms, bathrooms, price range, date range, floor plan, renovated/premium
- **Date range filter** — select a from/to date to find units available in that window; contextual empty state when no matches
- **Apartment cards** — unit number, plan, specs, price, availability, property badge
- **Unit detail** — full metadata, price history SVG chart, direct link to property site
- **⌘K command palette** — quick-nav to any property, page, or action; fuzzy search, arrow keys, Enter
- **Re-scrape from UI** — trigger a scrape without leaving the browser
- **Keyboard shortcuts** — `⌘K` command palette, `/` to focus filters, `r` to scrape, `Escape` to clear

## Adding a New Property

Each property is a Go package implementing the `models.Provider` interface:

```go
type Provider interface {
    ID() string       // "my-property"
    Name() string     // "My Property Apartments"
    Scrape(ctx context.Context, fetch Fetcher) (*ScrapeData, error)
}
```

Steps:

1. Create `internal/provider/myproperty/myproperty.go` implementing the interface.
2. Add one line to `internal/provider/registry.go`:
   ```go
   "my-property": myproperty.New(),
   ```
3. That's it. `aptscout scrape` will include the new property automatically.

The `Fetcher` function is provided by the scraper — your provider just calls `fetch(ctx, url)` and parses the HTML. No need to manage HTTP clients, retries, or rate limiting.

## Data Storage

SQLite database at `~/.local/share/aptscout/aptscout.db`. Created automatically on first run.

- **`floor_plans`** — plan metadata per property (beds, baths, sqft, deposit)
- **`apartments`** — individual units with availability and pricing
- **`price_history`** — every price change recorded with timestamps
- **`scrape_runs`** — log of each scrape execution with stats

All tables are keyed by `(property, ...)` so data from different complexes never collides.

## Dependencies

- **Go 1.22+**
- **[Cobra](https://github.com/spf13/cobra)** — CLI framework
- **[golang.org/x/net/html](https://pkg.go.dev/golang.org/x/net/html)** — HTML parser
- **[modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite)** — Pure-Go SQLite (no CGO)

No external JavaScript. No Node.js. No Playwright. Single static binary.

## License

This project is licensed under the [PolyForm Shield License 1.0.0](https://polyformproject.org/licenses/shield/1.0.0/) — see [LICENSE](LICENSE) for details.