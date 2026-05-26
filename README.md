# Cremio

<img src="cremio.png" width="300" />

A functional way to access Stremio. Cremio is a TUI client built in Go with [Bubbletea](https://github.com/charmbracelet/bubbletea) that talks to Stremio addons, lets you browse catalogs, search for movies and series, pick episodes, and fire streams straight into mpv. No browser, no Electron, no nonsense (_I really didn't want a UI_)

<img width="640" height="385" alt="usage" src="https://github.com/user-attachments/assets/7d8c0eba-7797-4425-97ab-ce8d20cd0555" />


`Disclaimer: This tool neither allows nor encourages streaming and distribution of pirated media. The tool reflects a Proof-of-concept of a Stremio client without any Graphical interface, and the usage in all aspects is subject to the Users' liability and not the tools'.`

## Features

- Browse catalogs from all installed Stremio addons (Note: Cremio doesn't come with any addon to provide Search capabilities or Streams, add your own addons)
- Full-text search with automatic fallback to client-side filtering
- Series support with season/episode navigation 
- Stream resolution across multiple addons
- Playback via mpv (should be in PATH)
- Addon management (add/remove by URL with manifest validation)
- Persistent configuration stored as JSON (in USERPROFILE)

## Prerequisites

- **Go 1.25 or later** (for building from source)
- **mpv** in your system PATH (for stream playback)
- **go-winres** (optional, only needed to embed a custom icon on Windows builds)

## Development Setup

Clone the repository and install dependencies:

```
git clone https://github.com/itssoap/cremio.git
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

On Linux or macOS, omit the `.exe` extension (I haven't tested the build on these systems, will need assistance for this):

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

```json
{
  "addons": [
    "https://example-addon.com/manifest.json"
  ]
}
```

## Watch History

Cremio tracks watched movies and episodes in a local JSON file:

- **Windows:** `%APPDATA%\cremio\history.json`
- **Linux/macOS:** `~/.config/cremio/history.json`

Episodes are automatically marked as watched when played via mpv. You can also manually toggle watched status with the `w` key — on individual episodes, whole seasons, or movies.

The file uses a [Trakt](https://trakt.tv)-compatible structure, so it can be exported and imported directly via Trakt's `/sync/history` API:

```json
{
  "movies": [
    {
      "watched_at": "2026-05-26T12:00:00Z",
      "ids": { "imdb": "tt1234567" }
    }
  ],
  "shows": [
    {
      "ids": { "imdb": "tt7654321" },
      "seasons": [
        {
          "number": 1,
          "episodes": [
            { "number": 1, "watched_at": "2026-05-26T12:00:00Z" },
            { "number": 2, "watched_at": "2026-05-26T12:30:00Z" }
          ]
        }
      ]
    }
  ]
}
```

## Controls

| Key        | Action                                      |
|------------|---------------------------------------------|
| `tab`      | Cycle between Home, Search, and Addons tabs |
| `/`        | Focus the search input (Search tab)         |
| `enter`    | Select item / submit input                  |
| `esc`      | Go back / unfocus input                     |
| `w`        | Toggle watched (episode, season, or movie)  |
| `f`        | Fetch streams for all episodes (series)     |
| `a`        | Add a new addon (Addons tab)                |
| `d`        | Remove selected addon (Addons tab)          |
| `q`       | Quit                                         |
| `ctrl+c`   | Quit                                        |

## Project Structure

```
main.go                  Entry point
internal/
  config/config.go       Configuration loading, saving, addon management
  history/history.go     Watch history tracking and Trakt-compatible export
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

## FAQ/Known issues

Q. Search is broken? I see results showing in Stremio, but not on this app.

A. Please add the Cinemeta Add-on. Some addons lack Search functionality on their catalogues by-default.

Q. Sometimes the Search results appear blank, but it seems like I am able to navigate across them.

A. Please restart the app. This issue occurs very seldomly.

## License

This project is licensed under the [MIT License](LICENSE).
