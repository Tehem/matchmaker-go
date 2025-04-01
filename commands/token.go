package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"matchmaker/util"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		util.LogError(err, "Unable to get path to cached credential file")
		return nil
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	} else {
		util.LogInfo("Using cached token", map[string]interface{}{
			"file": cacheFile,
		})
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	ch := make(chan string)
	randState := fmt.Sprintf("st%d", time.Now().UnixNano())
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/favicon.ico" {
			http.Error(rw, "", 404)
			return
		}
		if req.FormValue("state") != randState {
			util.LogError(nil, "State doesn't match")
			http.Error(rw, "", 500)
			return
		}
		if code := req.FormValue("code"); code != "" {
			fmt.Fprintf(rw, "<h1>Success</h1>Authorized.")
			rw.(http.Flusher).Flush()
			ch <- code
			return
		}
		util.LogError(nil, "No code received")
		http.Error(rw, "", 500)
	}))
	defer ts.Close()

	config.RedirectURL = ts.URL
	authURL := config.AuthCodeURL(randState)
	go openURL(authURL)
	util.LogInfo("Authorize this app", map[string]interface{}{
		"url": authURL,
	})
	code := <-ch
	util.LogInfo("Got authorization code", map[string]interface{}{
		"code": code,
	})

	tok, err := config.Exchange(context.TODO(), code)
	if err != nil {
		util.LogError(err, "Token exchange error")
		return nil
	}
	return tok
}

func openURL(url string) {
	try := []string{"xdg-open", "google-chrome", "open"}
	for _, bin := range try {
		err := exec.Command(bin, url).Run()
		if err == nil {
			return
		}
	}
	util.LogError(nil, "Error opening URL in browser")
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("calendar-api.json")), err
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	util.LogInfo("Saving credential file", map[string]interface{}{
		"path": path,
	})
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		util.LogError(err, "Unable to cache oauth token")
		return
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func init() {
	rootCmd.AddCommand(tokenCmd)
}

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Retrieve a Google Calendar API token.",
	Long:  `Authorize the app to access your Google Agenda and get an auth token for Google Calendar API.`,
	Run: func(cmd *cobra.Command, args []string) {
		b, err := os.ReadFile(filepath.Join("configs", "client_secret.json"))
		if err != nil {
			util.LogError(err, "Unable to read client secret file")
			return
		}

		// If modifying these scopes, delete your previously saved token file
		config, err := google.ConfigFromJSON(b, calendar.CalendarScope)
		if err != nil {
			util.LogError(err, "Unable to parse client secret file to config")
			return
		}
		client := getClient(config)

		srv, err := calendar.New(client)
		if err != nil {
			util.LogError(err, "Unable to retrieve Calendar client")
			return
		}

		t := time.Now().Format(time.RFC3339)
		events, err := srv.Events.List("primary").ShowDeleted(false).
			SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
		if err != nil {
			util.LogError(err, "Unable to retrieve next ten of the user's events")
			return
		}
		util.LogInfo("Upcoming events", nil)
		if len(events.Items) == 0 {
			util.LogInfo("No upcoming events found", nil)
		} else {
			for _, item := range events.Items {
				date := item.Start.DateTime
				if date == "" {
					date = item.Start.Date
				}
				util.LogInfo("Event", map[string]interface{}{
					"summary": item.Summary,
					"date":    date,
				})
			}
		}
	},
}
