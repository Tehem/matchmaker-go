package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"time"

	"matchmaker/internal/fs"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// TokenCmd represents the token command
var TokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Get a Google Calendar API token",
	Long: `This command will help you get a Google Calendar API token. It will first check for an existing token.
If no valid token is found, it will open a browser window where you can authorize the application.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get user's home directory
		usr, err := user.Current()
		if err != nil {
			return fmt.Errorf("unable to get user's home directory: %w", err)
		}

		// Create credentials directory if it doesn't exist
		credentialsDir := filepath.Join(usr.HomeDir, ".credentials")
		if err := os.MkdirAll(credentialsDir, 0700); err != nil {
			return fmt.Errorf("unable to create credentials directory: %w", err)
		}

		// Check for existing token
		tokenFile := filepath.Join(credentialsDir, "calendar-api.json")
		token, err := loadToken(tokenFile)
		if err == nil {
			// Token exists, verify it's still valid
			if err := verifyCalendarAccess(token); err == nil {
				fmt.Println("Using existing token")
				return nil
			}
			// Token is invalid, we'll get a new one
			fmt.Println("Existing token is invalid, getting a new one...")
		}

		// Read client secret file
		b, err := os.ReadFile("client_secret.json")
		if err != nil {
			return fmt.Errorf("unable to read client secret file: %w", err)
		}

		// Configure OAuth2 config
		config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/calendar")
		if err != nil {
			return fmt.Errorf("unable to parse client secret file to config: %w", err)
		}

		// Get token from web
		token, err = getTokenFromWeb(config)
		if err != nil {
			return fmt.Errorf("unable to get token from web: %w", err)
		}

		// Save token to file
		if err := saveToken(tokenFile, token); err != nil {
			return fmt.Errorf("unable to save token: %w", err)
		}

		fmt.Printf("Token saved to %s\n", tokenFile)

		// Verify calendar access by displaying next 10 events
		if err := verifyCalendarAccess(token); err != nil {
			return fmt.Errorf("failed to verify calendar access: %w", err)
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(TokenCmd)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	ch := make(chan string)
	randState := fmt.Sprintf("st%d", time.Now().UnixNano())
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/favicon.ico" {
			http.Error(rw, "", 404)
			return
		}
		if req.FormValue("state") != randState {
			fmt.Printf("State doesn't match: req = %#v", req)
			http.Error(rw, "", 500)
			return
		}
		if code := req.FormValue("code"); code != "" {
			fmt.Fprintf(rw, "<h1>Success</h1>Authorized.")
			rw.(http.Flusher).Flush()
			ch <- code
			return
		}
		fmt.Println("no code")
		http.Error(rw, "", 500)
	}))
	defer ts.Close()

	config.RedirectURL = ts.URL
	authURL := config.AuthCodeURL(randState)
	go openURL(authURL)
	fmt.Printf("Authorize this app at: %s", authURL)
	code := <-ch
	fmt.Printf("Got code: %s", code)

	tok, err := config.Exchange(context.TODO(), code)
	if err != nil {
		return nil, fmt.Errorf("token exchange error: %v", err)
	}
	return tok, nil
}

func openURL(url string) {
	try := []string{"xdg-open", "google-chrome", "open"}
	for _, bin := range try {
		err := exec.Command(bin, url).Run()
		if err == nil {
			return
		}
	}
	fmt.Println("Error opening URL in browser.")
}

// loadToken loads the OAuth2 token from a file
func loadToken(filepath string) (*oauth2.Token, error) {
	data, err := fs.Default.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("failed to decode token: %w", err)
	}

	return &token, nil
}

// saveToken saves the OAuth2 token to a file
func saveToken(filepath string, token *oauth2.Token) error {
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to encode token: %w", err)
	}

	if err := fs.Default.WriteFile(filepath, data, 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}

// verifyCalendarAccess verifies the calendar access by displaying the next 10 events
func verifyCalendarAccess(token *oauth2.Token) error {
	ctx := context.Background()
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("unable to create calendar service: %w", err)
	}

	// Get the user's primary calendar
	calendar, err := srv.Calendars.Get("primary").Do()
	if err != nil {
		return fmt.Errorf("unable to get primary calendar: %w", err)
	}

	// Get the next 10 events
	timeMin := time.Now().Format(time.RFC3339)
	events, err := srv.Events.List(calendar.Id).
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(timeMin).
		MaxResults(10).
		OrderBy("startTime").
		Do()
	if err != nil {
		return fmt.Errorf("unable to retrieve events: %w", err)
	}

	fmt.Println("\nUpcoming events:")
	if len(events.Items) == 0 {
		fmt.Println("No upcoming events found.")
		return nil
	}

	for _, item := range events.Items {
		date := item.Start.DateTime
		if date == "" {
			date = item.Start.Date
		}
		fmt.Printf("%v (%v)\n", item.Summary, date)
	}

	return nil
}
