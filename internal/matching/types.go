package matching

import (
	"time"

	"matchmaker/internal/calendar"
)

// Person represents a reviewer or reviewee
type Person struct {
	Email              string   `yaml:"email"`
	IsGoodReviewer     bool     `yaml:"isgoodreviewer"`
	MaxSessionsPerWeek int      `yaml:"maxsessionsperweek,omitempty"`
	Skills             []string `yaml:"skills,omitempty"`
	FreeSlots          []calendar.TimeSlot
}

// Match represents a matched pair of reviewers
type Match struct {
	Reviewer1    *Person           `yaml:"reviewer1"`
	Reviewer2    *Person           `yaml:"reviewer2"`
	TimeSlot     calendar.TimeSlot `yaml:"time_slot"`
	CommonSkills []string          `yaml:"common_skills"`
}

// Config represents the matcher configuration
type Config struct {
	SessionDuration     time.Duration `yaml:"duration"`
	MinSessionSpacing   time.Duration `yaml:"min_spacing"`
	MaxPerPersonPerWeek int           `yaml:"max_per_person_per_week"`
}
