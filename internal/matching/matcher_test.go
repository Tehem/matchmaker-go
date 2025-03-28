package matching

import (
	"testing"
	"time"

	"matchmaker/internal/calendar"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMatcher(t *testing.T) {
	people := []*Person{
		{
			Email:              "john.doe@example.com",
			IsGoodReviewer:     true,
			Skills:             []string{"frontend", "backend"},
			MaxSessionsPerWeek: 3,
		},
	}
	config := &Config{
		SessionDuration:     60 * time.Minute,
		MinSessionSpacing:   2 * time.Hour,
		MaxPerPersonPerWeek: 3,
	}

	matcher := NewMatcher(people, config)
	assert.NotNil(t, matcher)
	assert.Equal(t, people, matcher.people)
	assert.Equal(t, config, matcher.config)
}

func TestFindMatches(t *testing.T) {
	// Create test data
	people := []*Person{
		{
			Email:              "john.doe@example.com",
			IsGoodReviewer:     true,
			Skills:             []string{"frontend", "backend"},
			MaxSessionsPerWeek: 3,
			FreeSlots: []calendar.TimeSlot{
				{
					Start: time.Date(2024, 3, 25, 10, 0, 0, 0, time.UTC),
					End:   time.Date(2024, 3, 25, 11, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			Email:              "jane.doe@example.com",
			IsGoodReviewer:     false,
			Skills:             []string{"frontend"},
			MaxSessionsPerWeek: 3,
			FreeSlots: []calendar.TimeSlot{
				{
					Start: time.Date(2024, 3, 25, 10, 0, 0, 0, time.UTC),
					End:   time.Date(2024, 3, 25, 11, 0, 0, 0, time.UTC),
				},
			},
		},
	}
	config := &Config{
		SessionDuration:     60 * time.Minute,
		MinSessionSpacing:   2 * time.Hour,
		MaxPerPersonPerWeek: 3,
	}

	matcher := NewMatcher(people, config)
	matches, err := matcher.FindMatches()
	require.NoError(t, err)
	require.Len(t, matches, 1)

	// Check match details
	match := matches[0]
	assert.Equal(t, people[0], match.Reviewer2)
	assert.Equal(t, people[1], match.Reviewer1)
	assert.Equal(t, people[0].FreeSlots[0].Start, match.TimeSlot.Start)
	assert.Equal(t, people[0].FreeSlots[0].End, match.TimeSlot.End)
	assert.Equal(t, []string{"frontend"}, match.CommonSkills)
}

func TestFindMatchesNoCommonSlots(t *testing.T) {
	people := []*Person{
		{
			Email:              "john.doe@example.com",
			IsGoodReviewer:     true,
			Skills:             []string{"frontend", "backend"},
			MaxSessionsPerWeek: 3,
			FreeSlots: []calendar.TimeSlot{
				{
					Start: time.Date(2024, 3, 25, 10, 0, 0, 0, time.UTC),
					End:   time.Date(2024, 3, 25, 11, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			Email:              "jane.doe@example.com",
			IsGoodReviewer:     false,
			Skills:             []string{"frontend"},
			MaxSessionsPerWeek: 3,
			FreeSlots: []calendar.TimeSlot{
				{
					Start: time.Date(2024, 3, 25, 14, 0, 0, 0, time.UTC),
					End:   time.Date(2024, 3, 25, 15, 0, 0, 0, time.UTC),
				},
			},
		},
	}
	config := &Config{
		SessionDuration:     60 * time.Minute,
		MinSessionSpacing:   2 * time.Hour,
		MaxPerPersonPerWeek: 3,
	}

	matcher := NewMatcher(people, config)
	matches, err := matcher.FindMatches()
	require.NoError(t, err)
	assert.Empty(t, matches)
}

func TestFindMatchesNoCommonSkills(t *testing.T) {
	people := []*Person{
		{
			Email:              "john.doe@example.com",
			IsGoodReviewer:     true,
			Skills:             []string{"frontend", "backend"},
			MaxSessionsPerWeek: 3,
			FreeSlots: []calendar.TimeSlot{
				{
					Start: time.Date(2024, 3, 25, 10, 0, 0, 0, time.UTC),
					End:   time.Date(2024, 3, 25, 11, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			Email:              "jane.doe@example.com",
			IsGoodReviewer:     false,
			Skills:             []string{"data"},
			MaxSessionsPerWeek: 3,
			FreeSlots: []calendar.TimeSlot{
				{
					Start: time.Date(2024, 3, 25, 10, 0, 0, 0, time.UTC),
					End:   time.Date(2024, 3, 25, 11, 0, 0, 0, time.UTC),
				},
			},
		},
	}
	config := &Config{
		SessionDuration:     60 * time.Minute,
		MinSessionSpacing:   2 * time.Hour,
		MaxPerPersonPerWeek: 3,
	}

	matcher := NewMatcher(people, config)
	matches, err := matcher.FindMatches()
	require.NoError(t, err)
	assert.Empty(t, matches)
}

func TestFindMatchesMaxSessionsPerWeek(t *testing.T) {
	people := []*Person{
		{
			Email:              "john.doe@example.com",
			IsGoodReviewer:     true,
			Skills:             []string{"frontend", "backend"},
			MaxSessionsPerWeek: 3,
			FreeSlots: []calendar.TimeSlot{
				{
					Start: time.Date(2024, 3, 25, 10, 0, 0, 0, time.UTC),
					End:   time.Date(2024, 3, 25, 11, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			Email:              "jane.doe@example.com",
			IsGoodReviewer:     false,
			Skills:             []string{"frontend"},
			MaxSessionsPerWeek: 1,
			FreeSlots: []calendar.TimeSlot{
				{
					Start: time.Date(2024, 3, 25, 10, 0, 0, 0, time.UTC),
					End:   time.Date(2024, 3, 25, 11, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			Email:              "bob.doe@example.com",
			IsGoodReviewer:     false,
			Skills:             []string{"frontend"},
			MaxSessionsPerWeek: 3,
			FreeSlots: []calendar.TimeSlot{
				{
					Start: time.Date(2024, 3, 25, 10, 0, 0, 0, time.UTC),
					End:   time.Date(2024, 3, 25, 11, 0, 0, 0, time.UTC),
				},
			},
		},
	}
	config := &Config{
		SessionDuration:     60 * time.Minute,
		MinSessionSpacing:   2 * time.Hour,
		MaxPerPersonPerWeek: 3,
	}

	matcher := NewMatcher(people, config)
	matches, err := matcher.FindMatches()
	require.NoError(t, err)
	require.Len(t, matches, 1)

	// Check that jane.doe@example.com is not matched again
	match := matches[0]
	assert.NotEqual(t, "jane.doe@example.com", match.Reviewer1.Email)
	assert.NotEqual(t, "jane.doe@example.com", match.Reviewer2.Email)
}
