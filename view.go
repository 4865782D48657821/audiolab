package main

import (
 "errors"
 "context"
 "fmt"
 "path/filepath"
 "strings"

 tea "charm.land/bubbletea/v2"
)

func (m model) View() tea.View {
 switch m.screen {
 case sourceScreen: return m.viewSource()
 case previewScreen: return m.viewPreview()
 case destinationScreen: return m.viewDestination()
 case albumsScreen: return m.viewAlbums()
 case operationScreen: return m.viewOperation()
 default: return m.viewMenu()
 }
}

func (m model) frame(content string) tea.View {
 width:=max(1,min(90,m.width)-6)
 style:=panelStyle.Width(width)
 if m.width<40 || m.height<12 {
  content=titleStyle.Render("♫ AUDIOLAB")+"\n"+helpStyle.Render("Bitte Terminal vergrößern.")
  if m.operation.running {content+="\n"+m.spinner.View()+" "+safeText(m.operation.title)+"\nEsc Abbrechen"}
  style=panelStyle.Padding(0).Width(max(1,m.width-2))
 }
 return tea.NewView(style.Render(content))
}

func (m model) viewOperation() tea.View {
 op:=m.operation
 content:=titleStyle.Render("♫ AUDIOLAB / "+strings.ToUpper(op.title))+"\n"
 switch {
 case op.running && op.cancelling:
  content+=warningStyle.Render(m.spinner.View()+" Vorgang wird abgebrochen …")
 case op.running:
  label:="Album wird übertragen …"
  if op.kind==coverOperation {label="Cover wird eingebettet …"}
  if op.kind==ripOperation {label="CD-Verarbeitung läuft …"}
  content+=m.spinner.View()+" "+label
 case errors.Is(op.err,context.Canceled):
  content+=warningStyle.Render("■ "+op.title+" abgebrochen.")
 case op.err!=nil:
  content+=errorStyle.Render("✗ "+op.title+" fehlgeschlagen.")
 default:
  content+=successStyle.Render("✓ "+op.title+" abgeschlossen.")
 }
 content+="\n\n"
 if op.file!="" {content+=valueStyle.Render(safeText(op.file))+"\n\n"}
 if op.kind!=ripOperation {
  fraction:=0.0
  if op.total>0 {fraction=min(1,float64(op.current)/float64(op.total))}
  if !op.running&&op.err==nil {fraction=1}
  content+=m.progress.ViewAs(fraction)+"\n"
  if op.kind==copyOperation {content+=subtitleStyle.Render(formatBytes(op.current)+" / "+formatBytes(op.total))} else {
   content+=subtitleStyle.Render(fmt.Sprintf("%d von %d Dateien",op.current,op.total))
  }
  content+="\n\n"
 }
 if op.status!="" {content+=labelStyle.Render("AKTUELLER STATUS")+"\n"+safeText(op.status)+"\n"}
 if op.err!=nil {content+="\n"+errorStyle.Render(safeText(op.err.Error()))+"\n"}
 if !op.running&&op.output!="" {content+="\n"+m.details.View()+"\n"}
 if op.running {
  content+=helpStyle.Render("Esc/q Abbrechen  •  Ctrl+C Abbrechen und beenden")
  if op.kind==ripOperation {content+="\n"+subtitleStyle.Render("abcde ohne Terminal-Eingabe; interaktive Konfiguration benötigt Anpassung.")}
 } else {
  content+=helpStyle.Render("Enter/Esc Zurück  •  ↑/↓ Ausgabe scrollen  •  Ctrl+C Beenden")
 }
 return m.frame(content)
}

func formatBytes(n int64)string {
 if n<1024 {return fmt.Sprintf("%d B",n)}
 value:=float64(n)
 for _,unit:=range []string{"KiB","MiB","GiB","TiB"} {
  value/=1024
  if value<1024||unit=="TiB" {return fmt.Sprintf("%.1f %s",value,unit)}
 }
 return ""
}

func (m model) viewMenu() tea.View {
 content:=titleStyle.Render("♫ AUDIOLAB")+"\n"+subtitleStyle.Render("Dein Music Manager")+"\n\n"
 for i,item:=range m.items {
  if i==m.cursor {content+=selectedStyle.Render("❯  "+item.label)} else {content+=normalStyle.Render(item.label)}
  content+="\n"
 }
 if m.message!="" {content+="\n"+errorStyle.Render(safeText(m.message))+"\n"}
 content+=helpStyle.Render("↑/↓ oder j/k Navigieren  •  Enter Auswählen  •  q Beenden")
 return m.frame(content)
}
func (m model) viewAlbums() tea.View {
 content:=titleStyle.Render("♫ AUDIOLAB / ALBEN")+"\n"+subtitleStyle.Render(safeText(m.musicDir))+"\n\n"+m.albumList.View()
 if m.message!="" {content+="\n"+warningStyle.Render(safeText(m.message))}
 help:="↑/↓ oder j/k Navigieren  •  / Suchen  •  Enter Auswählen  •  Esc Zurück"
 if m.albumList.SettingFilter() {help="Suchtext eingeben  •  Enter Filter anwenden  •  Esc Filter abbrechen"} else if m.albumList.IsFiltered() {help="↑/↓ oder j/k Navigieren  •  Enter Auswählen  •  Esc Filter löschen"}
 content+="\n"+helpStyle.Render(help)
 return m.frame(content)
}
func (m model) viewPreview() tea.View {
 content:=titleStyle.Render("♫ AUDIOLAB / KOPIERVORSCHAU")+"\n"
 content+=labelStyle.Render("QUELLE")+"\n"+valueStyle.Render(safeText(m.selectedAlbum))+"\n\n"
 content+=labelStyle.Render("ZIEL")+"\n"+valueStyle.Render(safeText(filepath.Join(m.destination,filepath.Base(m.selectedAlbum))))+"\n\n"
 if m.message!="" {content+=errorStyle.Render(safeText(m.message))+"\n\n"} else {content+=successStyle.Render("✓ Bereit zum Kopieren")+"\n\n"}
 content+=helpStyle.Render("Enter Kopieren  •  Esc Zurück  •  Ctrl+C Beenden")
 return m.frame(content)
}
func (m model) viewSource() tea.View {
 content:=titleStyle.Render("♫ AUDIOLAB / MUSIKQUELLE")+"\n"+subtitleStyle.Render("Wähle den Ordner mit deiner Musiksammlung.")+"\n\n"
 content+=labelStyle.Render("QUELLPFAD")+"\n"+m.input.View()+"\n"
 if m.message!="" {content+="\n"+errorStyle.Render(safeText(m.message))+"\n"}
 content+=helpStyle.Render("Enter Alben anzeigen  •  Esc Zurück  •  Ctrl+C Beenden")
 return m.frame(content)
}
func (m model) viewDestination() tea.View {
 content:=titleStyle.Render("♫ AUDIOLAB / ZIELORDNER")+"\n"+subtitleStyle.Render("Wo soll das Album gespeichert werden?")+"\n\n"
 content+=labelStyle.Render("ZIELPFAD")+"\n"+m.input.View()+"\n"
 if m.message!="" {content+="\n"+errorStyle.Render(safeText(m.message))+"\n"}
 content+=helpStyle.Render("Enter Prüfen  •  Esc Zurück  •  Ctrl+C Beenden")
 return m.frame(content)
}

