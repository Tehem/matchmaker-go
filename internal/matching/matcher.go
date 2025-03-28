package matching

import (
	"fmt"
	"sort"
	"time"

	"matchmaker/internal/calendar"
)

// Matcher handles the matching logic
type Matcher struct {
	people []*Person
	config *Config
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

	// Sort people by MaxSessionsPerWeek (higher first)
	sortedPeople := make([]*Person, len(m.people))
	copy(sortedPeople, m.people)
	sort.Slice(sortedPeople, func(i, j int) bool {
		return sortedPeople[i].MaxSessionsPerWeek > sortedPeople[j].MaxSessionsPerWeek
	})

	var matches []Match
	usedSlots := make(map[string][]calendar.TimeSlot)
	matched := make(map[string]bool) // Track who has been matched

	for i, person1 := range sortedPeople {
		if matched[person1.Email] || !canReview(person1, usedSlots) {
			continue
		}

		for j := i + 1; j < len(sortedPeople); j++ {
			person2 := sortedPeople[j]
			if matched[person2.Email] || !canReview(person2, usedSlots) {
				continue
			}

			commonSlots := findCommonSlots(person1.FreeSlots, person2.FreeSlots)
			commonSkills := findCommonSkills(person1.Skills, person2.Skills)

			// Skip if no common skills
			if len(commonSkills) == 0 {
				continue
			}

			// Try to find a valid slot
			for _, slot := range commonSlots {
				if isValidSlot(slot, m.config.SessionDuration) {
					match := Match{
						TimeSlot:     slot,
						CommonSkills: commonSkills,
					}

					// Sort reviewers alphabetically by email
					if person1.Email < person2.Email {
						match.Reviewer1 = person1
						match.Reviewer2 = person2
					} else {
						match.Reviewer1 = person2
						match.Reviewer2 = person1
					}

					matches = append(matches, match)
					updateUsedSlots(usedSlots, person1.Email, slot)
					updateUsedSlots(usedSlots, person2.Email, slot)
					matched[person1.Email] = true
					matched[person2.Email] = true
					break
				}
			}

			// If we matched person1, move to the next one
			if matched[person1.Email] {
				break
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
