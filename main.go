package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

type model struct {
	cursor int
	items  []string
}

func initialModel() model {
	return model{
		items: []string{
			"Musik auf Player kopieren",
			"CD als FLAC rippen",
			"Cover in FLAC-Dateien einbetten",
			"Beenden",
		},
	}
}
func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.items[m.cursor] == "Beenden" {
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m model) View() tea.View {
	text := "Audiolab\n\n"
	for i, item := range m.items {
		cursor := " "
		if i == m.cursor {
			cursor = "> "
		}
		text += fmt.Sprintf("%s%s\n", cursor, item)
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
