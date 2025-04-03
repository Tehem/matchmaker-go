package squads

import (
	"matchmaker/libs/testutils"
	"matchmaker/libs/types"
	"testing"
	"time"
)

func TestGenerateSquads(t *testing.T) {
	// Create a config mock
	configMock := testutils.NewConfigMock()
	configMock.SetupWorkHours()
	defer configMock.Restore()

	// Create test persons
	master1 := &types.Person{Email: "master1@example.com", IsGoodReviewer: true}
	master2 := &types.Person{Email: "master2@example.com", IsGoodReviewer: true}
	disciple1 := &types.Person{Email: "disciple1@example.com", IsGoodReviewer: false}
	disciple2 := &types.Person{Email: "disciple2@example.com", IsGoodReviewer: false}

	people := []*types.Person{master1, master2, disciple1, disciple2}

	// Create test busy times
	start := time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC)
	busyTimes := []*types.BusyTime{
		{
			Person: master1,
			Range: &types.Range{
				Start: start,
				End:   start.Add(time.Hour),
			},
		},
		{
			Person: disciple1,
			Range: &types.Range{
				Start: start.Add(time.Hour),
				End:   start.Add(2 * time.Hour),
			},
		},
	}

	// Generate squads
	squads := GenerateSquads(people, busyTimes)

	// Verify the results
	// We expect:
	// - 4 master-disciple pairs (2 masters * 2 disciples)
	// - 1 master-master pair (2 masters)
	expectedSquads := 5
	if len(squads) != expectedSquads {
		t.Errorf("GenerateSquads() returned %d squads, want %d", len(squads), expectedSquads)
	}

	// Verify that each squad has the correct people
	for _, squad := range squads {
		if len(squad.People) != 2 {
			t.Errorf("Squad has %d people, want 2", len(squad.People))
		}

		// Verify that the squad has the correct busy ranges based on its members
		hasMaster1 := false
		hasDisciple1 := false
		for _, person := range squad.People {
			if person == master1 {
				hasMaster1 = true
			}
			if person == disciple1 {
				hasDisciple1 = true
			}
		}

		// If the squad contains master1 or disciple1, it should have busy ranges
		if hasMaster1 || hasDisciple1 {
			if len(squad.BusyRanges) == 0 {
				t.Error("Squad with busy members has no busy ranges")
			}
		}
	}

	// Verify that we have all expected combinations
	expectedCombinations := map[string]bool{
		"master1@example.com-master2@example.com":   false,
		"master1@example.com-disciple1@example.com": false,
		"master1@example.com-disciple2@example.com": false,
		"master2@example.com-disciple1@example.com": false,
		"master2@example.com-disciple2@example.com": false,
	}

	for _, squad := range squads {
		key := squad.People[0].Email + "-" + squad.People[1].Email
		expectedCombinations[key] = true
	}

	for combination, found := range expectedCombinations {
		if !found {
			t.Errorf("Expected combination %s not found in generated squads", combination)
		}
	}
}

func TestFilterPersons(t *testing.T) {
	// Create test persons
	master := &types.Person{Email: "master@example.com", IsGoodReviewer: true}
	disciple := &types.Person{Email: "disciple@example.com", IsGoodReviewer: false}
	people := []*types.Person{master, disciple}

	// Test filtering masters
	masters := filterPersons(people, true)
	if len(masters) != 1 || masters[0] != master {
		t.Error("filterPersons(true) did not return correct masters")
	}

	// Test filtering disciples
	disciples := filterPersons(people, false)
	if len(disciples) != 1 || disciples[0] != disciple {
		t.Error("filterPersons(false) did not return correct disciples")
	}
}

func TestMergeBusyRanges(t *testing.T) {
	// Create test persons
	person := &types.Person{Email: "person@example.com"}

	// Create test busy times with overlapping ranges
	start := time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC)
	busyTimes := []*types.BusyTime{
		{
			Person: person,
			Range: &types.Range{
				Start: start,
				End:   start.Add(time.Hour),
			},
		},
		{
			Person: person,
			Range: &types.Range{
				Start: start.Add(30 * time.Minute),
				End:   start.Add(90 * time.Minute),
			},
		},
	}

	// Merge busy ranges
	mergedRanges := mergeBusyRanges(busyTimes, []*types.Person{person})

	// Verify that overlapping ranges are merged
	if len(mergedRanges) != 1 {
		t.Errorf("mergeBusyRanges() returned %d ranges, want 1", len(mergedRanges))
	}

	// Verify the merged range
	expectedStart := start
	expectedEnd := start.Add(90 * time.Minute)
	if !mergedRanges[0].Start.Equal(expectedStart) || !mergedRanges[0].End.Equal(expectedEnd) {
		t.Errorf("Merged range is %v-%v, want %v-%v",
			mergedRanges[0].Start, mergedRanges[0].End,
			expectedStart, expectedEnd)
	}
}
