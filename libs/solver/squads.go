package solver

import (
	"matchmaker/libs/types"
	"matchmaker/util"
	"math/rand"
	"sort"
	"time"
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
			util.LogInfo("Busy range", map[string]interface{}{
				"start": squads[i].BusyRanges[j].Start.Format(time.Stamp),
				"end":   squads[i].BusyRanges[j].End.Format(time.Stamp),
			})
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
	return mergeRanges(busyRanges)
}

func mergeRanges(ranges []*types.Range) []*types.Range {
	if len(ranges) == 0 {
		return ranges
	}

	// Sort ranges by start time
	sort.Slice(ranges, func(i, j int) bool {
		return ranges[i].Start.Before(ranges[j].Start)
	})

	merged := []*types.Range{}
	current := ranges[0]

	for i := 1; i < len(ranges); i++ {
		next := ranges[i]
		if current.End.After(next.Start) || current.End.Equal(next.Start) {
			// Ranges overlap or are adjacent, merge them
			if next.End.After(current.End) {
				current.End = next.End
			}
		} else {
			// Ranges don't overlap, add current to merged and move to next
			merged = append(merged, current)
			current = next
		}
	}

	// Add the last range
	merged = append(merged, current)

	return merged
}

func haveIntersection(range1 *types.Range, range2 *types.Range) bool {
	return range1.End.After(range2.Start) && range2.End.After(range1.Start)
}

func (squads Squads) Print() {
	for _, squad := range squads {
		util.LogInfo("Squad", map[string]interface{}{
			"person1": squad.People[0].Email,
			"person2": squad.People[1].Email,
		})
	}
}
