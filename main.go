package main

import (
 "fmt"
 "os"

 tea "charm.land/bubbletea/v2"
)

func run() error {
 m:=initialModel()
 defer m.runtime.close()
 program:=tea.NewProgram(m)
 _,err:=program.Run()
 return err
}
func main() {
 if err:=run();err!=nil {
  fmt.Fprintln(os.Stderr,"Fehler beim Starten:",err)
  os.Exit(1)
 }
}
