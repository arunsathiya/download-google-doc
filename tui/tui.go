package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/arunsathiya/download-google-doc/tui/keys"
	"github.com/arunsathiya/download-google-doc/tui/styles"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

const (
	defaultWidth = 20
	listHeight   = 16
)

type item struct {
	name string
	id   string
}

func (i item) FilterValue() string { return i.name }

type state int

const (
	browsing state = iota
)

type Model struct {
	keyMap *keys.KeyMap
	list   list.Model
	styles styles.Styles
	state  state
}

func NewModel() Model {
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
	client := getClient(config)

	driveService, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	driveFiles, err := getDocs(driveService)
	if err != nil {
		log.Fatalf("Unable to retrieve files: %v", err)
	}

	items := []list.Item{}
	for _, file := range driveFiles.Files {
		if file.MimeType == "application/vnd.google-apps.document" {
			items = append(items, item{
				name: file.Name,
				id:   file.Id,
			})
		}
	}

	styles := styles.DefaultStyles()
	keys := keys.NewKeyMap()

	l := list.New(items, newItemDelegate(keys, &styles), defaultWidth, listHeight)
	l.Title = "Your docs"
	l.SetShowStatusBar(true)
	l.Styles.PaginationStyle = styles.Pagination
	l.Styles.HelpStyle = styles.Help

	return Model{
		keyMap: keys,
		list:   l,
		styles: styles,
		state:  browsing,
	}
}

func (m *Model) updateKeybindings() {
	if m.list.SettingFilter() {
		m.keyMap.Enter.SetEnabled(false)
	}

	switch m.state {
	case browsing:
		m.keyMap.Enter.SetEnabled(true)
		m.keyMap.Cancel.SetEnabled(false)

	default:
		m.keyMap.Enter.SetEnabled(true)
		m.keyMap.Cancel.SetEnabled(false)
	}
}

func listUpdate(msg tea.Msg, m Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.list.KeyMap.AcceptWhileFiltering):
			m.state = browsing
			m.updateKeybindings()

		case key.Matches(msg, m.keyMap.CursorUp):
			m.list.CursorUp()

		case key.Matches(msg, m.keyMap.CursorDown):
			m.list.CursorDown()

		case key.Matches(msg, m.keyMap.Enter):
			if i, ok := m.list.SelectedItem().(item); ok {
				var wg sync.WaitGroup
				wg.Add(2)

				go downloadAndSave(i.id, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", &wg)
				go downloadAndSave(i.id, "application/pdf", &wg)

				wg.Wait()
				return m, tea.Quit
			}
		}
	}

	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.list.SettingFilter() {
		m.keyMap.Enter.SetEnabled(false)
	}

	switch m.state {
	case browsing:
		return listUpdate(msg, m)

	default:
		return m, nil
	}
}

func (m Model) View() string {
	switch m.state {
	case browsing:
		return "\n" + m.list.View()

	default:
		return ""
	}
}

// Retrieves a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Requests a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	defer f.Close()
	if err != nil {
		log.Fatalf("Unable to cache OAuth token: %v", err)
	}
	json.NewEncoder(f).Encode(token)
}

func getDocs(driveService *drive.Service) (*drive.FileList, error) {
	driveFiles, err := driveService.Files.List().Do()
	if err != nil {
		log.Fatalf("Unable to retrieve files: %v", err)
		return nil, err
	}
	return driveFiles, nil
}

func getExtension(mimeType string) string {
	switch mimeType {
	case "application/pdf":
		return "pdf"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return "docx"
	default:
		return "unknown"
	}
}

func downloadAndSave(fileID string, mimeType string, wg *sync.WaitGroup) {
	defer wg.Done()

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
	client := getClient(config)

	driveService, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	docResponse, err := driveService.Files.Export(fileID, mimeType).Download()
	if err != nil {
		log.Printf("Unable to retrieve data from document: %v", err)
		return
	}
	defer docResponse.Body.Close()

	doc, err := driveService.Files.Get(fileID).Do()
	if err != nil {
		log.Printf("Unable to retrieve document: %v", err)
		return
	}

	if docResponse.StatusCode != http.StatusOK {
		log.Printf("Unable to retrieve data from document: %v", docResponse.Status)
		return
	}

	file, err := os.Create(filepath.Join("exports", doc.Name+"."+getExtension(mimeType)))
	if err != nil {
		log.Printf("Unable to save file: %v", err)
		return
	}
	defer file.Close()

	_, err = io.Copy(file, docResponse.Body)
	if err != nil {
		log.Printf("Unable to save file: %v", err)
		return
	}
}
