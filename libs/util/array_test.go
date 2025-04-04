package util

import (
	"reflect"
	"testing"
)

func TestIntersection(t *testing.T) {
	tests := []struct {
		name     string
		array1   []string
		array2   []string
		expected []string
	}{
		{
			name:     "Empty arrays",
			array1:   []string{},
			array2:   []string{},
			expected: []string{},
		},
		{
			name:     "One empty array",
			array1:   []string{"a", "b", "c"},
			array2:   []string{},
			expected: []string{},
		},
		{
			name:     "No common elements",
			array1:   []string{"a", "b", "c"},
			array2:   []string{"d", "e", "f"},
			expected: []string{},
		},
		{
			name:     "Some common elements",
			array1:   []string{"a", "b", "c"},
			array2:   []string{"b", "c", "d"},
			expected: []string{"b", "c"},
		},
		{
			name:     "All elements common",
			array1:   []string{"a", "b", "c"},
			array2:   []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "Duplicate elements",
			array1:   []string{"a", "b", "b", "c"},
			array2:   []string{"b", "c", "c", "d"},
			expected: []string{"b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Intersection(tt.array1, tt.array2)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Intersection() = %v, want %v", result, tt.expected)
			}
		})
	}
}
