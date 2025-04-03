package squads

import (
	"matchmaker/libs/types"
)

// GenerateSquads creates all possible squads from the given people, considering their busy times
func GenerateSquads(people []*types.Person, busyTimes []*types.BusyTime) []*types.Squad {
	// Filter people into masters and disciples
	masters := filterPersons(people, true)
	disciples := filterPersons(people, false)

	// Generate all possible master-disciple pairs
	squads := make([]*types.Squad, 0)
	for _, master := range masters {
		for _, disciple := range disciples {
			// Skip if master and disciple are the same person
			if master == disciple {
				continue
			}

			// Create a squad with the master and disciple
			squad := &types.Squad{
				People: []*types.Person{master, disciple},
			}

			// Add busy ranges for this squad
			squad.BusyRanges = mergeBusyRanges(busyTimes, squad.People)

			squads = append(squads, squad)
		}
	}

	// Generate all possible master-master pairs
	for i, master1 := range masters {
		for j, master2 := range masters {
			// Skip if same person or already processed pair
			if i >= j {
				continue
			}

			// Create a squad with two masters
			squad := &types.Squad{
				People: []*types.Person{master1, master2},
			}

			// Add busy ranges for this squad
			squad.BusyRanges = mergeBusyRanges(busyTimes, squad.People)

			squads = append(squads, squad)
		}
	}

	return squads
}

// filterPersons filters people based on whether they are good reviewers
func filterPersons(people []*types.Person, isGoodReviewer bool) []*types.Person {
	filtered := make([]*types.Person, 0)
	for _, person := range people {
		if person.IsGoodReviewer == isGoodReviewer {
			filtered = append(filtered, person)
		}
	}
	return filtered
}

// mergeBusyRanges merges busy ranges for a group of people
func mergeBusyRanges(busyTimes []*types.BusyTime, people []*types.Person) []*types.Range {
	// Collect all busy ranges for the people
	busyRanges := make([]*types.Range, 0)
	for _, busyTime := range busyTimes {
		for _, person := range people {
			if busyTime.Person == person {
				busyRanges = append(busyRanges, busyTime.Range)
			}
		}
	}

	// Merge overlapping or adjacent ranges
	return types.MergeRanges(busyRanges)
}
