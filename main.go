package main

import (
	"fmt"
	"os"

	"github.com/arunsathiya/download-google-doc/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(tui.NewModel())

	err := p.Start()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
