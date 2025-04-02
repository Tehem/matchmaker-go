package commands

import (
	"matchmaker/libs"
	"matchmaker/util"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
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
	Short: "Create random pairs of people with no common skills.",
	Long: `Create random pairs of people from a group file, ensuring that paired people have no common skills.
The output is a 'tuples.yml' file containing the pairs and any unpaired people.`,
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

		// Output the tuples to a file
		yml, err := yaml.Marshal(tuples)
		util.PanicOnError(err, "Can't marshal tuples")
		writeErr := os.WriteFile("./tuples.yml", yml, os.FileMode(0644))
		util.PanicOnError(writeErr, "Can't write tuples result")

		util.LogInfo("Weekly match process completed", map[string]interface{}{
			"totalPairs":     len(tuples.Pairs),
			"unpairedPeople": len(tuples.UnpairedPeople),
			"outputFile":     "./tuples.yml",
		})
	},
}
