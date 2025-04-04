package gcalendar

import (
	"matchmaker/libs/testutils"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
)

// MockCalendarService creates a mock calendar service for testing
func MockCalendarService() (*calendar.Service, error) {
	// Create a mock service
	service := &calendar.Service{}
	return service, nil
}

func TestTokenCacheFile(t *testing.T) {
	// Create a config mock
	configMock := testutils.NewConfigMock()
	configMock.SetupWorkHours()
	defer configMock.Restore()

	// Get the token cache file path
	cacheFile, err := TokenCacheFile()
	if err != nil {
		t.Fatalf("TokenCacheFile() error = %v", err)
	}

	// Check that the cache file path is not empty
	if cacheFile == "" {
		t.Error("TokenCacheFile() returned empty path")
	}

	// Check that the cache file path contains the expected filename
	expectedFilename := "calendar-api.json"
	if filepath.Base(cacheFile) != expectedFilename {
		t.Errorf("TokenCacheFile() = %v, want to contain %v", cacheFile, expectedFilename)
	}
}

func TestGetCalendarService(t *testing.T) {
	// Create a config mock
	configMock := testutils.NewConfigMock()
	configMock.SetupWorkHours()
	defer configMock.Restore()

	// Check if the client_secret.json file exists
	clientSecretPath := filepath.Join("configs", "client_secret.json")
	if _, err := os.Stat(clientSecretPath); os.IsNotExist(err) {
		t.Skip("Skipping test: client_secret.json not found")
	}

	// Get the calendar service
	service, err := GetCalendarService()
	if err != nil {
		t.Fatalf("GetCalendarService() error = %v", err)
	}

	// Check that the service is not nil
	if service == nil {
		t.Error("GetCalendarService() returned nil service")
	}
}

func TestSaveToken(t *testing.T) {
	// Create a config mock
	configMock := testutils.NewConfigMock()
	configMock.SetupWorkHours()
	defer configMock.Restore()

	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "token-test-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Create a mock token
	token := &oauth2.Token{
		AccessToken:  "test-access-token",
		TokenType:    "Bearer",
		RefreshToken: "test-refresh-token",
		Expiry:       time.Now().Add(1 * time.Hour),
	}

	// Save the token
	SaveToken(tempFile.Name(), token)

	// Check that the file was created
	if _, err := os.Stat(tempFile.Name()); os.IsNotExist(err) {
		t.Error("SaveToken() did not create the file")
	}

	// Read the file and check that it contains the token
	fileContent, err := os.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	// Check that the file content is not empty
	if len(fileContent) == 0 {
		t.Error("SaveToken() wrote empty file")
	}

	// Check that the file content contains the token fields
	contentStr := string(fileContent)
	if !strings.Contains(contentStr, "test-access-token") {
		t.Error("SaveToken() did not write the access token")
	}
	if !strings.Contains(contentStr, "test-refresh-token") {
		t.Error("SaveToken() did not write the refresh token")
	}
}

func TestTokenFromFile(t *testing.T) {
	// Create a config mock
	configMock := testutils.NewConfigMock()
	configMock.SetupWorkHours()
	defer configMock.Restore()

	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "token-test-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Create a mock token
	token := &oauth2.Token{
		AccessToken:  "test-access-token",
		TokenType:    "Bearer",
		RefreshToken: "test-refresh-token",
		Expiry:       time.Now().Add(1 * time.Hour),
	}

	// Save the token
	SaveToken(tempFile.Name(), token)

	// Read the token from the file
	readToken, err := TokenFromFile(tempFile.Name())
	if err != nil {
		t.Fatalf("TokenFromFile() error = %v", err)
	}

	// Check that the read token has the correct fields
	if readToken.AccessToken != token.AccessToken {
		t.Errorf("TokenFromFile() access token = %v, want %v", readToken.AccessToken, token.AccessToken)
	}
	if readToken.TokenType != token.TokenType {
		t.Errorf("TokenFromFile() token type = %v, want %v", readToken.TokenType, token.TokenType)
	}
	if readToken.RefreshToken != token.RefreshToken {
		t.Errorf("TokenFromFile() refresh token = %v, want %v", readToken.RefreshToken, token.RefreshToken)
	}
}

func TestGetClient(t *testing.T) {
	// Create a config mock
	configMock := testutils.NewConfigMock()
	configMock.SetupWorkHours()
	defer configMock.Restore()

	// Create a mock OAuth2 config
	config := &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080",
		Scopes:       []string{"https://www.googleapis.com/auth/calendar"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}

	// Get the client
	client := GetClient(config)

	// Check that the client is not nil
	if client == nil {
		t.Error("GetClient() returned nil client")
	}
}
