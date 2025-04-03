package types

import (
	"time"
)

// Range represents a time range
type Range struct {
	Start time.Time
	End   time.Time
}

// GetStart returns the start time of the range
func (r *Range) GetStart() time.Time {
	return r.Start
}

// GetEnd returns the end time of the range
func (r *Range) GetEnd() time.Time {
	return r.End
}

// Duration returns the duration of the range
func (r *Range) Duration() time.Duration {
	return r.End.Sub(r.Start)
}

// Contains returns true if the given time is within the range
func (r *Range) Contains(t time.Time) bool {
	return !t.Before(r.Start) && !t.After(r.End)
}

// Overlaps returns true if the given range overlaps with this range
func (r *Range) Overlaps(other *Range) bool {
	return r.End.After(other.Start) && other.End.After(r.Start)
}

// Before returns true if this range ends before the given time
func (r *Range) Before(t time.Time) bool {
	return r.End.Before(t)
}

// After returns true if this range starts after the given time
func (r *Range) After(t time.Time) bool {
	return r.Start.After(t)
}

// Minutes returns the duration of the range in minutes
func (r *Range) Minutes() float64 {
	return r.End.Sub(r.Start).Minutes()
}
