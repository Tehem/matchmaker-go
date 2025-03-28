package config

import (
	"testing"
	"time"

	"matchmaker/internal/calendar"
	"matchmaker/internal/matching"
	"matchmaker/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadPeople(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	SetFileSystem(fs)
	defer SetFileSystem(DefaultFileSystem{})

	fs.WriteFile("persons.yml", []byte(`- email: john.doe@example.com
  isgoodreviewer: true
  skills:
    - frontend
    - backend
- email: chuck.norris@example.com
  isgoodreviewer: true
  maxsessionsperweek: 1
  skills:
    - frontend
    - data`), 0644)

	// Test loading
	people, err := LoadPeople("persons.yml")
	require.NoError(t, err)
	require.Len(t, people, 2)

	// Check first person
	assert.Equal(t, "john.doe@example.com", people[0].Email)
	assert.True(t, people[0].IsGoodReviewer)
	assert.Equal(t, []string{"frontend", "backend"}, people[0].Skills)

	// Check second person
	assert.Equal(t, "chuck.norris@example.com", people[1].Email)
	assert.True(t, people[1].IsGoodReviewer)
	assert.Equal(t, 1, people[1].MaxSessionsPerWeek)
	assert.Equal(t, []string{"frontend", "data"}, people[1].Skills)
}

func TestProblemYAML(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	SetFileSystem(fs)
	defer SetFileSystem(DefaultFileSystem{})

	// Create test data
	people := []*matching.Person{
		{
			Email:          "john.doe@example.com",
			IsGoodReviewer: true,
			Skills:         []string{"frontend", "backend"},
		},
	}
	targetWeek := time.Date(2024, 3, 25, 0, 0, 0, 0, time.UTC)

	// Test saving
	err := SaveProblem(people, targetWeek, "problem.yml")
	require.NoError(t, err)

	// Test loading
	loadedPeople, err := LoadProblem("problem.yml")
	require.NoError(t, err)
	require.Len(t, loadedPeople, 1)

	// Check loaded data
	assert.Equal(t, people[0].Email, loadedPeople[0].Email)
	assert.Equal(t, people[0].IsGoodReviewer, loadedPeople[0].IsGoodReviewer)
	assert.Equal(t, people[0].Skills, loadedPeople[0].Skills)
}

func TestPlanningYAML(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	SetFileSystem(fs)
	defer SetFileSystem(DefaultFileSystem{})

	// Create test data
	matches := []matching.Match{
		{
			Reviewer1: &matching.Person{
				Email:          "john.doe@example.com",
				IsGoodReviewer: true,
				Skills:         []string{"frontend", "backend"},
			},
			Reviewer2: &matching.Person{
				Email:          "jane.doe@example.com",
				IsGoodReviewer: false,
				Skills:         []string{"frontend"},
			},
			TimeSlot: calendar.TimeSlot{
				Start: time.Date(2024, 3, 25, 10, 0, 0, 0, time.UTC),
				End:   time.Date(2024, 3, 25, 11, 0, 0, 0, time.UTC),
			},
			CommonSkills: []string{"frontend"},
		},
	}

	// Test saving
	err := SavePlanning(matches, "planning.yml")
	require.NoError(t, err)

	// Test loading
	loadedMatches, err := LoadPlanning("planning.yml")
	require.NoError(t, err)
	require.Len(t, loadedMatches, 1)

	// Check loaded data
	assert.Equal(t, matches[0].Reviewer1.Email, loadedMatches[0].Reviewer1.Email)
	assert.Equal(t, matches[0].Reviewer2.Email, loadedMatches[0].Reviewer2.Email)
	assert.Equal(t, matches[0].TimeSlot.Start, loadedMatches[0].TimeSlot.Start)
	assert.Equal(t, matches[0].TimeSlot.End, loadedMatches[0].TimeSlot.End)
	assert.Equal(t, matches[0].CommonSkills, loadedMatches[0].CommonSkills)
}

func TestInvalidYAML(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	SetFileSystem(fs)
	defer SetFileSystem(DefaultFileSystem{})

	fs.WriteFile("invalid.yml", []byte(`invalid: yaml: content: - email: john.doe@example.com`), 0644)

	// Test loading invalid YAML
	_, err := LoadPeople("invalid.yml")
	assert.Error(t, err)

	_, err = LoadProblem("invalid.yml")
	assert.Error(t, err)

	_, err = LoadPlanning("invalid.yml")
	assert.Error(t, err)
}
