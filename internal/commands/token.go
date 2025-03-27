package commands

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

// tokenCmd represents the token command
var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Get a Google Calendar API access token",
	Long: `This command will open a browser window for you to authorize the application 
to access your Google Calendar. The token will be stored in your credentials directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		// Get user's home directory
		usr, err := user.Current()
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		// Create credentials directory
		credentialsDir := filepath.Join(usr.HomeDir, ".credentials")
		if err := os.MkdirAll(credentialsDir, 0700); err != nil {
			return fmt.Errorf("failed to create credentials directory: %w", err)
		}

		// Load client secret
		clientSecret, err := os.ReadFile("client_secret.json")
		if err != nil {
			return fmt.Errorf("failed to read client secret: %w", err)
		}

		// Configure OAuth2
		config, err := google.JWTConfigFromJSON(clientSecret, calendar.CalendarScope)
		if err != nil {
			return fmt.Errorf("failed to parse client secret: %w", err)
		}

		// Create token source
		tokenSource := config.TokenSource(ctx)
		token, err := tokenSource.Token()
		if err != nil {
			return fmt.Errorf("failed to get token: %w", err)
		}

		// Save token
		tokenFile := filepath.Join(credentialsDir, "calendar-api.json")
		if err := saveToken(tokenFile, token); err != nil {
			return fmt.Errorf("failed to save token: %w", err)
		}

		slog.Info("Token saved successfully", "file", tokenFile)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(tokenCmd)
}

// saveToken saves the OAuth2 token to a file
func saveToken(filepath string, token *oauth2.Token) error {
	f, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open token file: %w", err)
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(token); err != nil {
		return fmt.Errorf("failed to encode token: %w", err)
	}

	return nil
}
