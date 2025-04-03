package types

import (
	"testing"
	"time"
)

func TestProblemSerialization(t *testing.T) {
	// Create a test problem
	person1 := &Person{Email: "person1@example.com"}
	person2 := &Person{Email: "person2@example.com"}
	workRange := &Range{
		Start: time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 4, 1, 17, 0, 0, 0, time.UTC),
	}
	busyTime := &BusyTime{
		Person: person1,
		Range: &Range{
			Start: time.Date(2024, 4, 1, 10, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 4, 1, 11, 0, 0, 0, time.UTC),
		},
	}

	problem := &Problem{
		People:           []*Person{person1, person2},
		WorkRanges:       []*Range{workRange},
		BusyTimes:        []*BusyTime{busyTime},
		TargetCoverage:   5,
		MaxTotalCoverage: 8,
	}

	// Test serialization
	yml, err := problem.ToYaml()
	if err != nil {
		t.Fatalf("ToYaml() error = %v", err)
	}

	// Test deserialization
	loadedProblem, err := LoadProblem(yml)
	if err != nil {
		t.Fatalf("LoadProblem() error = %v", err)
	}

	// Verify the loaded problem
	if len(loadedProblem.People) != len(problem.People) {
		t.Errorf("Loaded problem has %d people, want %d", len(loadedProblem.People), len(problem.People))
	}

	if len(loadedProblem.WorkRanges) != len(problem.WorkRanges) {
		t.Errorf("Loaded problem has %d work ranges, want %d", len(loadedProblem.WorkRanges), len(problem.WorkRanges))
	}

	if len(loadedProblem.BusyTimes) != len(problem.BusyTimes) {
		t.Errorf("Loaded problem has %d busy times, want %d", len(loadedProblem.BusyTimes), len(problem.BusyTimes))
	}

	if loadedProblem.TargetCoverage != problem.TargetCoverage {
		t.Errorf("Loaded problem has target coverage %d, want %d", loadedProblem.TargetCoverage, problem.TargetCoverage)
	}

	// Verify busy time person reference
	if loadedProblem.BusyTimes[0].Person.Email != problem.BusyTimes[0].Person.Email {
		t.Errorf("Loaded busy time person email = %v, want %v", loadedProblem.BusyTimes[0].Person.Email, problem.BusyTimes[0].Person.Email)
	}

	// Verify busy time range
	if !loadedProblem.BusyTimes[0].Range.Start.Equal(problem.BusyTimes[0].Range.Start) ||
		!loadedProblem.BusyTimes[0].Range.End.Equal(problem.BusyTimes[0].Range.End) {
		t.Errorf("Loaded busy time range = %v-%v, want %v-%v",
			loadedProblem.BusyTimes[0].Range.Start, loadedProblem.BusyTimes[0].Range.End,
			problem.BusyTimes[0].Range.Start, problem.BusyTimes[0].Range.End)
	}
}

func TestTuples(t *testing.T) {
	person1 := &Person{Email: "person1@example.com"}
	person2 := &Person{Email: "person2@example.com"}
	person3 := &Person{Email: "person3@example.com"}

	tuples := &Tuples{
		Pairs: []Tuple{
			{Person1: person1, Person2: person2},
		},
		UnpairedPeople: []*Person{person3},
	}

	if len(tuples.Pairs) != 1 {
		t.Errorf("Tuples has %d pairs, want 1", len(tuples.Pairs))
	}

	if tuples.Pairs[0].Person1.Email != person1.Email || tuples.Pairs[0].Person2.Email != person2.Email {
		t.Errorf("Pair contains wrong people: %v and %v, want %v and %v",
			tuples.Pairs[0].Person1.Email, tuples.Pairs[0].Person2.Email,
			person1.Email, person2.Email)
	}

	if len(tuples.UnpairedPeople) != 1 || tuples.UnpairedPeople[0].Email != person3.Email {
		t.Errorf("Unpaired people contains wrong person: %v, want %v",
			tuples.UnpairedPeople[0].Email, person3.Email)
	}
}

func TestSolution(t *testing.T) {
	person1 := &Person{Email: "person1@example.com"}
	person2 := &Person{Email: "person2@example.com"}
	timeRange := &Range{
		Start: time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 4, 1, 10, 0, 0, 0, time.UTC),
	}

	squad := &Squad{
		People: []*Person{person1, person2},
	}

	session := &ReviewSession{
		Reviewers: squad,
		Range:     timeRange,
	}

	solution := &Solution{
		Sessions: []*ReviewSession{session},
	}

	if len(solution.Sessions) != 1 {
		t.Errorf("Solution has %d sessions, want 1", len(solution.Sessions))
	}

	if solution.Sessions[0].Reviewers.People[0].Email != person1.Email ||
		solution.Sessions[0].Reviewers.People[1].Email != person2.Email {
		t.Errorf("Session contains wrong people: %v and %v, want %v and %v",
			solution.Sessions[0].Reviewers.People[0].Email, solution.Sessions[0].Reviewers.People[1].Email,
			person1.Email, person2.Email)
	}

	if !solution.Sessions[0].Range.Start.Equal(timeRange.Start) || !solution.Sessions[0].Range.End.Equal(timeRange.End) {
		t.Errorf("Session has wrong time range: %v-%v, want %v-%v",
			solution.Sessions[0].Range.Start, solution.Sessions[0].Range.End,
			timeRange.Start, timeRange.End)
	}
}
