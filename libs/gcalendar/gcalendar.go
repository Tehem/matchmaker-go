package gcalendar

import (
	"context"
	"encoding/json"
	"fmt"
	"matchmaker/libs/types"
	"matchmaker/util"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func GetGoogleCalendarService() (*calendar.Service, error) {
	ctx := context.Background()

	b, err := os.ReadFile(filepath.Join("configs", "client_secret.json"))
	if err != nil {
		return nil, err
	}

	config, err := google.ConfigFromJSON(b, calendar.CalendarScope)
	if err != nil {
		return nil, err
	}

	client := GetHttpClient(ctx, config)
	if client == nil {
		return nil, fmt.Errorf("failed to create HTTP client")
	}

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	return srv, nil
}

func FormatTime(date time.Time) string {
	return date.Format(time.RFC3339)
}

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func GetHttpClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		util.LogError(err, "Unable to get path to cached credential file")
		return nil
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	util.LogInfo("Please go to the following link in your browser", map[string]interface{}{
		"url": authURL,
	})

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		util.LogError(err, "Unable to read authorization code")
		return nil
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		util.LogError(err, "Unable to retrieve token from web")
		return nil
	}
	return tok
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

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	util.LogInfo("Saving credential file", map[string]interface{}{
		"path": file,
	})
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		util.LogError(err, "Unable to cache oauth token")
		return
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// GetBusyTimes retrieves busy time slots for a person within a given time range
func GetBusyTimes(cal *calendar.Service, person *types.Person, timeRange *types.Range) ([]*types.BusyTime, error) {
	util.LogInfo("Loading busy detail", map[string]interface{}{
		"person": person.Email,
		"start":  timeRange.Start,
		"end":    timeRange.End,
	})

	result, err := cal.Freebusy.Query(&calendar.FreeBusyRequest{
		TimeMin: FormatTime(timeRange.Start),
		TimeMax: FormatTime(timeRange.End),
		Items: []*calendar.FreeBusyRequestItem{
			{
				Id: person.Email,
			},
		},
	}).Do()
	if err != nil {
		return nil, fmt.Errorf("can't retrieve free/busy data for %s: %w", person.Email, err)
	}

	busyTimes := make([]*types.BusyTime, 0)
	busyTimePeriods := result.Calendars[person.Email].Busy
	util.LogInfo("Person busy times", map[string]interface{}{
		"person": person.Email,
	})

	for _, busyTimePeriod := range busyTimePeriods {
		util.LogInfo("Busy time period", map[string]interface{}{
			"start": busyTimePeriod.Start,
			"end":   busyTimePeriod.End,
		})
		busyTimes = append(busyTimes, &types.BusyTime{
			Person: person,
			Range: &types.Range{
				Start: parseTime(busyTimePeriod.Start),
				End:   parseTime(busyTimePeriod.End),
			},
		})
	}

	return busyTimes, nil
}

// parseTime parses a time string in RFC3339 format
func parseTime(dateStr string) time.Time {
	result, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		util.LogError(err, "Impossible to parse date "+dateStr)
		return time.Time{}
	}
	return result
}
