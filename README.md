# Cremio

Stremio belongs in the terminal. Cremio is a TUI client built in Go with Bubbletea that talks to Stremio addons, lets you browse catalogs, search for movies and series, pick episodes, and fire streams straight into mpv. No browser, no Electron, no nonsense.

## Features

- Browse catalogs from all installed Stremio addons
- Full-text search with automatic fallback to client-side filtering
- Series support with season/episode navigation
- Stream resolution across multiple addons
- Playback via mpv (any stream URL, magnet link, or YouTube ID)
- Addon management (add/remove by URL with manifest validation)
- Persistent configuration stored as JSON
- Animated loading indicators on all async operations

## Prerequisites

- **Go 1.25 or later** (for building from source)
- **mpv** in your system PATH (for stream playback)
- **go-winres** (optional, only needed to embed a custom icon on Windows builds)

## Development Setup

Clone the repository and install dependencies:

```
git clone https://github.com/soakhan/cremio.git
cd cremio
go mod tidy
```

Run the application directly:

```
go run .
```

## Compilation

Build a standalone binary:

```
go build -o cremio.exe .
```

On Linux or macOS, omit the `.exe` extension:

```
go build -o cremio .
```

### Windows Icon Embedding

To embed a custom icon into the Windows executable, install go-winres and regenerate the resource files before building:

```
go install github.com/tc-hib/go-winres@latest
```

Place your icon as a PNG in `winres/icon.png` (max 256x256) and `winres/icon16.png` (16x16), then:

```
go-winres make
go build -o cremio.exe .
```

The generated `.syso` files are excluded from version control via `.gitignore` and must be regenerated locally.

## Configuration

Cremio stores its configuration as a JSON file:

- **Windows:** `%APPDATA%\cremio\config.json`
- **Linux/macOS:** `~/.config/cremio/config.json`

The config file holds the list of installed addon base URLs. It is created automatically on first use.

## Controls

| Key        | Action                                      |
|------------|---------------------------------------------|
| `tab`      | Cycle between Home, Search, and Addons tabs |
| `/`        | Focus the search input (Search tab)         |
| `enter`    | Select item / submit input                  |
| `esc`      | Go back / unfocus input                     |
| `a`        | Add a new addon (Addons tab)                |
| `d`        | Remove selected addon (Addons tab)          |
| `q`       | Quit                                         |
| `ctrl+c`   | Quit                                        |

## Project Structure

```
main.go                  Entry point
internal/
  config/config.go       Configuration loading, saving, addon management
  player/mpv.go          mpv process launcher
  stremio/
    client.go            HTTP client for the Stremio Addon Protocol
    types.go             Manifest, catalog, meta, stream type definitions
  tui/
    app.go               Root Bubbletea model and screen routing
    home.go              Home screen with catalog browsing
    search.go            Search screen with addon querying
    addons.go            Addon management screen
    detail.go            Movie/series detail and episode list
    streams.go           Stream list and mpv launch
    styles.go            Lipgloss style definitions
winres/
  winres.json            Windows resource manifest for go-winres
  icon.png               Application icon (256x256)
  icon16.png             Application icon (16x16)
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Keep changes focused and minimal
4. Submit a pull request

Please avoid introducing new dependencies unless strictly necessary. Keep the codebase simple and the TUI responsive.

## License

This project is provided as-is without a specific license. Contact the author for usage terms.
