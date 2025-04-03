package types

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// ReviewSession represents a review session between two people
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
		return fmt.Errorf("reviewers are required")
	}
	if s.Range == nil {
		return fmt.Errorf("time range is required")
	}
	if err := s.Reviewers.Validate(); err != nil {
		return fmt.Errorf("invalid reviewers: %w", err)
	}
	return nil
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
	for _, person := range s.People {
		if person == nil {
			return fmt.Errorf("person cannot be nil")
		}
		if err := person.Validate(); err != nil {
			return fmt.Errorf("invalid person: %w", err)
		}
	}
	return nil
}

// GetDisplayName returns a display name for the squad
func (s *Squad) GetDisplayName() string {
	if len(s.People) != 2 {
		return "Invalid Squad"
	}
	return fmt.Sprintf("%s / %s", s.People[0].Email, s.People[1].Email)
}

// GenerateSessions generates all possible sessions for the given squads and ranges
func GenerateSessions(squads []*Squad, ranges []*Range) []*ReviewSession {
	sessions := []*ReviewSession{}
	for _, currentRange := range ranges {
		for _, squad := range squads {
			sessionPossible := true

			for _, busyRange := range squad.BusyRanges {
				if HaveIntersection(currentRange, busyRange) {
					sessionPossible = false
					break
				}
			}

			if sessionPossible {
				sessions = append(sessions, &ReviewSession{
					Reviewers: squad,
					Range:     currentRange,
				})
			}
		}
	}
	return sessions
}

// ByStart is a type for sorting sessions by start time
type ByStart []*ReviewSession

func (a ByStart) Len() int      { return len(a) }
func (a ByStart) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByStart) Less(i, j int) bool {
	iStart := a[i].Start()
	jStart := a[j].Start()
	return iStart.Before(jStart)
}
