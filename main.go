package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

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

func main() {
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

	docFileId := "1oLrOwwqDF7bSLtVM9ls1D3LLHWkdoku80APUMzAzEBM"

	var wg sync.WaitGroup
	wg.Add(2)

	go downloadAndSave(driveService, docFileId, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", &wg)
	go downloadAndSave(driveService, docFileId, "application/pdf", &wg)

	wg.Wait()
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

func downloadAndSave(driveService *drive.Service, fileID string, mimeType string, wg *sync.WaitGroup) {
	defer wg.Done()

	doc, err := driveService.Files.Export(fileID, mimeType).Download()
	if err != nil {
		log.Printf("Unable to retrieve data from document: %v", err)
		return
	}
	defer doc.Body.Close()

	docName, err := driveService.Files.Get(fileID).Do()
	if err != nil {
		log.Printf("Unable to retrieve document: %v", err)
		return
	}

	if doc.StatusCode != http.StatusOK {
		log.Printf("Unable to retrieve data from document: %v", doc.Status)
		return
	}

	file, err := os.Create(docName.Name + "." + getExtension(mimeType))
	if err != nil {
		log.Printf("Unable to save file: %v", err)
		return
	}
	defer file.Close()

	_, err = io.Copy(file, doc.Body)
	if err != nil {
		log.Printf("Unable to save file: %v", err)
		return
	}
}