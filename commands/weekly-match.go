package commands

import (
	"matchmaker/libs"
	"matchmaker/libs/gcalendar"
	"matchmaker/util"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/api/calendar/v3"
	"gopkg.in/yaml.v3"
)

type Tuple struct {
	Person1 *libs.Person `yaml:"person1"`
	Person2 *libs.Person `yaml:"person2"`
}

type Tuples struct {
	Pairs          []Tuple        `yaml:"pairs"`
	UnpairedPeople []*libs.Person `yaml:"unpairedPeople"`
}

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

		// Load people from group file
		groupPath := filepath.Join("groups", groupFile)
		people, err := libs.LoadPersons(groupPath)
		util.PanicOnError(err, "Cannot load people file")
		util.LogInfo("People file loaded", map[string]interface{}{
			"count": len(people),
			"file":  groupPath,
		})

		// Filter out people with maxSessionsPerWeek = 0
		availablePeople := make([]*libs.Person, 0)
		for _, person := range people {
			if person.MaxSessionsPerWeek > 0 {
				availablePeople = append(availablePeople, person)
			}
		}
		util.LogInfo("Filtered available people", map[string]interface{}{
			"totalPeople":     len(people),
			"availablePeople": len(availablePeople),
		})

		// Create random pairs
		tuples := Tuples{
			Pairs:          make([]Tuple, 0),
			UnpairedPeople: make([]*libs.Person, 0),
		}

		// Create a local random generator
		r := rand.New(rand.NewSource(time.Now().UnixNano()))

		// Shuffle the people array
		util.LogInfo("Shuffling people for random pairing", nil)
		r.Shuffle(len(availablePeople), func(i, j int) {
			availablePeople[i], availablePeople[j] = availablePeople[j], availablePeople[i]
		})

		// Create pairs with no common skills
		used := make(map[*libs.Person]bool)
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
					tuples.Pairs = append(tuples.Pairs, Tuple{
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

		// Get Google Calendar service
		cal, err := gcalendar.GetGoogleCalendarService()
		util.PanicOnError(err, "Can't get Google Calendar client")
		util.LogInfo("Connected to Google Calendar", nil)

		// Create a combined solution for all tuples
		combinedSolution := &libs.Solution{
			Sessions: make([]*libs.ReviewSession, 0),
		}

		// Track all unmatched tuples and people
		allUnmatchedTuples := make([]libs.Tuple, 0)
		allUnmatchedPeople := make([]*libs.Person, 0)

		// Initialize allUnmatchedPeople with tuples.UnpairedPeople
		allUnmatchedPeople = append(allUnmatchedPeople, tuples.UnpairedPeople...)

		// For each tuple, create a session for a different week
		for i, tuple := range tuples.Pairs {
			weekShift := i // First tuple uses current week, second uses next week, etc.

			util.LogInfo("Processing tuple for week", map[string]interface{}{
				"tupleIndex": i,
				"weekShift":  weekShift,
				"person1":    tuple.Person1.Email,
				"person2":    tuple.Person2.Email,
			})

			// Get the beginning of the target week
			beginOfWeek := FirstDayOfISOWeek(weekShift)
			util.LogInfo("Planning for week", map[string]interface{}{
				"weekFirstDay": beginOfWeek,
			})

			// Get work ranges for the week
			workRanges := ToSlice(GetWeekWorkRanges(beginOfWeek))

			// Get busy times for both people in the tuple
			busyTimes := []*libs.BusyTime{}
			tuplePeople := []*libs.Person{tuple.Person1, tuple.Person2}

			for _, person := range tuplePeople {
				util.LogInfo("Loading busy detail", map[string]interface{}{
					"person": person.Email,
				})

				for _, workRange := range workRanges {
					util.LogInfo("Loading busy detail on range", map[string]interface{}{
						"start": workRange.Start,
						"end":   workRange.End,
					})

					result, err := cal.Freebusy.Query(&calendar.FreeBusyRequest{
						TimeMin: gcalendar.FormatTime(workRange.Start),
						TimeMax: gcalendar.FormatTime(workRange.End),
						Items: []*calendar.FreeBusyRequestItem{
							{
								Id: person.Email,
							},
						},
					}).Do()
					util.PanicOnError(err, "Can't retrieve free/busy data for "+person.Email)

					busyTimePeriods := result.Calendars[person.Email].Busy
					util.LogInfo("Person busy times", map[string]interface{}{
						"person": person.Email,
					})

					for _, busyTimePeriod := range busyTimePeriods {
						util.LogInfo("Busy time period", map[string]interface{}{
							"start": busyTimePeriod.Start,
							"end":   busyTimePeriod.End,
						})
						busyTimes = append(busyTimes, &libs.BusyTime{
							Person: person,
							Range: &libs.Range{
								Start: parseTime(busyTimePeriod.Start),
								End:   parseTime(busyTimePeriod.End),
							},
						})
					}
				}
			}

			// Create a problem for this tuple
			problem := &libs.Problem{
				People:         tuplePeople,
				WorkRanges:     workRanges,
				BusyTimes:      busyTimes,
				TargetCoverage: 0, // We don't need to cover all time slots
			}

			// Use our custom solver to find a session for this tuple
			solution := libs.WeeklySolve(problem)

			// Add the sessions to our combined solution
			if len(solution.Solution.Sessions) > 0 {
				// Take the first session (we only need one per tuple)
				combinedSolution.Sessions = append(combinedSolution.Sessions, solution.Solution.Sessions[0])
				util.LogInfo("Added session for tuple", map[string]interface{}{
					"tupleIndex":   i,
					"weekShift":    weekShift,
					"sessionStart": solution.Solution.Sessions[0].Start().Format(time.RFC3339),
					"sessionEnd":   solution.Solution.Sessions[0].End().Format(time.RFC3339),
				})
			} else {
				util.LogInfo("No session found for tuple", map[string]interface{}{
					"tupleIndex": i,
					"weekShift":  weekShift,
				})

				// Add the tuple to unmatched tuples if no session was found
				if len(solution.UnmatchedTuples) > 0 {
					allUnmatchedTuples = append(allUnmatchedTuples, solution.UnmatchedTuples[0])
					allUnmatchedPeople = append(allUnmatchedPeople, solution.UnmatchedTuples[0].Person1)
					allUnmatchedPeople = append(allUnmatchedPeople, solution.UnmatchedTuples[0].Person2)
				}
			}
		}

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
			util.LogInfo("Weekly session", map[string]interface{}{
				"person1": session.Reviewers.People[0].Email,
				"person2": session.Reviewers.People[1].Email,
				"start":   session.Range.Start.Format(time.RFC3339),
				"end":     session.Range.End.Format(time.RFC3339),
			})
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
	},
}
