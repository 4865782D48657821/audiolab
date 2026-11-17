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
	case operationScreen:
		return m.viewOperation()
	default:
		return m.viewMenu()
	}

}

func (m model) viewOperation() tea.View {
	text := "Vorgang läuft"
	return tea.NewView(m.spinner.View() + text)
}

func (m model) viewMenu() tea.View {
	text := titleStyle.Render("♫ AUDIOLAB") + "\n"
	text += subtitleStyle.Render("Dein Music Manager") + "\n\n"

	for i, item := range m.items {
		if i == m.cursor {
			text += selectedStyle.Render("❯  " + item.label)
		} else {
			text += normalStyle.Render(item.label)
		}
		text += "\n"
	}

	if m.message != "" {
		text += "\n" + errorStyle.Render(m.message)
	}

	text += helpStyle.Render(
		"↑/↓ Navigieren  •  Enter Auswählen  •  q Beenden",
	)

	return tea.NewView(panelStyle.Render(text))
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

func (m model) viewPreview() tea.View {
	content := titleStyle.Render(
		"♫ AUDIOLAB / KOPIERVORSCHAU",
	) + "\n"

	content += labelStyle.Render("QUELLE") + "\n"
	content += valueStyle.Render(m.selectedAlbum) + "\n\n"

	content += labelStyle.Render("ZIEL") + "\n"
	content += valueStyle.Render(m.destination) + "\n\n"

	if m.message != "" {
		content += errorStyle.Render(m.message) + "\n\n"
	} else {
		content += successStyle.Render(
			"✓ Bereit zum Kopieren",
		) + "\n\n"
	}

	content += helpStyle.Render(
		"Enter Kopieren  •  Esc Zurück",
	)

	return tea.NewView(panelStyle.Render(content))
}

func (m model) viewSource() tea.View {
	content := titleStyle.Render("♫ MUSIKQUELLE") + "\n"
	content += subtitleStyle.Render(
		"Wähle den Ordner mit deiner Musiksammlung.",
	) + "\n\n"

	content += labelStyle.Render("QUELLPFAD") + "\n"
	content += m.input.View() + "\n"

	content += helpStyle.Render(
		"Enter Alben anzeigen  •  Esc Zurück",
	)

	return tea.NewView(panelStyle.Render(content))
}

func (m model) viewDestination() tea.View {
	content := titleStyle.Render("♫ ZIELORDNER") + "\n"
	content += subtitleStyle.Render(
		"Wo soll das Album gespeichert werden?",
	) + "\n\n"

	content += labelStyle.Render("ZIELPFAD") + "\n"
	content += m.input.View() + "\n"

	if m.message != "" {
		content += "\n" + errorStyle.Render(m.message)
	}

	content += helpStyle.Render(
		"Enter Prüfen  •  Esc Zurück",
	)

	return tea.NewView(panelStyle.Render(content))
}
