package types

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// ReviewSession represents a review session
type ReviewSession struct {
	Reviewers *Squad
	Range     *Range
}

// Start returns the start time of the session
func (s *ReviewSession) Start() time.Time {
	return s.Range.Start
}

// End returns the end time of the session
func (s *ReviewSession) End() time.Time {
	return s.Range.End
}

// GetDisplayName returns a display name for the session
func (s *ReviewSession) GetDisplayName() string {
	sessionPrefix := viper.GetString("sessions.sessionPrefix")
	return sessionPrefix + " - " + s.Reviewers.GetDisplayName()
}

// Validate checks if the session is valid
func (s *ReviewSession) Validate() error {
	if s.Reviewers == nil {
		return fmt.Errorf("reviewers cannot be nil")
	}
	if s.Range == nil {
		return fmt.Errorf("range cannot be nil")
	}
	return s.Reviewers.Validate()
}

// Squad represents a pair of reviewers
type Squad struct {
	People     []*Person
	BusyRanges []*Range
}

// Validate checks if the squad is valid
func (s *Squad) Validate() error {
	if len(s.People) != 2 {
		return fmt.Errorf("squad must have exactly 2 people")
	}
	if s.People[0] == nil || s.People[1] == nil {
		return fmt.Errorf("people cannot be nil")
	}
	return nil
}

// GetDisplayName returns a display name for the squad
func (s *Squad) GetDisplayName() string {
	return fmt.Sprintf("%s & %s", s.People[0].Email, s.People[1].Email)
}

// DisplayName returns a display name for the squad
func (s *Squad) DisplayName() string {
	return fmt.Sprintf("%s & %s", s.People[0].Email, s.People[1].Email)
}
