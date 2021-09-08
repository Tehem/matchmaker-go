package util

import (
	"github.com/transcovo/go-chpr-logger"
)

func PanicOnError(err error, message string) {
	if err != nil {
		logger.GetLogger().WithError(err).Fatal(message)
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
