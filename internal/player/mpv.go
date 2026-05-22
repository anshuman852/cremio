package player

import (
	"fmt"
	"os/exec"
)

func PlayWithMPV(url string) error {
	if url == "" {
		return fmt.Errorf("no playable URL")
	}

	cmd := exec.Command("mpv", url)
	return cmd.Start()
}
