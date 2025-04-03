package types

import (
	"testing"
	"time"
)

func TestRangeMethods(t *testing.T) {
	// Create a fixed date for testing
	start := time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC)
	end := time.Date(2024, 4, 1, 17, 0, 0, 0, time.UTC)
	r := &Range{Start: start, End: end}

	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "GetStart",
			test: func(t *testing.T) {
				if got := r.GetStart(); !got.Equal(start) {
					t.Errorf("GetStart() = %v, want %v", got, start)
				}
			},
		},
		{
			name: "GetEnd",
			test: func(t *testing.T) {
				if got := r.GetEnd(); !got.Equal(end) {
					t.Errorf("GetEnd() = %v, want %v", got, end)
				}
			},
		},
		{
			name: "Duration",
			test: func(t *testing.T) {
				want := 8 * time.Hour
				if got := r.Duration(); got != want {
					t.Errorf("Duration() = %v, want %v", got, want)
				}
			},
		},
		{
			name: "Contains",
			test: func(t *testing.T) {
				tests := []struct {
					time time.Time
					want bool
				}{
					{start.Add(-time.Hour), false},
					{start, true},
					{start.Add(4 * time.Hour), true},
					{end, true},
					{end.Add(time.Hour), false},
				}
				for _, tt := range tests {
					if got := r.Contains(tt.time); got != tt.want {
						t.Errorf("Contains(%v) = %v, want %v", tt.time, got, tt.want)
					}
				}
			},
		},
		{
			name: "Overlaps",
			test: func(t *testing.T) {
				tests := []struct {
					other *Range
					want  bool
				}{
					{
						&Range{Start: start.Add(-2 * time.Hour), End: start.Add(-time.Hour)},
						false,
					},
					{
						&Range{Start: start.Add(-time.Hour), End: start.Add(time.Hour)},
						true,
					},
					{
						&Range{Start: start.Add(4 * time.Hour), End: end.Add(time.Hour)},
						true,
					},
					{
						&Range{Start: end.Add(time.Hour), End: end.Add(2 * time.Hour)},
						false,
					},
				}
				for _, tt := range tests {
					if got := r.Overlaps(tt.other); got != tt.want {
						t.Errorf("Overlaps(%v-%v) = %v, want %v", tt.other.Start, tt.other.End, got, tt.want)
					}
				}
			},
		},
		{
			name: "Before",
			test: func(t *testing.T) {
				tests := []struct {
					time time.Time
					want bool
				}{
					{start.Add(-time.Hour), false},
					{start, false},
					{end, false},
					{end.Add(time.Hour), true},
				}
				for _, tt := range tests {
					if got := r.Before(tt.time); got != tt.want {
						t.Errorf("Before(%v) = %v, want %v", tt.time, got, tt.want)
					}
				}
			},
		},
		{
			name: "After",
			test: func(t *testing.T) {
				tests := []struct {
					time time.Time
					want bool
				}{
					{start.Add(-time.Hour), true},
					{start, false},
					{end, false},
					{end.Add(time.Hour), false},
				}
				for _, tt := range tests {
					if got := r.After(tt.time); got != tt.want {
						t.Errorf("After(%v) = %v, want %v", tt.time, got, tt.want)
					}
				}
			},
		},
		{
			name: "Pad",
			test: func(t *testing.T) {
				padding := 30 * time.Minute
				got := r.Pad(padding)
				wantStart := start.Add(-padding)
				wantEnd := end.Add(padding)
				if !got.Start.Equal(wantStart) || !got.End.Equal(wantEnd) {
					t.Errorf("Pad(%v) = %v-%v, want %v-%v", padding, got.Start, got.End, wantStart, wantEnd)
				}
			},
		},
		{
			name: "Minutes",
			test: func(t *testing.T) {
				want := 480.0 // 8 hours in minutes
				if got := r.Minutes(); got != want {
					t.Errorf("Minutes() = %v, want %v", got, want)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}
