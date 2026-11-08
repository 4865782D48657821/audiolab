package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

func main() {
	program := tea.NewProgram(initialModel())
	if _, err := program.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Fehler beim Starten:", err)
		os.Exit(1)
	}
}
