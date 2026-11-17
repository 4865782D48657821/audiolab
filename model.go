package main

import (
	"context"
	"fmt"

	"charm.land/bubbles/v2/spinner"
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
	operationScreen
)

type model struct {
	screen                       screen
	spinner                      spinner.Model
	cancelOperation              context.CancelFunc
	pendingAction                menuAction
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
	ripCDAction
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
		spinner: spinner.New(),
		items: []menuItem{
			{label: "Musik auf Player kopieren", action: copyMusicAction},
			{label: "CD als FLAC rippen", action: ripCDAction},
			{label: "Cover in FLAC-Dateien einbetten", action: embedCoverAction},
			{label: "Beenden", action: quitAction},
			//{label: "Vorgang läuft...", action: pendingAction},
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
