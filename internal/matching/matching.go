package matching

import (
	"fmt"
	"time"

	"matchmaker/internal/calendar"
)

// Person represents a reviewer
type Person struct {
	Email              string   `yaml:"email"`
	IsGoodReviewer     bool     `yaml:"isgoodreviewer"`
	MaxSessionsPerWeek int      `yaml:"maxsessionsperweek,omitempty"`
	Skills             []string `yaml:"skills,omitempty"`
	FreeSlots          []calendar.TimeSlot
}

// Match represents a matched pair of reviewers
type Match struct {
	Reviewer1    *Person
	Reviewer2    *Person
	TimeSlot     calendar.TimeSlot
	CommonSkills []string
}

// Matcher handles the matching logic
type Matcher struct {
	people []*Person
	config *Config
}

// Config represents the matching configuration
type Config struct {
	SessionDuration     time.Duration
	MinSessionSpacing   time.Duration
	MaxPerPersonPerWeek int
}

// NewMatcher creates a new matcher instance
func NewMatcher(people []*Person, config *Config) *Matcher {
	return &Matcher{
		people: people,
		config: config,
	}
}

// FindMatches finds possible matches between reviewers
func (m *Matcher) FindMatches() ([]Match, error) {
	if len(m.people) < 2 {
		return nil, fmt.Errorf("need at least 2 people to create matches")
	}

	var matches []Match
	usedSlots := make(map[string][]calendar.TimeSlot)

	for i, person1 := range m.people {
		if !canReview(person1, usedSlots) {
			continue
		}

		for j := i + 1; j < len(m.people); j++ {
			person2 := m.people[j]
			if !canReview(person2, usedSlots) {
				continue
			}

			commonSlots := findCommonSlots(person1.FreeSlots, person2.FreeSlots)
			commonSkills := findCommonSkills(person1.Skills, person2.Skills)

			for _, slot := range commonSlots {
				if isValidSlot(slot, m.config.SessionDuration) {
					match := Match{
						Reviewer1:    person1,
						Reviewer2:    person2,
						TimeSlot:     slot,
						CommonSkills: commonSkills,
					}
					matches = append(matches, match)
					updateUsedSlots(usedSlots, person1.Email, slot)
					updateUsedSlots(usedSlots, person2.Email, slot)
					break
				}
			}
		}
	}

	return matches, nil
}

// canReview checks if a person can review based on their constraints
func canReview(person *Person, usedSlots map[string][]calendar.TimeSlot) bool {
	if person.MaxSessionsPerWeek == 0 {
		return false
	}

	used := len(usedSlots[person.Email])
	return used < person.MaxSessionsPerWeek
}

// findCommonSlots finds time slots that are available for both people
func findCommonSlots(slots1, slots2 []calendar.TimeSlot) []calendar.TimeSlot {
	var common []calendar.TimeSlot

	for _, slot1 := range slots1 {
		for _, slot2 := range slots2 {
			if slot1.Start.Before(slot2.End) && slot2.Start.Before(slot1.End) {
				start := slot1.Start
				if slot2.Start.After(start) {
					start = slot2.Start
				}
				end := slot1.End
				if slot2.End.Before(end) {
					end = slot2.End
				}
				common = append(common, calendar.TimeSlot{Start: start, End: end})
			}
		}
	}

	return common
}

// findCommonSkills finds skills that both people have
func findCommonSkills(skills1, skills2 []string) []string {
	skillMap := make(map[string]bool)
	for _, skill := range skills1 {
		skillMap[skill] = true
	}

	var common []string
	for _, skill := range skills2 {
		if skillMap[skill] {
			common = append(common, skill)
		}
	}

	return common
}

// isValidSlot checks if a time slot is valid for a session
func isValidSlot(slot calendar.TimeSlot, duration time.Duration) bool {
	return slot.End.Sub(slot.Start) >= duration
}

// updateUsedSlots updates the used slots for a person
func updateUsedSlots(usedSlots map[string][]calendar.TimeSlot, email string, slot calendar.TimeSlot) {
	usedSlots[email] = append(usedSlots[email], slot)
}
