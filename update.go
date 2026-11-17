package main

import (
	"context"
 "charm.land/bubbles/v2/list"
	"fmt"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

 switch msg := msg.(type) {
 case tea.WindowSizeMsg:
  m.width, m.height = msg.Width, msg.Height
  m.resize()
  return m,nil
 case operationProgressMsg:
  if !m.operation.running || msg.id != m.operation.id { return m,nil }
  m.operation.operationProgress = msg.operationProgress
  return m,waitOperation(m.operation.stream)
 case operationResultMsg:
  if !m.operation.running || msg.id != m.operation.id { return m,nil }
  m.operation.running = false
  m.operation.operationProgress = msg.operationProgress
  m.operation.err, m.operation.output = msg.err, msg.output
  if m.cancelOperation != nil {m.cancelOperation();m.cancelOperation=nil}
  m.details.SetContent(msg.output)
  m.details.GotoTop()
  if m.quitAfterOperation { return m,tea.Quit }
  return m,nil
 }
 if m.screen == operationScreen {
  if key,ok:=msg.(tea.KeyPressMsg);ok {
   switch key.String() {
   case "esc","q","ctrl+c","enter":
    if m.operation.running {
     if key.String()=="enter" {return m,nil}
     if key.String()=="ctrl+c" {m.quitAfterOperation=true}
     m.operation.cancelling=true
     m.cancelOperation()
    } else {
     if key.String()=="ctrl+c" {return m,tea.Quit}
     m.screen=menuScreen
     if m.operation.kind==copyOperation {m.screen=previewScreen}
     m.message=""
    }
    return m,nil
   }
  }
  if m.operation.running {
   var cmd tea.Cmd
   m.spinner,cmd=m.spinner.Update(msg)
   return m,cmd
  }
  var cmd tea.Cmd
  m.details,cmd=m.details.Update(msg)
  return m,cmd
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
				m.pendingAction = copyMusicAction
				m.screen = sourceScreen
				m.message = ""
				m.input.Placeholder="/home/user/Musik"
 m.input.SetValue(m.defaultMusicDir)
				return m, m.input.Focus()

   case ripCDAction:
    return m.startOperation(ripOperation,"CD rippen",ripCD)

			case embedCoverAction:
				m.pendingAction = embedCoverAction
				m.screen = sourceScreen
				m.message = ""
				m.input.Placeholder="/home/user/Musik"
 m.input.SetValue(m.defaultMusicDir)
				return m, m.input.Focus()

			case quitAction:
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m model) updateAlbums(msg tea.Msg) (tea.Model, tea.Cmd) {
 if key,ok:=msg.(tea.KeyPressMsg);ok {
  if key.String()=="ctrl+c" {return m,tea.Quit}
  // Escape first leaves the filter editor/applied filter. Enter first
  // confirms the search; neither key leaks into global navigation.
  if m.albumList.FilterState()!=list.Filtering {
   switch key.String() {
   case "esc":
    if m.albumList.FilterState()==list.Unfiltered {m.screen=menuScreen;return m,nil}
   case "enter":
    item,ok:=m.albumList.SelectedItem().(albumItem)
    if !ok {m.message="Es gibt kein Album zum Auswählen.";return m,nil}
    albumPath,err:=resolveAlbumPath(m.musicDir,string(item))
    if err!=nil {m.message=err.Error();return m,nil}
    m.selectedAlbum=albumPath
    if m.pendingAction==embedCoverAction {
     return m.startOperation(coverOperation,"Cover einbetten",func(ctx context.Context,emit func(operationProgress))(string,error){return embedCover(ctx,albumPath,emit)})
    }
    m.destination="";m.message=""
    m.input.Reset();m.input.Placeholder="/mnt/ak100/Music"
    m.input.SetValue(m.defaultMountedDeviceMusicDir)
    m.screen=destinationScreen
    return m,m.input.Focus()
   }
  }
 }
 var cmd tea.Cmd
 m.albumList,cmd=m.albumList.Update(msg)
 return m,cmd
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
			items := make([]list.Item,len(albums))
 for i,name := range albums {items[i]=albumItem(name)}
 m.albumList.ResetFilter()
 m.albumList.Select(0)
 listCmd := m.albumList.SetItems(items)
			m.screen = albumsScreen
			m.input.Blur()

			if len(m.albums) == 0 {
				m.message = "Keine Album-Unterordner gefunden."
			} else {
				m.message = ""
			}
			return m, listCmd
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

			source := m.selectedAlbum
 return m.startOperation(copyOperation,"Album kopieren",func(ctx context.Context,emit func(operationProgress))(string,error){return "",copyAlbum(ctx,source,target,emit)})
		}
	}

	return m, nil
}

