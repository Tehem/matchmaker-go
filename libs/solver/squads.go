package solver

// This file is intentionally left empty as the generateSquads function
// has been moved to the squads package.

import (
	"matchmaker/libs/types"
	"matchmaker/libs/util"
	"math/rand"
)

// Squads is a slice of Squad pointers
type Squads []*types.Squad

func generateSquads(people []*types.Person, busyTimes []*types.BusyTime) []*types.Squad {
	masters := filterPersons(people, true)
	disciples := filterPersons(people, false)

	squads := []*types.Squad{}
	for _, master := range masters {
		for _, disciple := range disciples {
			people := []*types.Person{master, disciple}
			squads = append(squads, &types.Squad{
				People:     people,
				BusyRanges: mergeBusyRanges(busyTimes, people),
			})
		}
	}

	for masterIndex1, master1 := range masters {
		for masterIndex2, master2 := range masters {
			if masterIndex1 < masterIndex2 {
				people := []*types.Person{master1, master2}
				squads = append(squads, &types.Squad{
					People:     people,
					BusyRanges: mergeBusyRanges(busyTimes, people),
				})
			}
		}
	}

	for i := range squads {
		j := rand.Intn(i + 1)
		squads[i], squads[j] = squads[j], squads[i]
	}

	util.LogInfo("Generated squads", map[string]interface{}{
		"count": len(squads),
	})
	for i := range squads {
		util.LogInfo("Squad", map[string]interface{}{
			"person1": squads[i].People[0].Email,
			"person2": squads[i].People[1].Email,
		})
		util.LogInfo("Busy ranges", nil)
		for j := range squads[i].BusyRanges {
			util.LogRange("Busy range", squads[i].BusyRanges[j])
		}
	}
	return squads
}

func filterPersons(people []*types.Person, isGoodReviewer bool) []*types.Person {
	result := []*types.Person{}
	for _, person := range people {
		if person.IsGoodReviewer == isGoodReviewer {
			result = append(result, person)
		}
	}
	return result
}

func mergeBusyRanges(busyTimes []*types.BusyTime, people []*types.Person) []*types.Range {
	busyRanges := []*types.Range{}
	for _, busyTime := range busyTimes {
		for _, person := range people {
			if busyTime.Person == person {
				busyRanges = append(busyRanges, busyTime.Range)
			}
		}
	}
	return types.MergeRanges(busyRanges)
}

func (squads Squads) Print() {
	for _, squad := range squads {
		util.LogInfo("Squad", map[string]interface{}{
			"person1": squad.People[0].Email,
			"person2": squad.People[1].Email,
		})
	}
}
