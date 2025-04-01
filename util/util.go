package util

import (
	"os"

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

func Intersection(array1 []string, array2 []string) []string {
	commonItems := []string{}
	for i := 0; i < len(array1); i++ {
		element := array1[i]
		if contains(array2, element) {
			commonItems = append(commonItems, element)
		}
	}
	return commonItems
}

func contains(array []string, element string) bool {
	for i := 0; i < len(array); i++ {
		if array[i] == element {
			return true
		}
	}
	return false
}
