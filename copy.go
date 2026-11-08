package main

import (
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
		cmd := exec.Command("cp", "-r", "--", source, target)
		output, err := cmd.CombinedOutput()
		return copyResultMsg{
			target: target,
			err:    err,
			output: strings.TrimSpace(string(output)),
		}
	}
}
