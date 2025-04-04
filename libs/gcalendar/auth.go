package gcalendar

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

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// GetClient retrieves a token, saves the token, then returns the generated client.
func GetClient(config *oauth2.Config) *http.Client {
	cacheFile, err := TokenCacheFile()
	if err != nil {
		util.LogError(err, "Unable to get path to cached credential file")
		return nil
	}
	tok, err := TokenFromFile(cacheFile)
	if err != nil {
		tok = GetTokenFromWeb(config)
		SaveToken(cacheFile, tok)
	} else {
		util.LogInfo("Using cached token", map[string]interface{}{
			"file": cacheFile,
		})
	}
	return config.Client(context.Background(), tok)
}

// GetTokenFromWeb requests a token from the web, then returns the retrieved token.
func GetTokenFromWeb(config *oauth2.Config) *oauth2.Token {
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
	go OpenURL(authURL)
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

// OpenURL opens a URL in the default browser
func OpenURL(url string) {
	try := []string{"xdg-open", "google-chrome", "open"}
	for _, bin := range try {
		err := exec.Command(bin, url).Run()
		if err == nil {
			return
		}
	}
	util.LogError(nil, "Error opening URL in browser")
}

// TokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func TokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("calendar-api.json")), err
}

// TokenFromFile retrieves a token from a local file.
func TokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// SaveToken saves a token to a file path.
func SaveToken(path string, token *oauth2.Token) {
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

// GetCalendarService creates a new Google Calendar service
func GetCalendarService() (*calendar.Service, error) {
	b, err := os.ReadFile(filepath.Join("configs", "client_secret.json"))
	if err != nil {
		return nil, err
	}

	// If modifying these scopes, delete your previously saved token file
	config, err := google.ConfigFromJSON(b, calendar.CalendarScope)
	if err != nil {
		return nil, err
	}

	client := GetClient(config)
	if client == nil {
		return nil, fmt.Errorf("failed to create HTTP client")
	}

	ctx := context.Background()
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	return srv, nil
}
