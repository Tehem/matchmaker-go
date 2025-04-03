package solver

import (
	"matchmaker/libs/types"
	"testing"
	"time"
)

func TestGenerateSquads(t *testing.T) {
	// Create test persons
	person1 := &types.Person{Email: "person1@example.com", IsGoodReviewer: true}
	person2 := &types.Person{Email: "person2@example.com", IsGoodReviewer: false}
	person3 := &types.Person{Email: "person3@example.com", IsGoodReviewer: true}
	person4 := &types.Person{Email: "person4@example.com", IsGoodReviewer: false}

	// Create test busy times
	start := time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC)
	busyTime1 := &types.BusyTime{
		Person: person1,
		Range: &types.Range{
			Start: start.Add(2 * time.Hour),
			End:   start.Add(4 * time.Hour),
		},
	}
	busyTime2 := &types.BusyTime{
		Person: person2,
		Range: &types.Range{
			Start: start.Add(3 * time.Hour),
			End:   start.Add(5 * time.Hour),
		},
	}

	// Test with no busy times
	squads := generateSquads([]*types.Person{person1, person2, person3, person4}, []*types.BusyTime{})

	// Verify that we got the correct number of squads
	// Expected: 2 masters (person1, person3) and 2 disciples (person2, person4)
	// Master-Disciple pairs: 2 * 2 = 4
	// Master-Master pairs: 1 (person1-person3)
	// Total: 5 squads
	if len(squads) != 5 {
		t.Errorf("generateSquads() returned %d squads, want 5", len(squads))
	}

	// Test with busy times
	squads = generateSquads([]*types.Person{person1, person2, person3, person4}, []*types.BusyTime{busyTime1, busyTime2})

	// Verify that we got the correct number of squads
	if len(squads) != 5 {
		t.Errorf("generateSquads() returned %d squads, want 5", len(squads))
	}

	// Verify that the busy ranges are correct for each squad
	for _, squad := range squads {
		// Check if the squad contains person1 or person2
		hasPerson1 := false
		hasPerson2 := false
		for _, person := range squad.People {
			if person == person1 {
				hasPerson1 = true
			}
			if person == person2 {
				hasPerson2 = true
			}
		}

		// If the squad contains both person1 and person2, it should have one merged busy range
		if hasPerson1 && hasPerson2 {
			if len(squad.BusyRanges) != 1 {
				t.Errorf("generateSquads() squad %s-%s has %d busy ranges, want 1 (merged)",
					squad.People[0].Email, squad.People[1].Email,
					len(squad.BusyRanges))
			}
			// Verify the merged range
			mergedRange := squad.BusyRanges[0]
			if !mergedRange.Start.Equal(start.Add(2 * time.Hour)) {
				t.Errorf("generateSquads() merged range start time is %v, want %v",
					mergedRange.Start, start.Add(2*time.Hour))
			}
			if !mergedRange.End.Equal(start.Add(5 * time.Hour)) {
				t.Errorf("generateSquads() merged range end time is %v, want %v",
					mergedRange.End, start.Add(5*time.Hour))
			}
		} else if hasPerson1 || hasPerson2 {
			// If the squad contains only person1 or person2, it should have one busy range
			if len(squad.BusyRanges) != 1 {
				t.Errorf("generateSquads() squad %s-%s has %d busy ranges, want 1",
					squad.People[0].Email, squad.People[1].Email,
					len(squad.BusyRanges))
			}
		} else {
			// If the squad contains neither person1 nor person2, it should have no busy ranges
			if len(squad.BusyRanges) != 0 {
				t.Errorf("generateSquads() squad %s-%s has %d busy ranges, want 0",
					squad.People[0].Email, squad.People[1].Email,
					len(squad.BusyRanges))
			}
		}
	}
}

func TestFilterPersons(t *testing.T) {
	// Create test persons
	person1 := &types.Person{Email: "person1@example.com", IsGoodReviewer: true}
	person2 := &types.Person{Email: "person2@example.com", IsGoodReviewer: false}
	person3 := &types.Person{Email: "person3@example.com", IsGoodReviewer: true}
	person4 := &types.Person{Email: "person4@example.com", IsGoodReviewer: false}

	// Test with isGoodReviewer = true
	masters := filterPersons([]*types.Person{person1, person2, person3, person4}, true)
	if len(masters) != 2 {
		t.Errorf("filterPersons() returned %d masters, want 2", len(masters))
	}

	// Test with isGoodReviewer = false
	disciples := filterPersons([]*types.Person{person1, person2, person3, person4}, false)
	if len(disciples) != 2 {
		t.Errorf("filterPersons() returned %d disciples, want 2", len(disciples))
	}
}

func TestMergeBusyRanges(t *testing.T) {
	// Create test persons
	person1 := &types.Person{Email: "person1@example.com"}
	person2 := &types.Person{Email: "person2@example.com"}

	// Create test busy times
	start := time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC)
	busyTime1 := &types.BusyTime{
		Person: person1,
		Range: &types.Range{
			Start: start.Add(2 * time.Hour),
			End:   start.Add(4 * time.Hour),
		},
	}
	busyTime2 := &types.BusyTime{
		Person: person1,
		Range: &types.Range{
			Start: start.Add(3 * time.Hour),
			End:   start.Add(5 * time.Hour),
		},
	}
	busyTime3 := &types.BusyTime{
		Person: person2,
		Range: &types.Range{
			Start: start.Add(6 * time.Hour),
			End:   start.Add(8 * time.Hour),
		},
	}

	// Test with no busy times
	busyRanges := mergeBusyRanges([]*types.BusyTime{}, []*types.Person{person1, person2})
	if len(busyRanges) != 0 {
		t.Errorf("mergeBusyRanges() returned %d busy ranges, want 0", len(busyRanges))
	}

	// Test with busy times
	busyRanges = mergeBusyRanges([]*types.BusyTime{busyTime1, busyTime2, busyTime3}, []*types.Person{person1, person2})

	// Verify that we got the correct number of busy ranges
	// Expected: 2 ranges (merged busyTime1 and busyTime2, and busyTime3)
	if len(busyRanges) != 2 {
		t.Errorf("mergeBusyRanges() returned %d busy ranges, want 2", len(busyRanges))
	}

	// Verify that the ranges are merged correctly
	// First range should be from 2:00 to 5:00 (merged busyTime1 and busyTime2)
	if !busyRanges[0].Start.Equal(start.Add(2 * time.Hour)) {
		t.Errorf("mergeBusyRanges() first range start time is %v, want %v", busyRanges[0].Start, start.Add(2*time.Hour))
	}
	if !busyRanges[0].End.Equal(start.Add(5 * time.Hour)) {
		t.Errorf("mergeBusyRanges() first range end time is %v, want %v", busyRanges[0].End, start.Add(5*time.Hour))
	}

	// Second range should be from 6:00 to 8:00 (busyTime3)
	if !busyRanges[1].Start.Equal(start.Add(6 * time.Hour)) {
		t.Errorf("mergeBusyRanges() second range start time is %v, want %v", busyRanges[1].Start, start.Add(6*time.Hour))
	}
	if !busyRanges[1].End.Equal(start.Add(8 * time.Hour)) {
		t.Errorf("mergeBusyRanges() second range end time is %v, want %v", busyRanges[1].End, start.Add(8*time.Hour))
	}
}

func TestMergeRanges(t *testing.T) {
	// Create test ranges
	start := time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC)
	range1 := &types.Range{
		Start: start.Add(2 * time.Hour),
		End:   start.Add(4 * time.Hour),
	}
	range2 := &types.Range{
		Start: start.Add(3 * time.Hour),
		End:   start.Add(5 * time.Hour),
	}
	range3 := &types.Range{
		Start: start.Add(6 * time.Hour),
		End:   start.Add(8 * time.Hour),
	}

	// Test with no ranges
	merged := mergeRanges([]*types.Range{})
	if len(merged) != 0 {
		t.Errorf("mergeRanges() returned %d ranges, want 0", len(merged))
	}

	// Test with one range
	merged = mergeRanges([]*types.Range{range1})
	if len(merged) != 1 {
		t.Errorf("mergeRanges() returned %d ranges, want 1", len(merged))
	}

	// Test with overlapping ranges
	merged = mergeRanges([]*types.Range{range1, range2, range3})

	// Verify that we got the correct number of ranges
	// Expected: 2 ranges (merged range1 and range2, and range3)
	if len(merged) != 2 {
		t.Errorf("mergeRanges() returned %d ranges, want 2", len(merged))
	}

	// Verify that the ranges are merged correctly
	// First range should be from 2:00 to 5:00 (merged range1 and range2)
	if !merged[0].Start.Equal(start.Add(2 * time.Hour)) {
		t.Errorf("mergeRanges() first range start time is %v, want %v", merged[0].Start, start.Add(2*time.Hour))
	}
	if !merged[0].End.Equal(start.Add(5 * time.Hour)) {
		t.Errorf("mergeRanges() first range end time is %v, want %v", merged[0].End, start.Add(5*time.Hour))
	}

	// Second range should be from 6:00 to 8:00 (range3)
	if !merged[1].Start.Equal(start.Add(6 * time.Hour)) {
		t.Errorf("mergeRanges() second range start time is %v, want %v", merged[1].Start, start.Add(6*time.Hour))
	}
	if !merged[1].End.Equal(start.Add(8 * time.Hour)) {
		t.Errorf("mergeRanges() second range end time is %v, want %v", merged[1].End, start.Add(8*time.Hour))
	}
}

func TestHaveIntersection(t *testing.T) {
	// Create test ranges
	start := time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC)
	range1 := &types.Range{
		Start: start.Add(2 * time.Hour),
		End:   start.Add(4 * time.Hour),
	}
	range2 := &types.Range{
		Start: start.Add(3 * time.Hour),
		End:   start.Add(5 * time.Hour),
	}
	range3 := &types.Range{
		Start: start.Add(6 * time.Hour),
		End:   start.Add(8 * time.Hour),
	}

	// Test with overlapping ranges
	if !haveIntersection(range1, range2) {
		t.Error("haveIntersection() returned false for overlapping ranges")
	}

	// Test with non-overlapping ranges
	if haveIntersection(range1, range3) {
		t.Error("haveIntersection() returned true for non-overlapping ranges")
	}

	// Test with adjacent ranges
	range4 := &types.Range{
		Start: start.Add(4 * time.Hour),
		End:   start.Add(6 * time.Hour),
	}
	if haveIntersection(range1, range4) {
		t.Error("haveIntersection() returned true for adjacent ranges")
	}
}
