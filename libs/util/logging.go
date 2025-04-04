package util

import (
	"matchmaker/libs/types"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// ConfigureLogging sets up logging with consistent formatting and level
func ConfigureLogging() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		PadLevelText:  true,
	})

	// Set log level based on environment
	if os.Getenv("DEBUG") == "true" {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
}

// PanicOnError logs an error with context and panics
func PanicOnError(err error, message string) {
	if err != nil {
		logrus.WithError(err).Fatal(message)
	}
}

// LogError logs an error with context
func LogError(err error, message string) {
	if err != nil {
		logrus.WithError(err).Error(message)
	}
}

// LogInfo logs an info message with optional fields
func LogInfo(message string, fields map[string]interface{}) {
	if fields != nil {
		logrus.WithFields(fields).Info(message)
	} else {
		logrus.Info(message)
	}
}

// LogDebug logs a debug message with optional fields
func LogDebug(message string, fields map[string]interface{}) {
	if fields != nil {
		logrus.WithFields(fields).Debug(message)
	} else {
		logrus.Debug(message)
	}
}

// LogRange logs a range with standardized formatting
func LogRange(message string, r *types.Range) {
	LogInfo(message, map[string]interface{}{
		"from": r.Start.Format(time.RFC3339),
		"to":   r.End.Format(time.RFC3339),
	})
}

// LogSession logs a session with standardized formatting
func LogSession(message string, session *types.ReviewSession) {
	LogInfo(message, map[string]interface{}{
		"person1": session.Reviewers.People[0].Email,
		"person2": session.Reviewers.People[1].Email,
		"from":    session.Range.Start.Format(time.RFC3339),
		"to":      session.Range.End.Format(time.RFC3339),
	})
}
