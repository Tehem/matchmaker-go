package types

import (
	"testing"
	"time"
)

func TestMergeRanges(t *testing.T) {
	// Create test ranges
	start := time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC)
	range1 := &Range{
		Start: start.Add(2 * time.Hour),
		End:   start.Add(4 * time.Hour),
	}
	range2 := &Range{
		Start: start.Add(3 * time.Hour),
		End:   start.Add(5 * time.Hour),
	}
	range3 := &Range{
		Start: start.Add(6 * time.Hour),
		End:   start.Add(8 * time.Hour),
	}

	// Test with no ranges
	merged := MergeRanges([]*Range{})
	if len(merged) != 0 {
		t.Errorf("MergeRanges() returned %d ranges, want 0", len(merged))
	}

	// Test with one range
	merged = MergeRanges([]*Range{range1})
	if len(merged) != 1 {
		t.Errorf("MergeRanges() returned %d ranges, want 1", len(merged))
	}

	// Test with overlapping ranges
	merged = MergeRanges([]*Range{range1, range2, range3})

	// Verify that we got the correct number of ranges
	// Expected: 2 ranges (merged range1 and range2, and range3)
	if len(merged) != 2 {
		t.Errorf("MergeRanges() returned %d ranges, want 2", len(merged))
	}

	// Verify that the ranges are merged correctly
	// First range should be from 2:00 to 5:00 (merged range1 and range2)
	if !merged[0].Start.Equal(start.Add(2 * time.Hour)) {
		t.Errorf("MergeRanges() first range start time is %v, want %v", merged[0].Start, start.Add(2*time.Hour))
	}
	if !merged[0].End.Equal(start.Add(5 * time.Hour)) {
		t.Errorf("MergeRanges() first range end time is %v, want %v", merged[0].End, start.Add(5*time.Hour))
	}

	// Second range should be from 6:00 to 8:00 (range3)
	if !merged[1].Start.Equal(start.Add(6 * time.Hour)) {
		t.Errorf("MergeRanges() second range start time is %v, want %v", merged[1].Start, start.Add(6*time.Hour))
	}
	if !merged[1].End.Equal(start.Add(8 * time.Hour)) {
		t.Errorf("MergeRanges() second range end time is %v, want %v", merged[1].End, start.Add(8*time.Hour))
	}
}

func TestPad(t *testing.T) {
	start := time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC)
	r := &Range{
		Start: start,
		End:   start.Add(1 * time.Hour),
	}

	// Test with 30 minutes padding
	padded := r.Pad(30 * time.Minute)
	if !padded.Start.Equal(start.Add(-30 * time.Minute)) {
		t.Errorf("Pad() start time is %v, want %v", padded.Start, start.Add(-30*time.Minute))
	}
	if !padded.End.Equal(start.Add(90 * time.Minute)) {
		t.Errorf("Pad() end time is %v, want %v", padded.End, start.Add(90*time.Minute))
	}
}

func TestGenerateTimeRanges(t *testing.T) {
	// Create a work range from 9:00 to 17:00
	workRange := &Range{
		Start: time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 4, 1, 17, 0, 0, 0, time.UTC),
	}

	// Generate time ranges with 30-minute sessions
	sessionDuration := 30 * time.Minute
	ranges := GenerateTimeRanges([]*Range{workRange}, sessionDuration)

	// Verify that we got the correct number of ranges
	// 8 hours = 480 minutes, with 30-minute sessions = 16 sessions
	if len(ranges) != 16 {
		t.Errorf("GenerateTimeRanges() returned %d ranges, want 16", len(ranges))
	}

	// Verify that each range has the correct duration
	for _, r := range ranges {
		if r.Minutes() != 30 {
			t.Errorf("GenerateTimeRanges() range duration is %v minutes, want 30", r.Minutes())
		}
	}

	// Verify that ranges are sorted by decreasing length
	for i := 1; i < len(ranges); i++ {
		if ranges[i-1].Minutes() < ranges[i].Minutes() {
			t.Errorf("GenerateTimeRanges() ranges are not sorted by decreasing length")
		}
	}

	// Verify that all ranges are within the work range
	for _, r := range ranges {
		if r.Start.Before(workRange.Start) || r.End.After(workRange.End) {
			t.Errorf("GenerateTimeRanges() range %v-%v is outside work range %v-%v", r.Start, r.End, workRange.Start, workRange.End)
		}
	}

	// Verify that ranges don't overlap
	for i := 0; i < len(ranges); i++ {
		for j := i + 1; j < len(ranges); j++ {
			if ranges[i].Overlaps(ranges[j]) {
				t.Errorf("GenerateTimeRanges() ranges %v-%v and %v-%v overlap", ranges[i].Start, ranges[i].End, ranges[j].Start, ranges[j].End)
			}
		}
	}
}
