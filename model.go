package main

import (
	"fmt"

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
	items                        []menuItem
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

type menuAction int

const (
	copyMusicAction menuAction = iota
	ripCdAction
	embedCoverAction
	quitAction
)

type menuItem struct {
	label  string
	action menuAction
}

func initialModel() model {
	cfg, err := loadConfig()

	input := textinput.New()
	input.Placeholder = "/home/user/Musik"
	input.Prompt = "> "
	input.SetWidth(60)

	m := model{
		items: []menuItem{
			{label: "Musik auf Player kopieren", action: copyMusicAction},
			{label: "CD als FLAC rippen", action: ripCdAction},
			{label: "Cover in FLAC-Dateien einbetten", action: embedCoverAction},
			{label: "Beenden", action: quitAction},
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
