package main

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

func (m model) View() tea.View {
	switch m.screen {

	case sourceScreen:
		return m.viewSource()
	case previewScreen:
		return m.viewPreview()
	case destinationScreen:
		return m.viewDestination()
	case albumsScreen:
		return m.viewAlbums()
	default:
		return m.viewMenu()
	}

}

func (m model) viewMenu() tea.View {
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

func (m model) viewAlbums() tea.View {
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

func (m model) viewDestination() tea.View {
	text := "Zielordner auf dem Player eingeben:\n\n"
	text += m.input.View()
	text += "\n\nEnter: prüfen · Esc: zurück"
	if m.message != "" {
		text += "\n\n" + m.message
	}
	return tea.NewView(text)
}

func (m model) viewPreview() tea.View {
	text := "Kopiervorschau\n\n"
	text += "Quelle:\n" + m.selectedAlbum + "\n\n"
	text += "Zielordner:\n" + m.destination + "\n\n"
	text += "Enter: kopieren · Esc: zurück\n"
	if m.message != "" {
		text += "\n" + m.message
	}
	return tea.NewView(text)
}

func (m model) viewSource() tea.View {
	return tea.NewView(
		"Pfad zum Music-Ordner eingeben:\n\n" +
			m.input.View() +
			"\n\nEnter: Alben anzeigen · Esc: zurück",
	)
}
