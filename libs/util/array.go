package util

// Intersection returns the common elements between two string arrays
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

// contains checks if a string array contains a specific element
func contains(array []string, element string) bool {
	for i := 0; i < len(array); i++ {
		if array[i] == element {
			return true
		}
	}
	return false
}
