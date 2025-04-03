package types

import (
	"matchmaker/libs/testutils"
	"sort"
	"testing"
	"time"
)

func TestGenerateSessions(t *testing.T) {
	// Create a config mock
	configMock := testutils.NewConfigMock()
	configMock.SetupWorkHours()
	defer configMock.Restore()

	// Create test persons
	person1 := &Person{Email: "person1@example.com"}
	person2 := &Person{Email: "person2@example.com"}

	// Create test squads
	squads := []*Squad{
		{
			People: []*Person{person1, person2},
		},
	}

	// Create test time ranges
	start := time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC)
	ranges := []*Range{
		{
			Start: start,
			End:   start.Add(time.Hour),
		},
		{
			Start: start.Add(time.Hour),
			End:   start.Add(2 * time.Hour),
		},
	}

	// Generate sessions
	sessions := GenerateSessions(squads, ranges)

	// Verify the results
	expectedSessions := len(squads) * len(ranges)
	if len(sessions) != expectedSessions {
		t.Errorf("GenerateSessions() returned %d sessions, want %d", len(sessions), expectedSessions)
	}

	// Verify that each session has the correct squad and range
	for i, session := range sessions {
		expectedSquad := squads[i/len(ranges)]
		expectedRange := ranges[i%len(ranges)]

		if session.Reviewers != expectedSquad {
			t.Errorf("Session %d has squad %v, want %v", i, session.Reviewers, expectedSquad)
		}
		if session.Range != expectedRange {
			t.Errorf("Session %d has range %v-%v, want %v-%v", i,
				session.Range.Start, session.Range.End,
				expectedRange.Start, expectedRange.End)
		}
	}
}

func TestByStart(t *testing.T) {
	// Create test sessions
	start := time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC)
	sessions := []*ReviewSession{
		{
			Range: &Range{
				Start: start.Add(2 * time.Hour),
				End:   start.Add(3 * time.Hour),
			},
		},
		{
			Range: &Range{
				Start: start,
				End:   start.Add(time.Hour),
			},
		},
		{
			Range: &Range{
				Start: start.Add(time.Hour),
				End:   start.Add(2 * time.Hour),
			},
		},
	}

	// Sort sessions
	sort.Sort(ByStart(sessions))

	// Verify that sessions are sorted by start time
	for i := 1; i < len(sessions); i++ {
		if sessions[i-1].Start().After(sessions[i].Start()) {
			t.Errorf("Sessions are not sorted by start time")
		}
	}
}
