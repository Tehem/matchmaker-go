package util

// Intersection returns the common elements between two string arrays
func Intersection(array1 []string, array2 []string) []string {
	// Use a map to track unique elements
	seen := make(map[string]bool)
	commonItems := []string{}

	// Create a map of array2 elements for O(1) lookup
	array2Map := make(map[string]bool)
	for _, element := range array2 {
		array2Map[element] = true
	}

	// Check each element in array1
	for _, element := range array1 {
		if array2Map[element] && !seen[element] {
			commonItems = append(commonItems, element)
			seen[element] = true
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
