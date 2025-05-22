package commands

import (
	"fmt"
	"matchmaker/libs/gcalendar"
	"matchmaker/libs/types"
	"matchmaker/libs/util"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func loadProblem(weekShift int, groupFile string) *types.Problem {
	groupPath := filepath.Join("groups", groupFile)
	people, err := types.LoadPersons(groupPath)
	util.PanicOnError(err, "Cannot load people file")
	util.LogInfo("People file loaded", map[string]interface{}{
		"count": len(people),
		"file":  groupPath,
	})

	cal, err := gcalendar.NewGCalendar()
	util.PanicOnError(err, "Cannot connect to Google Calendar")
	util.LogInfo("Connected to Google Calendar", nil)

	beginOfWeek := util.FirstDayOfISOWeek(weekShift)
	workRangesChan, err := util.GetWeekWorkRanges(beginOfWeek)
	if err != nil {
		panic(fmt.Errorf("failed to get work ranges: %w", err))
	}
	workRanges := util.ToSlice(workRangesChan)

	// Log work ranges in a human-readable format
	util.LogInfo("Work ranges for the week", map[string]interface{}{
		"start": beginOfWeek.Format("2006-01-02"),
		"end":   beginOfWeek.AddDate(0, 0, 4).Format("2006-01-02"), // Friday
		"ranges": func() []string {
			var ranges []string
			for _, r := range workRanges {
				ranges = append(ranges, fmt.Sprintf("%s %s-%s",
					r.Start.Format("Mon 2006-01-02"),
					r.Start.Format("15:04"),
					r.End.Format("15:04")))
			}
			return ranges
		}(),
	})

	busyTimes := cal.GetBusyTimesForPeople(people, workRanges)

	return &types.Problem{
		People:           people,
		WorkRanges:       workRanges,
		BusyTimes:        busyTimes,
		TargetCoverage:   1,
		MaxTotalCoverage: 2,
	}
}

func init() {
	prepareCmd.Flags().IntVarP(&weekShift, "week-shift", "w", 0, `define a week shift to plan for an upcoming 
week instead of next week. Default value (0) is next week, and 1 is the week after, etc.`)

	rootCmd.AddCommand(prepareCmd)
}

// One flag 'week-shift' can be set to plan for an upcoming week instead of next week
// Default = 0 (planning for next week)
// 1 = in two weeks, 2 = in 3 weeks, etc.
var weekShift int

var prepareCmd = &cobra.Command{
	Use:   "prepare [group-file]",
	Short: "Retrieve available slots for a group of people and parameters for the matching algorithm.",
	Long: `Compute work ranges for the target week, and check free slots for each potential
reviewer in a group and create an output file 'problem.yml'.

The group-file parameter specifies which group file to use from the groups directory.
You can create multiple group files (e.g., teams.yml, projects.yml) to manage different sets of people.
If no group file is specified, 'group.yml' will be used by default.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		groupFile := "group.yml"
		if len(args) > 0 {
			groupFile = args[0]
		}
		problem := loadProblem(weekShift, groupFile)
		yml, _ := problem.ToYaml()
		err := os.WriteFile("./problem.yml", yml, os.FileMode(0644))
		util.PanicOnError(err, "Can't yml problem file")
	},
}
