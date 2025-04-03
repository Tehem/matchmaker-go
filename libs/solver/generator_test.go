package solver

import (
	"matchmaker/libs/testutils"
	"matchmaker/libs/types"
	"testing"
	"time"
)

func TestGenerateTimeRanges(t *testing.T) {
	// Create a config mock
	configMock := testutils.NewConfigMock()
	configMock.SetupWorkHours()
	defer configMock.Restore()

	// Create a work range (Monday 9:00-17:00)
	start := time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC)
	end := time.Date(2024, 4, 1, 17, 0, 0, 0, time.UTC)
	workRange := &types.Range{Start: start, End: end}

	// Generate time ranges
	ranges := GenerateTimeRanges([]*types.Range{workRange})

	// Verify the results
	if len(ranges) == 0 {
		t.Error("GenerateTimeRanges() returned no ranges")
		return
	}

	// Check that all ranges are within the work range
	for _, r := range ranges {
		if r.Start.Before(workRange.Start) || r.End.After(workRange.End) {
			t.Errorf("Range %v-%v is outside work range %v-%v", r.Start, r.End, workRange.Start, workRange.End)
		}
	}

	// Check that ranges don't overlap
	for i := 0; i < len(ranges)-1; i++ {
		if ranges[i].End.After(ranges[i+1].Start) {
			t.Errorf("Ranges %v-%v and %v-%v overlap", ranges[i].Start, ranges[i].End, ranges[i+1].Start, ranges[i+1].End)
		}
	}

	// Check that ranges are consecutive
	for i := 0; i < len(ranges)-1; i++ {
		if !ranges[i].End.Equal(ranges[i+1].Start) {
			t.Errorf("Ranges %v-%v and %v-%v are not consecutive", ranges[i].Start, ranges[i].End, ranges[i+1].Start, ranges[i+1].End)
		}
	}
}

func TestGenerateSessions(t *testing.T) {
	// Create a squad
	squad := &types.Squad{
		People: []*types.Person{
			{Email: "person1@example.com"},
			{Email: "person2@example.com"},
		},
	}

	// Create time ranges
	start := time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC)
	ranges := []*types.Range{
		{Start: start, End: start.Add(time.Hour)},
		{Start: start.Add(time.Hour), End: start.Add(2 * time.Hour)},
	}

	// Generate sessions
	sessions := GenerateSessions([]*types.Squad{squad}, ranges)

	// Verify the results
	if len(sessions) != len(ranges) {
		t.Errorf("GenerateSessions() returned %d sessions, want %d", len(sessions), len(ranges))
		return
	}

	// Check that each session has the correct squad and time range
	for i, session := range sessions {
		if session.Reviewers != squad {
			t.Errorf("Session %d has incorrect squad", i)
		}
		if session.Range != ranges[i] {
			t.Errorf("Session %d has incorrect time range", i)
		}
	}
}

func TestByStart(t *testing.T) {
	// Create sessions with different start times
	start := time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC)
	sessions := []*types.ReviewSession{
		{
			Reviewers: &types.Squad{},
			Range:     &types.Range{Start: start.Add(2 * time.Hour), End: start.Add(3 * time.Hour)},
		},
		{
			Reviewers: &types.Squad{},
			Range:     &types.Range{Start: start, End: start.Add(time.Hour)},
		},
		{
			Reviewers: &types.Squad{},
			Range:     &types.Range{Start: start.Add(time.Hour), End: start.Add(2 * time.Hour)},
		},
	}

	// Sort the sessions
	byStart := ByStart(sessions)
	if byStart.Len() != len(sessions) {
		t.Errorf("ByStart.Len() = %d, want %d", byStart.Len(), len(sessions))
	}

	// Test Less
	if !byStart.Less(1, 0) {
		t.Error("ByStart.Less(1, 0) = false, want true")
	}
	if byStart.Less(0, 1) {
		t.Error("ByStart.Less(0, 1) = true, want false")
	}

	// Test Swap
	byStart.Swap(0, 1)
	if !sessions[0].Start().Equal(start) {
		t.Errorf("After Swap, sessions[0].Start() = %v, want %v", sessions[0].Start(), start)
	}
}
