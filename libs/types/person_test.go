package types

import (
	"os"
	"testing"
)

func TestPersonValidation(t *testing.T) {
	tests := []struct {
		name    string
		person  *Person
		wantErr bool
	}{
		{
			name: "valid person",
			person: &Person{
				Email:              "test@example.com",
				MaxSessionsPerWeek: 2,
			},
			wantErr: false,
		},
		{
			name: "empty email",
			person: &Person{
				Email:              "",
				MaxSessionsPerWeek: 2,
			},
			wantErr: true,
		},
		{
			name: "negative max sessions",
			person: &Person{
				Email:              "test@example.com",
				MaxSessionsPerWeek: -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.person.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPersonSessionParticipation(t *testing.T) {
	tests := []struct {
		name               string
		maxSessionsPerWeek int
		want               bool
	}{
		{
			name:               "can participate",
			maxSessionsPerWeek: 2,
			want:               true,
		},
		{
			name:               "cannot participate",
			maxSessionsPerWeek: 0,
			want:               false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Person{
				Email:              "test@example.com",
				MaxSessionsPerWeek: tt.maxSessionsPerWeek,
			}
			if got := p.CanParticipateInSession(); got != tt.want {
				t.Errorf("CanParticipateInSession() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPersonSessionCount(t *testing.T) {
	p := &Person{
		Email:              "test@example.com",
		MaxSessionsPerWeek: 2,
	}

	// Test initial count
	if got := p.GetSessionCount(); got != 0 {
		t.Errorf("GetSessionCount() = %v, want 0", got)
	}

	// Test increment
	p.IncrementSessionCount()
	if got := p.GetSessionCount(); got != 1 {
		t.Errorf("GetSessionCount() after increment = %v, want 1", got)
	}

	// Test reset
	p.ResetSessionCount()
	if got := p.GetSessionCount(); got != 0 {
		t.Errorf("GetSessionCount() after reset = %v, want 0", got)
	}
}

func TestLoadPersons(t *testing.T) {
	// Create a temporary YAML file
	content := `
- email: person1@example.com
  isgoodreviewer: true
  maxsessionsperweek: 2
  skills:
    - go
    - python
- email: person2@example.com
  isgoodreviewer: false
  maxsessionsperweek: 1
  skills:
    - java
`

	tmpFile, err := os.CreateTemp("", "persons-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temporary file: %v", err)
	}

	// Test loading persons
	persons, err := LoadPersons(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadPersons() error = %v", err)
	}

	if len(persons) != 2 {
		t.Errorf("LoadPersons() returned %d persons, want 2", len(persons))
	}

	// Verify first person
	person1 := persons[0]
	if person1.Email != "person1@example.com" {
		t.Errorf("Person1 email = %v, want person1@example.com", person1.Email)
	}
	if !person1.IsGoodReviewer {
		t.Error("Person1 IsGoodReviewer = false, want true")
	}
	if person1.MaxSessionsPerWeek != 2 {
		t.Errorf("Person1 MaxSessionsPerWeek = %v, want 2", person1.MaxSessionsPerWeek)
	}
	if len(person1.Skills) != 2 || person1.Skills[0] != "go" || person1.Skills[1] != "python" {
		t.Errorf("Person1 Skills = %v, want [go python]", person1.Skills)
	}

	// Verify second person
	person2 := persons[1]
	if person2.Email != "person2@example.com" {
		t.Errorf("Person2 email = %v, want person2@example.com", person2.Email)
	}
	if person2.IsGoodReviewer {
		t.Error("Person2 IsGoodReviewer = true, want false")
	}
	if person2.MaxSessionsPerWeek != 1 {
		t.Errorf("Person2 MaxSessionsPerWeek = %v, want 1", person2.MaxSessionsPerWeek)
	}
	if len(person2.Skills) != 1 || person2.Skills[0] != "java" {
		t.Errorf("Person2 Skills = %v, want [java]", person2.Skills)
	}
}
