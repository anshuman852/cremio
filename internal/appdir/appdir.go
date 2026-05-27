package appdir

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Dir returns the cremio config/data directory, creating it if needed.
// On Windows it uses %APPDATA%\cremio; elsewhere ~/.config/cremio.
func Dir() (string, error) {
	var base string

	if runtime.GOOS == "windows" {
		if appData := os.Getenv("APPDATA"); appData != "" {
			base = filepath.Join(appData, "cremio")
		}
	}

	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not determine home directory: %w", err)
		}
		base = filepath.Join(home, ".config", "cremio")
	}

	if err := os.MkdirAll(base, 0o700); err != nil {
		return "", fmt.Errorf("could not create app directory %q: %w", base, err)
	}

	return base, nil
}
