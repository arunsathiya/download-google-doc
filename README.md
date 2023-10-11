# 💾 Download Google Doc

Go program to download a Google document in specific file types. At this time, the program fetches the list of documents in an authenticated Google Drive allows downloading the doc file in ".docx" and ".pdf" formats.

<img width="726" alt="image" src="https://github.com/arunsathiya/download-google-doc/assets/18581859/7211a5a3-e1a3-4da2-a513-e7567e39f2a1">

### How to run

(Work in progress)

At this time, the program runs with a custom Google Cloud Oauth2 client with Google Drive and Google Docs API enabled. If you are keen to set up right away:

- Create a [Google Cloud project](https://console.cloud.google.com/apis/dashboard) and enable Drive API and Docs API.
- [Create a Oauth2 client for your project](https://console.cloud.google.com/apis/credentials) and [add your email address as a test user](https://console.cloud.google.com/apis/credentials/consent).
- Clone repository.
- Download the `client_secret_xxx.json` file into the cloned folder, and rename it to `credentials.json`.
- Run `go run main.go`
- Authorize your Google Cloud app with your Google account. Copy redirect URL, extract just the token and paste on the terminal screen.
- `token.json` file should now be created on your folder, which will be used for all API queries. Don't delete this file.

### Plans for the future

- Support downloading in specific formats.
- Support bulk downloading of documents.
- [Submit ideas here!](https://github.com/arunsathiya/download-google-doc/issues)
