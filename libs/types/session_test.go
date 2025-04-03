package types

import (
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestReviewSession(t *testing.T) {
	// Create test persons
	person1 := &Person{Email: "john.doe@example.com"}
	person2 := &Person{Email: "jane.smith@example.com"}

	// Create test squad
	squad := &Squad{
		People: []*Person{person1, person2},
	}

	// Create test time range
	start := time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	timeRange := &Range{
		Start: start,
		End:   end,
	}

	// Create test session
	session := &ReviewSession{
		Reviewers: squad,
		Range:     timeRange,
	}

	// Test Start() method
	if !session.Start().Equal(start) {
		t.Errorf("Start() returned %v, want %v", session.Start(), start)
	}

	// Test End() method
	if !session.End().Equal(end) {
		t.Errorf("End() returned %v, want %v", session.End(), end)
	}

	// Test GetDisplayName() method
	viper.SetDefault("sessions.sessionPrefix", "Review")
	expectedDisplayName := "Review - john.doe & jane.smith"
	if session.GetDisplayName() != expectedDisplayName {
		t.Errorf("GetDisplayName() returned %q, want %q", session.GetDisplayName(), expectedDisplayName)
	}

	// Test Validate() method
	if err := session.Validate(); err != nil {
		t.Errorf("Validate() returned error: %v", err)
	}

	// Test Validate() with nil reviewers
	invalidSession := &ReviewSession{
		Reviewers: nil,
		Range:     timeRange,
	}
	if err := invalidSession.Validate(); err == nil {
		t.Error("Validate() with nil reviewers should return error")
	}

	// Test Validate() with nil range
	invalidSession = &ReviewSession{
		Reviewers: squad,
		Range:     nil,
	}
	if err := invalidSession.Validate(); err == nil {
		t.Error("Validate() with nil range should return error")
	}
}

func TestSquad(t *testing.T) {
	// Create test persons
	person1 := &Person{Email: "john.doe@example.com"}
	person2 := &Person{Email: "jane.smith@example.com"}

	// Create test squad
	squad := &Squad{
		People: []*Person{person1, person2},
	}

	// Test GetDisplayName() method
	expectedDisplayName := "john.doe & jane.smith"
	if squad.GetDisplayName() != expectedDisplayName {
		t.Errorf("GetDisplayName() returned %q, want %q", squad.GetDisplayName(), expectedDisplayName)
	}

	// Test Validate() method
	if err := squad.Validate(); err != nil {
		t.Errorf("Validate() returned error: %v", err)
	}

	// Test Validate() with wrong number of people
	invalidSquad := &Squad{
		People: []*Person{person1},
	}
	if err := invalidSquad.Validate(); err == nil {
		t.Error("Validate() with wrong number of people should return error")
	}

	// Test Validate() with nil person
	invalidSquad = &Squad{
		People: []*Person{person1, nil},
	}
	if err := invalidSquad.Validate(); err == nil {
		t.Error("Validate() with nil person should return error")
	}

	// Test Validate() with too many people
	invalidSquad = &Squad{
		People: []*Person{person1, person2, person1},
	}
	if err := invalidSquad.Validate(); err == nil {
		t.Error("Validate() with too many people should return error")
	}
}
