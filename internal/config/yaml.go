package config

import (
	"fmt"
	"time"

	"matchmaker/internal/matching"

	"gopkg.in/yaml.v3"
)

// Problem represents the problem configuration
type Problem struct {
	TargetWeek time.Time          `yaml:"targetWeek"`
	People     []*matching.Person `yaml:"people"`
}

// Planning represents the planning configuration
type Planning struct {
	Matches []matching.Match `yaml:"matches"`
}

// LoadPeople loads the people configuration from YAML
func LoadPeople(filename string) ([]*matching.Person, error) {
	data, err := fs.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read people file: %w", err)
	}

	var people []*matching.Person
	if err := yaml.Unmarshal(data, &people); err != nil {
		return nil, fmt.Errorf("failed to unmarshal people: %w", err)
	}

	return people, nil
}

// SaveProblem saves the problem configuration to YAML
func SaveProblem(people []*matching.Person, targetWeek time.Time, filename string) error {
	problem := Problem{
		TargetWeek: targetWeek,
		People:     people,
	}

	data, err := yaml.Marshal(problem)
	if err != nil {
		return fmt.Errorf("failed to marshal problem: %w", err)
	}

	if err := fs.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write problem file: %w", err)
	}

	return nil
}

// LoadProblem loads the problem configuration from YAML
func LoadProblem(filename string) ([]*matching.Person, error) {
	data, err := fs.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read problem file: %w", err)
	}

	var problem Problem
	if err := yaml.Unmarshal(data, &problem); err != nil {
		return nil, fmt.Errorf("failed to unmarshal problem: %w", err)
	}

	return problem.People, nil
}

// SavePlanning saves the planning configuration to YAML
func SavePlanning(matches []matching.Match, filename string) error {
	planning := Planning{
		Matches: matches,
	}

	data, err := yaml.Marshal(planning)
	if err != nil {
		return fmt.Errorf("failed to marshal planning: %w", err)
	}

	if err := fs.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write planning file: %w", err)
	}

	return nil
}

// LoadPlanning loads the planning configuration from YAML
func LoadPlanning(filename string) ([]matching.Match, error) {
	data, err := fs.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read planning file: %w", err)
	}

	var planning Planning
	if err := yaml.Unmarshal(data, &planning); err != nil {
		return nil, fmt.Errorf("failed to unmarshal planning: %w", err)
	}

	return planning.Matches, nil
}
