package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type operationResultMsg struct {
	title  string
	err    error
	output string
}

func ripCDCmd(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.CommandContext(ctx, "abcde", "-o", "flac")
		output, err := cmd.CombinedOutput()

		return operationResultMsg{
			title:  "CD rippen",
			err:    err,
			output: strings.TrimSpace(string(output)),
		}
	}
}

func embedCoverCmd(albumPath string) tea.Cmd {
	return func() tea.Msg {
		coverPath := ""
		for _, name := range []string{"cover.jpg", "cover.png", "folder.jpg"} {
			candidate := filepath.Join(albumPath, name)
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
				coverPath = candidate
				break
			}
		}
		if coverPath == "" {
			return operationResultMsg{
				title: "Cover einbetten",
				err:   fmt.Errorf("kein cover.jpg, cover.png oder folder.jpg in %s gefunden", albumPath),
			}
		}

		entries, err := os.ReadDir(albumPath)
		if err != nil {
			return operationResultMsg{title: "Cover einbetten", err: err}
		}

		var flacFiles []string
		for _, entry := range entries {
			if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), ".flac") {
				flacFiles = append(flacFiles, filepath.Join(albumPath, entry.Name()))
			}
		}
		if len(flacFiles) == 0 {
			return operationResultMsg{
				title: "Cover einbetten",
				err:   fmt.Errorf("keine FLAC-Dateien in %s gefunden", albumPath),
			}
		}

		var output strings.Builder
		for _, file := range flacFiles {
			cmd := exec.Command("metaflac", "--import-picture-from="+coverPath, file)
			result, err := cmd.CombinedOutput()
			if len(result) > 0 {
				output.Write(result)
			}
			if err != nil {
				return operationResultMsg{
					title:  "Cover einbetten",
					err:    fmt.Errorf("%s: %w", filepath.Base(file), err),
					output: strings.TrimSpace(output.String()),
				}
			}
		}

		return operationResultMsg{
			title: "Cover eingebettet",
		}
	}
}
