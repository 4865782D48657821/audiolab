package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type copyResultMsg struct {
	target string
	err    error
	output string
}

func copyAlbumCmd(source, target string) tea.Cmd {
	return func() tea.Msg {
		if _, err := os.Lstat(target); err == nil {
			return copyResultMsg{
				target: target,
				err:    fmt.Errorf("der Zielordner existiert bereits: %s", target),
			}
		} else if !os.IsNotExist(err) {
			return copyResultMsg{
				target: target,
				err:    fmt.Errorf("Zielordner kann nicht geprüft werden: %w", err),
			}
		}

		cmd := exec.Command("cp", "-r", "--", source, target)
		output, err := cmd.CombinedOutput()

		return copyResultMsg{
			target: target,
			err:    err,
			output: strings.TrimSpace(string(output)),
		}
	}
}
