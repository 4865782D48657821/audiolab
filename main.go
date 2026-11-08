package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	if result, ok := msg.(copyResultMsg); ok {
		m.screen = previewScreen
		if result.err != nil {
			m.message = fmt.Sprintf("Kopieren fehlgeschlagen: %v", result.err)
			if result.output != "" {
				m.message += "\n" + result.output
			}
		} else {
			m.message = "Kopieren abgeschlossen: " + result.target
		}

		return m, nil
	}

	if m.screen == previewScreen {
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch msg.String() {
			case "esc":
				m.screen = albumsScreen
				return m, nil
			case "ctrl+c":
				return m, tea.Quit
			case "enter":
				target := filepath.Join(
					m.destination,
					filepath.Base(m.selectedAlbum),
				)

				// Nicht in einen bereits vorhandenen Zielordner kopieren.
				if _, err := os.Lstat(target); err == nil {
					m.message = "Abbruch: Der Zielordner existiert bereits: " + target
					return m, nil
				} else if !os.IsNotExist(err) {
					m.message = fmt.Sprintf("Zielordner kann nicht geprüft werden: %v", err)
					return m, nil
				}

				m.message = "Kopiere …"
				return m, copyAlbumCmd(m.selectedAlbum, target)
			}
		}

		return m, nil
	}

	if m.screen == sourceScreen {
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch msg.String() {
			case "esc":
				m.screen = menuScreen
				m.input.Blur()
				return m, nil
			case "enter":
				dir := m.input.Value()
				entries, err := os.ReadDir(dir)
				if err != nil {
					m.message = fmt.Sprintf("Ordner kann nicht gelesen werden: %v", err)
					return m, nil
				}

				m.musicDir = dir
				m.albums = nil
				for _, entry := range entries {
					if entry.IsDir() {
						m.albums = append(m.albums, entry.Name())
					}
				}

				m.albumCursor = 0
				m.screen = albumsScreen
				m.input.Blur()

				if len(m.albums) == 0 {
					m.message = "Keine Album-Unterordner gefunden."
				} else {
					m.message = ""
				}
				return m, nil
			case "ctrl+c":
				return m, tea.Quit
			}
		}

		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	if m.screen == destinationScreen {
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch msg.String() {
			case "esc":
				m.screen = albumsScreen
				m.input.Blur()
				return m, nil
			case "enter":
				destination, err := filepath.Abs(m.input.Value())
				if err != nil {
					m.message = fmt.Sprintf("Zielpfad konnte nicht aufgelöst werden: %v", err)
					return m, nil
				}

				info, err := os.Stat(destination)
				if err != nil {
					m.message = fmt.Sprintf("Zielordner nicht erreichbar: %v", err)
					return m, nil
				}
				if !info.IsDir() {
					m.message = "Der Zielpfad ist kein Ordner."
					return m, nil
				}

				m.destination = destination
				m.message = ""
				m.input.Blur()
				m.screen = previewScreen
				return m, nil
			case "ctrl+c":
				return m, tea.Quit
			}
		}

		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	if m.screen == albumsScreen {
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch msg.String() {
			case "esc":
				m.screen = menuScreen
				return m, nil
			case "up", "k":
				if m.albumCursor > 0 {
					m.albumCursor--
				}
			case "down", "j":
				if m.albumCursor < len(m.albums)-1 {
					m.albumCursor++
				}
			case "ctrl+c":
				return m, tea.Quit

			case "enter":
				if len(m.albums) == 0 {
					m.message = "Es gibt kein Album zum Auswählen."
					return m, nil
				}

				albumPath, err := filepath.Abs(
					filepath.Join(m.musicDir, m.albums[m.albumCursor]),
				)
				if err != nil {
					m.message = fmt.Sprintf("Album-Pfad konnte nicht aufgelöst werden: %v", err)
					return m, nil
				}

				info, err := os.Stat(albumPath)
				if err != nil {
					m.message = fmt.Sprintf("Albumordner nicht erreichbar: %v", err)
					return m, nil
				}
				if !info.IsDir() {
					m.message = "Der ausgewählte Pfad ist kein Ordner."
					return m, nil
				}

				m.selectedAlbum = albumPath
				m.destination = ""
				m.input.Reset()
				m.input.Placeholder = "/mnt/ak100/Music"
				m.input.SetValue(m.defaultMountedDeviceMusicDir)
				m.screen = destinationScreen
				return m, m.input.Focus()
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter":
			switch m.items[m.cursor] {
			case "Musik auf Player kopieren":
				m.screen = sourceScreen
				m.message = ""
				m.input.SetValue(m.defaultMusicDir)
				return m, m.input.Focus()
			case "Beenden":
				return m, tea.Quit
			}
		}
	}

	return m, nil
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
