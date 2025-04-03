package types

import (
	"matchmaker/libs/config"
	"matchmaker/libs/testutils"
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

func TestGenerateTimeRanges(t *testing.T) {
	// Create a config mock
	configMock := testutils.NewConfigMock()
	configMock.SetupWorkHours()
	defer configMock.Restore()

	// Create a work range (Monday 9:00-17:00)
	start := time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC)
	end := time.Date(2024, 4, 1, 17, 0, 0, 0, time.UTC)
	workRange := &Range{Start: start, End: end}

	// Generate time ranges with a timeout
	done := make(chan bool)
	var ranges []*Range
	go func() {
		ranges = GenerateTimeRanges([]*Range{workRange})
		done <- true
	}()

	select {
	case <-done:
		// Test completed successfully
	case <-time.After(1 * time.Second):
		t.Fatal("GenerateTimeRanges timed out after 1 second")
	}

	// Verify the results
	if len(ranges) == 0 {
		t.Error("GenerateTimeRanges() returned no ranges")
		return
	}

	// Calculate expected number of ranges
	expectedRanges := int(workRange.Duration().Minutes() / config.GetSessionDuration().Minutes())
	if len(ranges) != expectedRanges {
		t.Errorf("GenerateTimeRanges() returned %d ranges, want %d", len(ranges), expectedRanges)
	}

	// Check that ranges are sorted by decreasing length
	for i := 1; i < len(ranges); i++ {
		if ranges[i].Minutes() > ranges[i-1].Minutes() {
			t.Errorf("Ranges are not sorted by decreasing length: %v > %v", ranges[i].Minutes(), ranges[i-1].Minutes())
		}
	}

	// Check that all ranges are within the work range
	for _, r := range ranges {
		if r.Start.Before(workRange.Start) || r.End.After(workRange.End) {
			t.Errorf("Range %v-%v is outside work range %v-%v", r.Start, r.End, workRange.Start, workRange.End)
		}
	}

	// Check that ranges don't overlap
	for i := 0; i < len(ranges)-1; i++ {
		for j := i + 1; j < len(ranges); j++ {
			if HaveIntersection(ranges[i], ranges[j]) {
				t.Errorf("Ranges %v-%v and %v-%v overlap", ranges[i].Start, ranges[i].End, ranges[j].Start, ranges[j].End)
			}
		}
	}
}

func TestHaveIntersection(t *testing.T) {
	// Create a fixed date for testing
	start := time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC)
	end := time.Date(2024, 4, 1, 17, 0, 0, 0, time.UTC)
	r1 := &Range{Start: start, End: end}

	tests := []struct {
		name string
		r2   *Range
		want bool
	}{
		{
			name: "no intersection before",
			r2:   &Range{Start: start.Add(-2 * time.Hour), End: start.Add(-time.Hour)},
			want: false,
		},
		{
			name: "intersection at start",
			r2:   &Range{Start: start.Add(-time.Hour), End: start.Add(time.Hour)},
			want: true,
		},
		{
			name: "intersection in middle",
			r2:   &Range{Start: start.Add(4 * time.Hour), End: end.Add(time.Hour)},
			want: true,
		},
		{
			name: "no intersection after",
			r2:   &Range{Start: end.Add(time.Hour), End: end.Add(2 * time.Hour)},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HaveIntersection(r1, tt.r2); got != tt.want {
				t.Errorf("HaveIntersection() = %v, want %v", got, tt.want)
			}
		})
	}
}
