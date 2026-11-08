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
)

type model struct {
	screen      screen
	cursor      int
	items       []string
	sourcePath  string
	message     string
	input       textinput.Model
	musicDir    string
	albums      []string
	albumCursor int
}

func initialModel() model {
	input := textinput.New()
	input.Placeholder = "/home/user/Musik/Album"
	input.Prompt = "> "
	input.SetWidth(60)

	return model{
		items: []string{
			"Musik auf Player kopieren",
			"CD als FLAC rippen",
			"Cover in FLAC-Dateien einbetten",
			"Beenden",
		},
		input: input,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

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
				m.input.Reset()
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
