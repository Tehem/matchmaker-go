package commands

import (
	"flag"
	"matchmaker/libs/solver"
	"matchmaker/libs/types"
	"matchmaker/libs/util"
	"os"
	"runtime/pprof"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var cpuprofile string

func init() {
	matchCmd.Flags().StringVarP(&cpuprofile, "cpuprofile", "c", "", `Target file to write cpu profile output.`)

	rootCmd.AddCommand(matchCmd)
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
	},
}
