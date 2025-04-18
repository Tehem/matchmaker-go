package commands

import (
	"matchmaker/libs/gcalendar"
	"matchmaker/libs/solver"
	"matchmaker/libs/types"
	"matchmaker/libs/util"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func init() {
	rootCmd.AddCommand(weeklyMatchCmd)
}

var weeklyMatchCmd = &cobra.Command{
	Use:   "weekly-match [group-file]",
	Short: "Create random pairs of people with no common skills and schedule sessions across weeks.",
	Long: `Create random pairs of people from a group file, ensuring that paired people have no common skills.
Then schedule pairing sessions for each tuple across consecutive weeks. The output is a 'weekly-planning.yml' file.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		groupFile := "group.yml"
		if len(args) > 0 {
			groupFile = args[0]
		}

		util.LogInfo("Starting weekly match process", map[string]interface{}{
			"groupFile": groupFile,
		})

		// Load and filter available people
		availablePeople := loadAndFilterPeople(groupFile)

		// Create random pairs
		tuples := createRandomPairs(availablePeople)

		// Get Google Calendar service
		cal, err := gcalendar.NewGCalendar()
		util.PanicOnError(err, "Can't get Google Calendar client")
		util.LogInfo("Connected to Google Calendar", nil)

		// Process tuples and create sessions
		combinedSolution, allUnmatchedTuples, allUnmatchedPeople := processTuplesAndCreateSessions(tuples, cal)

		// Output results
		outputResults(combinedSolution, tuples, allUnmatchedTuples, allUnmatchedPeople)
	},
}

func loadAndFilterPeople(groupFile string) []*types.Person {
	groupPath := filepath.Join("groups", groupFile)
	people, err := types.LoadPersons(groupPath)
	util.PanicOnError(err, "Cannot load people file")
	util.LogInfo("People file loaded", map[string]interface{}{
		"count": len(people),
		"file":  groupPath,
	})

	// Filter out people with maxSessionsPerWeek = 0
	availablePeople := make([]*types.Person, 0)
	for _, person := range people {
		if person.MaxSessionsPerWeek > 0 {
			availablePeople = append(availablePeople, person)
		}
	}
	util.LogInfo("Filtered available people", map[string]interface{}{
		"totalPeople":     len(people),
		"availablePeople": len(availablePeople),
	})

	return availablePeople
}

func createRandomPairs(availablePeople []*types.Person) types.Tuples {
	tuples := types.Tuples{
		Pairs:          make([]types.Tuple, 0),
		UnpairedPeople: make([]*types.Person, 0),
	}

	// Create a local random generator
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Shuffle the people array
	util.LogInfo("Shuffling people for random pairing", nil)
	r.Shuffle(len(availablePeople), func(i, j int) {
		availablePeople[i], availablePeople[j] = availablePeople[j], availablePeople[i]
	})

	// Create pairs with no common skills
	used := make(map[*types.Person]bool)
	for i, person1 := range availablePeople {
		if used[person1] {
			continue
		}

		// Find a person with no common skills
		for j := i + 1; j < len(availablePeople); j++ {
			person2 := availablePeople[j]
			if used[person2] {
				continue
			}

			// Check if they have no common skills
			commonSkills := util.Intersection(person1.Skills, person2.Skills)
			if len(commonSkills) == 0 {
				tuples.Pairs = append(tuples.Pairs, types.Tuple{
					Person1: person1,
					Person2: person2,
				})
				used[person1] = true
				used[person2] = true
				util.LogInfo("Created pair", map[string]interface{}{
					"person1": person1.Email,
					"person2": person2.Email,
				})
				break
			}
		}
		if !used[person1] {
			tuples.UnpairedPeople = append(tuples.UnpairedPeople, person1)
			util.LogInfo("Person could not be paired", map[string]interface{}{
				"email": person1.Email,
			})
		}
	}

	return tuples
}

func processTuplesAndCreateSessions(tuples types.Tuples, cal *gcalendar.GCalendar) (*types.Solution, []types.Tuple, []*types.Person) {
	combinedSolution := &types.Solution{
		Sessions: make([]*types.ReviewSession, 0),
	}

	allUnmatchedTuples := make([]types.Tuple, 0)
	allUnmatchedPeople := make([]*types.Person, 0)
	allUnmatchedPeople = append(allUnmatchedPeople, tuples.UnpairedPeople...)

	for i, tuple := range tuples.Pairs {
		weekShift := i
		util.LogInfo("Processing tuple for week", map[string]interface{}{
			"tupleIndex": i,
			"weekShift":  weekShift,
			"person1":    tuple.Person1.Email,
			"person2":    tuple.Person2.Email,
		})

		beginOfWeek := util.FirstDayOfISOWeek(weekShift)
		workRangesChan, err := util.GetWeekWorkRanges(beginOfWeek)
		util.PanicOnError(err, "Failed to get work ranges")
		workRanges := util.ToSlice(workRangesChan)
		busyTimes := getBusyTimesForTuple(tuple, workRanges, cal)

		// Log problem details
		util.LogInfo("Problem details", map[string]interface{}{
			"people": []string{tuple.Person1.Email, tuple.Person2.Email},
			"workRanges": []string{
				workRanges[0].Start.Format(time.RFC3339),
				workRanges[0].End.Format(time.RFC3339),
			},
			"numBusyTimes": len(busyTimes),
		})

		// Log busy time details
		for i, busyTime := range busyTimes {
			util.LogInfo("Busy time", map[string]interface{}{
				"index":  i,
				"start":  busyTime.Range.Start.Format(time.RFC3339),
				"end":    busyTime.Range.End.Format(time.RFC3339),
				"person": busyTime.Person.Email,
			})
		}

		session := solver.FindSessionForTuple(tuple, workRanges, busyTimes)

		if session != nil {
			combinedSolution.Sessions = append(combinedSolution.Sessions, session)
			util.LogInfo("Added session for tuple", map[string]interface{}{
				"tupleIndex":   i,
				"weekShift":    weekShift,
				"sessionStart": session.Start().Format(time.RFC3339),
				"sessionEnd":   session.End().Format(time.RFC3339),
			})
		} else {
			util.LogInfo("No session found for tuple", map[string]interface{}{
				"tupleIndex": i,
				"weekShift":  weekShift,
			})

			allUnmatchedTuples = append(allUnmatchedTuples, tuple)
			allUnmatchedPeople = append(allUnmatchedPeople, tuple.Person1)
			allUnmatchedPeople = append(allUnmatchedPeople, tuple.Person2)
		}
	}

	return combinedSolution, allUnmatchedTuples, allUnmatchedPeople
}

func getBusyTimesForTuple(tuple types.Tuple, workRanges []*types.Range, cal *gcalendar.GCalendar) []*types.BusyTime {
	tuplePeople := []*types.Person{tuple.Person1, tuple.Person2}
	return cal.GetBusyTimesForPeople(tuplePeople, workRanges)
}

func outputResults(combinedSolution *types.Solution, tuples types.Tuples, allUnmatchedTuples []types.Tuple, allUnmatchedPeople []*types.Person) {
	// Print summary of all sessions
	util.LogInfo("Weekly match process completed", map[string]interface{}{
		"totalPairs":      len(tuples.Pairs),
		"unpairedPeople":  len(tuples.UnpairedPeople),
		"totalSessions":   len(combinedSolution.Sessions),
		"unmatchedTuples": len(allUnmatchedTuples),
		"outputFile":      "./weekly-planning.yml",
	})

	// Print all sessions
	util.LogInfo("Generated weekly sessions", map[string]interface{}{
		"count": len(combinedSolution.Sessions),
	})
	for _, session := range combinedSolution.Sessions {
		util.LogSession("Weekly session", session)
	}

	// Print unmatched tuples
	if len(allUnmatchedTuples) > 0 {
		util.LogInfo("Unmatched tuples", map[string]interface{}{
			"count": len(allUnmatchedTuples),
		})
		for _, tuple := range allUnmatchedTuples {
			util.LogInfo("Unmatched tuple", map[string]interface{}{
				"person1": tuple.Person1.Email,
				"person2": tuple.Person2.Email,
			})
		}
	}

	// Print unpaired people
	if len(allUnmatchedPeople) > 0 {
		util.LogInfo("Unpaired people", map[string]interface{}{
			"count": len(allUnmatchedPeople),
		})
		for _, person := range allUnmatchedPeople {
			util.LogInfo("Unpaired person", map[string]interface{}{
				"email": person.Email,
			})
		}
	}

	// Output the combined solution to a file
	yml, err := yaml.Marshal(combinedSolution)
	util.PanicOnError(err, "Can't marshal solution")
	writeErr := os.WriteFile("./weekly-planning.yml", yml, os.FileMode(0644))
	util.PanicOnError(writeErr, "Can't write weekly planning result")
}
