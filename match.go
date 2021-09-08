package main

import (
	"flag"
	"github.com/transcovo/matchmaker/match"
	"github.com/transcovo/matchmaker/util"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"runtime/pprof"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	yml, err := ioutil.ReadFile("./problem.yml")
	util.PanicOnError(err, "Can't yml problem description")
	problem, err := match.LoadProblem(yml)
	solution := match.Solve(problem)

	planYml, _ := yaml.Marshal(solution)
	ioutil.WriteFile("./planning.yml", planYml, os.FileMode(0644))
}
