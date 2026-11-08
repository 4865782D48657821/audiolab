package main

import (
	"fmt"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	if result, ok := msg.(copyResultMsg); ok {
		return m.handleCopyResult(result)
	}

	switch m.screen {

	case previewScreen:
		return m.updatePreview(msg)
	case sourceScreen:
		return m.updateSource(msg)
	case destinationScreen:
		return m.updateDestination(msg)
	case albumsScreen:
		return m.updateAlbums(msg)
	default:
		return m.updateMenu(msg)
	}
}

func (m model) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {

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
			switch m.items[m.cursor].action {
			case copyMusicAction:
				m.screen = sourceScreen
				m.message = ""
				m.input.SetValue(m.defaultMusicDir)
				return m, m.input.Focus()
			case quitAction:
				return m, tea.Quit
			case ripCdAction, embedCoverAction:
				// Noch nicht implementiert.
			}
		}
	}

	return m, nil
}

func (m model) updateAlbums(msg tea.Msg) (tea.Model, tea.Cmd) {
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

			albumPath, err := resolveAlbumPath(m.musicDir, m.albums[m.albumCursor])
			if err != nil {
				m.message = err.Error()
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

func (m model) updateDestination(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			m.screen = albumsScreen
			m.input.Blur()
			return m, nil
		case "enter":
			destination, err := validateDestination(m.input.Value())
			if err != nil {
				m.message = err.Error()
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

func (m model) updateSource(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			m.screen = menuScreen
			m.input.Blur()
			return m, nil
		case "enter":
			dir := m.input.Value()

			albums, err := listAlbums(dir)

			if err != nil {
				m.message = fmt.Sprintf("Ordner kann nicht gelesen werden: %v", err)
				return m, nil
			}
			m.musicDir = dir
			m.albums = albums
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

func (m model) updatePreview(msg tea.Msg) (tea.Model, tea.Cmd) {
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

			m.message = "Kopiere …"
			return m, copyAlbumCmd(m.selectedAlbum, target)
		}
	}

	return m, nil
}

func (m model) handleCopyResult(result copyResultMsg) (tea.Model, tea.Cmd) {
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
