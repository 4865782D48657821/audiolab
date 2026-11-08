package main

import (
	"fmt"
	"os"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

type screen int

const (
	menuScreen screen = iota
	sourceScreen
	albumsScreen
	destinationScreen
	previewScreen
)

type model struct {
	screen                       screen
	cursor                       int
	items                        []string
	sourcePath                   string
	message                      string
	input                        textinput.Model
	musicDir                     string
	albums                       []string
	albumCursor                  int
	selectedAlbum                string
	destination                  string
	defaultMusicDir              string
	defaultMountedDeviceMusicDir string
}

func initialModel() model {
	cfg, err := loadConfig()

	input := textinput.New()
	input.Placeholder = "/home/user/Musik"
	input.Prompt = "> "
	input.SetWidth(60)

	m := model{
		items: []string{
			"Musik auf Player kopieren",
			"CD als FLAC rippen",
			"Cover in FLAC-Dateien einbetten",
			"Beenden",
		},
		input:                        input,
		defaultMusicDir:              cfg.DefaultMusicDir,
		defaultMountedDeviceMusicDir: cfg.DefaultMountedDeviceMusicDir,
	}

	if err != nil {
		m.message = fmt.Sprintf("Config-Fehler: %v", err)
	}

	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) View() tea.View {
	if m.screen == sourceScreen {
		return tea.NewView(
			"Pfad zum Music-Ordner eingeben:\n\n" +
				m.input.View() +
				"\n\nEnter: Alben anzeigen · Esc: zurück",
		)
	}

	if m.screen == previewScreen {
		text := "Kopiervorschau\n\n"
		text += "Quelle:\n" + m.selectedAlbum + "\n\n"
		text += "Zielordner:\n" + m.destination + "\n\n"
		text += "Enter: kopieren · Esc: zurück\n"
		if m.message != "" {
			text += "\n" + m.message
		}
		return tea.NewView(text)
	}

	if m.screen == destinationScreen {
		text := "Zielordner auf dem Player eingeben:\n\n"
		text += m.input.View()
		text += "\n\nEnter: prüfen · Esc: zurück"
		if m.message != "" {
			text += "\n\n" + m.message
		}
		return tea.NewView(text)
	}

	if m.screen == albumsScreen {
		text := "Alben in " + m.musicDir + "\n\n"
		for i, album := range m.albums {
			cursor := "  "
			if i == m.albumCursor {
				cursor = "> "
			}
			text += fmt.Sprintf("%s%s\n", cursor, album)
		}
		if m.message != "" {
			text += "\n" + m.message + "\n"
		}
		text += "\n↑/↓ oder j/k: navigieren · Esc: zurück"
		return tea.NewView(text)
	}

	text := "Audiolab\n\n"
	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		text += fmt.Sprintf("%s%s\n", cursor, item)
	}
	if m.message != "" {
		text += "\n" + m.message + "\n"
	}
	text += "\n↑/↓ oder j/k: navigieren · Enter: auswählen · q: beenden"

	return tea.NewView(text)
}

func main() {
	program := tea.NewProgram(initialModel())
	if _, err := program.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Fehler beim Starten:", err)
		os.Exit(1)
	}
}
