package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/arunsathiya/download-google-doc/tui"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func main() {
	ui := flag.Bool("ui", false, "Bubble up UI")
	flag.Parse()

	if *ui {
		p := tea.NewProgram(tui.NewModel())

		_, err := p.Run()
		if err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	} else {
		ctx := context.Background()
		b, err := os.ReadFile("credentials.json")
		if err != nil {
			log.Fatalf("Unable to read client secret file: %v", err)
		}

		// If modifying these scopes, delete your previously saved token.json.
		config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/drive.readonly")
		if err != nil {
			log.Fatalf("Unable to parse client secret file to config: %v", err)
		}
		client := tui.GetClient(config)

		driveService, err := drive.NewService(ctx, option.WithHTTPClient(client))
		if err != nil {
			log.Fatalf("Unable to retrieve Drive client: %v", err)
		}

		driveFiles, err := tui.GetDocs(driveService)
		if err != nil {
			log.Fatalf("Unable to retrieve files: %v", err)
		}
		for _, file := range driveFiles.Files {
			fmt.Println(file.Id)
		}
	}
}
