package types

import (
	"gopkg.in/yaml.v3"
)

// Solution represents a solution to a problem
type Solution struct {
	Sessions []*ReviewSession
}

// Tuples represents a collection of pairs and unpaired people
type Tuples struct {
	Pairs          []Tuple
	UnpairedPeople []*Person
}

// Tuple represents a pair of people
type Tuple struct {
	Person1 *Person
	Person2 *Person
}

type BusyTime struct {
	Person *Person
	Range  *Range
}

type Problem struct {
	People           []*Person
	WorkRanges       []*Range
	BusyTimes        []*BusyTime
	TargetCoverage   int
	MaxTotalCoverage int
}

type SerializedBusyTime struct {
	Email string
	Range *Range
}

type SerializedProblem struct {
	People         []*Person
	WorkRanges     []*Range
	BusyTimes      []*SerializedBusyTime
	TargetCoverage int
}

func (problem *Problem) ToYaml() ([]byte, error) {
	serializedBusyTimes := make([]*SerializedBusyTime, len(problem.BusyTimes))
	for i, busyTime := range problem.BusyTimes {
		serializedBusyTimes[i] = &SerializedBusyTime{
			Email: busyTime.Person.Email,
			Range: busyTime.Range,
		}
	}

	serializedProblem := SerializedProblem{
		People:         problem.People,
		WorkRanges:     problem.WorkRanges,
		BusyTimes:      serializedBusyTimes,
		TargetCoverage: problem.TargetCoverage,
	}
	data, err := yaml.Marshal(serializedProblem)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func LoadProblem(yml []byte) (*Problem, error) {
	var serializedProblem SerializedProblem
	err := yaml.Unmarshal(yml, &serializedProblem)
	if err != nil {
		return nil, err
	}

	personsByEmail := map[string]*Person{}
	for _, person := range serializedProblem.People {
		personsByEmail[person.Email] = person
	}

	busyTimes := make([]*BusyTime, len(serializedProblem.BusyTimes))
	for i, serializedBusyTime := range serializedProblem.BusyTimes {
		busyTimes[i] = &BusyTime{
			Person: personsByEmail[serializedBusyTime.Email],
			Range:  serializedBusyTime.Range,
		}
	}

	return &Problem{
		People:           serializedProblem.People,
		WorkRanges:       serializedProblem.WorkRanges,
		BusyTimes:        busyTimes,
		TargetCoverage:   serializedProblem.TargetCoverage,
		MaxTotalCoverage: 8,
	}, nil
}
