package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/arunsathiya/download-google-doc/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	ui := flag.Bool("ui", false, "Bubble up UI")
	doc := flag.String("doc", "", "Doc ID to download")
	flag.Parse()

	if *ui {
		p := tea.NewProgram(tui.NewModel())
		_, err := p.Run()
		if err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	} else {
		if *doc == "" {
			flag.PrintDefaults()
			os.Exit(1)
		}
		docId := strings.Split(strings.Split(*doc, "/edit")[0], "d/")[1]
		var wg sync.WaitGroup
		wg.Add(2)
		go tui.DownloadAndSave(docId, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", &wg)
		go tui.DownloadAndSave(docId, "application/pdf", &wg)
		wg.Wait()
	}
}
