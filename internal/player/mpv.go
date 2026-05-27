package player

import (
	"fmt"
	"os/exec"
)

// PlayWithMPV launches mpv with the given URL and any additional mpv flags.
// The call is non-blocking; mpv runs as a detached child process.
func PlayWithMPV(url string, extraArgs ...string) error {
	if url == "" {
		return fmt.Errorf("no playable URL")
	}

	args := append(extraArgs, url) //nolint:gocritic // intentional: flags before URL
	cmd := exec.Command("mpv", args...)
	return cmd.Start()
}
