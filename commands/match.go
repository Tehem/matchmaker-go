package commands

import (
	"flag"
	"fmt"
	"matchmaker/libs/solver"
	"matchmaker/libs/types"
	"matchmaker/libs/util"
	"os"
	"runtime/pprof"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var cpuprofile string

func init() {
	matchCmd.Flags().StringVarP(&cpuprofile, "cpuprofile", "c", "", `Target file to write cpu profile output.`)

	rootCmd.AddCommand(matchCmd)
}

func printMatchSummary(solution *solver.Solution) {
	// Print summary message to standard output
	fmt.Printf("\n✅ Planning file generated successfully!\n")
	fmt.Printf("📅 Week: %s to %s\n",
		solution.Sessions[0].Range.Start.Format("2006-01-02"),
		solution.Sessions[len(solution.Sessions)-1].Range.End.Format("2006-01-02"))
	fmt.Printf("👥 Sessions: %d\n", len(solution.Sessions))

	// Group sessions by day
	sessionsByDay := make(map[string][]*types.ReviewSession)
	for _, session := range solution.Sessions {
		day := session.Range.Start.Format("2006-01-02")
		sessionsByDay[day] = append(sessionsByDay[day], session)
	}

	// Sort days to ensure consistent output
	days := make([]string, 0, len(sessionsByDay))
	for day := range sessionsByDay {
		days = append(days, day)
	}
	sort.Strings(days)

	// Print sessions for each day
	for _, day := range days {
		date, _ := time.Parse("2006-01-02", day)
		weekday := util.FrenchWeekdays[date.Weekday().String()]
		month := util.FrenchMonths[date.Month().String()]
		fmt.Printf("\n📅 %s %d %s:\n", weekday, date.Day(), month)

		// Sort sessions by start time
		sessions := sessionsByDay[day]
		sort.Slice(sessions, func(i, j int) bool {
			return sessions[i].Range.Start.Before(sessions[j].Range.Start)
		})

		for _, session := range sessions {
			// Get reviewer names
			reviewers := make([]string, len(session.Reviewers.People))
			for i, person := range session.Reviewers.People {
				reviewers[i] = person.Email
			}

			fmt.Printf("   ⏰ %s - %s: %s\n",
				session.Range.Start.Format("15:04"),
				session.Range.End.Format("15:04"),
				strings.Join(reviewers, " & "))
		}
	}

	fmt.Printf("\n📝 Output file: planning.yml\n\n")
}

var matchCmd = &cobra.Command{
	Use:   "match",
	Short: "Match participants in sessions and create a planning proposal.",
	Long: `Match reviewers together in review slots for the target week. The output is a 'planning.yml' 
file with reviewers tuples and planned slots.`,
	Run: func(cmd *cobra.Command, args []string) {
		flag.Parse()
		if cpuprofile != "" {
			f, err := os.Create(cpuprofile)
			if err != nil {
				util.LogError(err, "Failed to create CPU profile file")
				return
			}
			_ = pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
		}

		yml, err := os.ReadFile("./problem.yml")
		util.PanicOnError(err, "Can't read problem description")
		problem, err := types.LoadProblem(yml)
		util.PanicOnError(err, "Can't load problem")
		solution := solver.Solve(problem)

		planYml, err := yaml.Marshal(solution)
		util.PanicOnError(err, "Can't marshal solution")
		writeErr := os.WriteFile("./planning.yml", planYml, os.FileMode(0644))
		util.PanicOnError(writeErr, "Can't write planning result")

		printMatchSummary(solution)
	},
}
