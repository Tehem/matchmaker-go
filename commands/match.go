package commands

import (
	"flag"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"matchmaker/libs"
	"matchmaker/util"
	"os"
	"runtime/pprof"
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
				log.Fatal(err)
			}
			_ = pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
		}

		yml, err := ioutil.ReadFile("./problem.yml")
		util.PanicOnError(err, "Can't yml problem description")
		problem, err := libs.LoadProblem(yml)
		solution := libs.Solve(problem)

		planYml, _ := yaml.Marshal(solution)
		writeErr := ioutil.WriteFile("./planning.yml", planYml, os.FileMode(0644))
		util.PanicOnError(writeErr, "Can't yml planning result")
	},
}
