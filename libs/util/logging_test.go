package util

import (
	"bytes"
	"matchmaker/libs/types"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestConfigureLogging(t *testing.T) {
	// Save original log level
	originalLevel := logrus.GetLevel()
	defer logrus.SetLevel(originalLevel)

	// Test with DEBUG=true
	os.Setenv("DEBUG", "true")
	ConfigureLogging()
	assert.Equal(t, logrus.DebugLevel, logrus.GetLevel())

	// Test with DEBUG=false
	os.Setenv("DEBUG", "false")
	ConfigureLogging()
	assert.Equal(t, logrus.InfoLevel, logrus.GetLevel())

	// Test with DEBUG not set
	os.Unsetenv("DEBUG")
	ConfigureLogging()
	assert.Equal(t, logrus.InfoLevel, logrus.GetLevel())
}

func TestLogFunctions(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Save original logger settings
	originalOut := logrus.StandardLogger().Out
	originalFormatter := logrus.StandardLogger().Formatter
	originalLevel := logrus.GetLevel()

	// Configure logger for testing
	logrus.SetOutput(&buf)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
		DisableColors:    true,
	})
	logrus.SetLevel(logrus.DebugLevel)

	// Restore original settings after test
	defer func() {
		logrus.SetOutput(originalOut)
		logrus.SetFormatter(originalFormatter)
		logrus.SetLevel(originalLevel)
	}()

	// Test LogInfo
	LogInfo("test info message", nil)
	assert.Contains(t, buf.String(), "test info message")
	assert.Contains(t, buf.String(), "level=info")
	buf.Reset()

	// Test LogInfo with fields
	fields := map[string]interface{}{"key": "value"}
	LogInfo("test info with fields", fields)
	assert.Contains(t, buf.String(), "test info with fields")
	assert.Contains(t, buf.String(), "key=value")
	buf.Reset()

	// Test LogDebug
	LogDebug("test debug message", nil)
	assert.Contains(t, buf.String(), "test debug message")
	assert.Contains(t, buf.String(), "level=debug")
	buf.Reset()

	// Test LogDebug with fields
	LogDebug("test debug with fields", fields)
	assert.Contains(t, buf.String(), "test debug with fields")
	assert.Contains(t, buf.String(), "key=value")
	buf.Reset()

	// Test LogError
	LogError(nil, "test error message")
	output := buf.String()
	assert.Contains(t, output, "test error message")
	assert.Contains(t, output, "level=error")
	buf.Reset()

	// Test LogError with error
	err := assert.AnError
	LogError(err, "test error with error")
	output = buf.String()
	assert.Contains(t, output, "test error with error")
	assert.Contains(t, output, "level=error")
	assert.Contains(t, output, err.Error())
	buf.Reset()
}

func TestLogRange(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	defer logrus.SetOutput(os.Stderr)

	// Create a test range
	now := time.Now()
	r := &types.Range{
		Start: now,
		End:   now.Add(time.Hour),
	}

	// Test LogRange
	LogRange("test range", r)
	assert.Contains(t, buf.String(), "test range")
	assert.Contains(t, buf.String(), "from=")
	assert.Contains(t, buf.String(), "to=")
}

func TestLogSession(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	defer logrus.SetOutput(os.Stderr)

	// Create a test session
	now := time.Now()
	session := &types.ReviewSession{
		Reviewers: &types.Squad{
			People: []*types.Person{
				{Email: "person1@example.com"},
				{Email: "person2@example.com"},
			},
		},
		Range: &types.Range{
			Start: now,
			End:   now.Add(time.Hour),
		},
	}

	// Test LogSession
	LogSession("test session", session)
	assert.Contains(t, buf.String(), "test session")
	assert.Contains(t, buf.String(), "person1=person1@example.com")
	assert.Contains(t, buf.String(), "person2=person2@example.com")
	assert.Contains(t, buf.String(), "from=")
	assert.Contains(t, buf.String(), "to=")
}
