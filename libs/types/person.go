package types

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Person represents a person who can participate in review sessions
type Person struct {
	Email                           string   `yaml:"email"`
	IsGoodReviewer                  bool     `yaml:"isgoodreviewer"`
	MaxSessionsPerWeek              int      `yaml:"maxsessionsperweek"`
	Skills                          []string `yaml:"skills"`
	isSessionCompatibleSessionCount int      `yaml:"-"`
	sessionCount                    int
}

// Validate checks if the person's data is valid
func (p *Person) Validate() error {
	if p.Email == "" {
		return fmt.Errorf("email is required")
	}
	if p.MaxSessionsPerWeek < 0 {
		return fmt.Errorf("maxSessionsPerWeek must be non-negative")
	}
	return nil
}

// CanParticipateInSession checks if the person can participate in a session
func (p *Person) CanParticipateInSession() bool {
	return p.MaxSessionsPerWeek > 0
}

// LoadPersons loads a list of persons from a YAML file
func LoadPersons(path string) ([]*Person, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var persons []*Person
	err = yaml.Unmarshal(data, &persons)
	if err != nil {
		return nil, err
	}

	// Validate all persons
	for _, person := range persons {
		if err := person.Validate(); err != nil {
			return nil, fmt.Errorf("invalid person %s: %w", person.Email, err)
		}
	}

	return persons, nil
}

func (p *Person) IncrementSessionCount() {
	p.sessionCount++
}

func (p *Person) GetSessionCount() int {
	return p.sessionCount
}

func (p *Person) ResetSessionCount() {
	p.sessionCount = 0
}
