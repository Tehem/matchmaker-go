package types

import (
	"math/rand"
	"sort"
	"time"
)

// MergeRanges merges overlapping or adjacent ranges
func MergeRanges(ranges []*Range) []*Range {
	if len(ranges) == 0 {
		return ranges
	}

	// Sort ranges by start time
	sort.Slice(ranges, func(i, j int) bool {
		return ranges[i].Start.Before(ranges[j].Start)
	})

	merged := []*Range{}
	current := ranges[0]

	for i := 1; i < len(ranges); i++ {
		next := ranges[i]
		if current.End.After(next.Start) || current.End.Equal(next.Start) {
			// Ranges overlap or are adjacent, merge them
			if next.End.After(current.End) {
				current.End = next.End
			}
		} else {
			// Ranges don't overlap, add current to merged and move to next
			merged = append(merged, current)
			current = next
		}
	}

	// Add the last range
	merged = append(merged, current)

	return merged
}

// Pad adds padding to a range
func (r *Range) Pad(padding time.Duration) *Range {
	return &Range{
		Start: r.Start.Add(-padding),
		End:   r.End.Add(padding),
	}
}

// GenerateTimeRanges generates time ranges for the given work ranges
func GenerateTimeRanges(workRanges []*Range, sessionDuration time.Duration) []*Range {
	ranges := []*Range{}

	for _, workRange := range workRanges {
		start := workRange.Start
		for start.Before(workRange.End) {
			end := start.Add(sessionDuration)
			if end.After(workRange.End) {
				break
			}
			ranges = append(ranges, &Range{
				Start: start,
				End:   end,
			})
			start = end
		}
	}

	// Shuffle the ranges
	for i := range ranges {
		j := rand.Intn(i + 1)
		ranges[i], ranges[j] = ranges[j], ranges[i]
	}

	// Sort by decreasing length
	sort.Sort(byDecreasingLength(ranges))

	return ranges
}

type byDecreasingLength []*Range

func (a byDecreasingLength) Len() int      { return len(a) }
func (a byDecreasingLength) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byDecreasingLength) Less(i, j int) bool {
	return a[j].Minutes() < a[i].Minutes()
}
