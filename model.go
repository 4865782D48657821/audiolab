package main

import (
	"context"
	"fmt"

	"charm.land/bubbles/v2/spinner"
 "charm.land/bubbles/v2/list"
 "charm.land/bubbles/v2/progress"
 "charm.land/bubbles/v2/viewport"
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
	albumList                    list.Model
 width, height int
 progress progress.Model
 details viewport.Model
 operation operationState
 runtime *operationRuntime
 nextOperationID uint64
 quitAfterOperation bool
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
 inputStyles := textinput.DefaultStyles(true)
 inputStyles.Focused.Prompt = labelStyle
 inputStyles.Focused.Text = valueStyle.PaddingLeft(0)
 input.SetStyles(inputStyles)
 delegate := list.NewDefaultDelegate()
 delegate.ShowDescription = false
 delegate.SetSpacing(0)
 delegate.Styles.SelectedTitle = selectedStyle.BorderLeft(true).BorderForeground(primary)
 albums := list.New(nil, delegate, 64, 12)
 albums.Title = "Alben auswählen"
 albums.SetShowHelp(false)
 albums.SetStatusBarItemName("Album", "Alben")
 albums.DisableQuitKeybindings()
 albums.Styles.Title = titleStyle
 albums.FilterInput.SetStyles(inputStyles)

	m := model{
		spinner: spinner.New(spinner.WithSpinner(spinner.Dot), spinner.WithStyle(labelStyle)),
 albumList: albums,
 progress: progress.New(progress.WithColors(primary, accent), progress.WithFillCharacters('█', '░')),
 details: viewport.New(viewport.WithWidth(64), viewport.WithHeight(6)),
 runtime: newOperationRuntime(),
 width: 72, height: 24,
		items: []menuItem{
			{label: "Musik auf Player kopieren", action: copyMusicAction},
			{label: "CD als FLAC rippen", action: ripCDAction},
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

 m.resize()
 return m
}

func (m model) Init() tea.Cmd {
	return nil
}


type albumItem string
func (a albumItem) Title() string { return safeText(string(a)) }
func (a albumItem) Description() string { return "" }
func (a albumItem) FilterValue() string { return string(a) }

func (m *model) resize() {
 w := max(1,min(90,m.width)-6)
 m.input.SetWidth(max(1,w-3))
 m.albumList.SetSize(w,max(4,m.height-10))
 m.progress.SetWidth(w)
 m.details.SetWidth(w)
 m.details.SetHeight(max(1,min(8,m.height-16)))
 m.details.SoftWrap = true
}
