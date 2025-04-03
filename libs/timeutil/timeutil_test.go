package timeutil

import (
	"matchmaker/libs/testutils"
	"matchmaker/libs/types"
	"testing"
	"time"
)

func TestFirstDayOfISOWeek(t *testing.T) {
	tests := []struct {
		name      string
		weekShift int
		check     func(time.Time) bool
	}{
		{
			name:      "current week",
			weekShift: 0,
			check: func(date time.Time) bool {
				return date.Weekday() == time.Monday && date.Hour() == 0
			},
		},
		{
			name:      "next week",
			weekShift: 1,
			check: func(date time.Time) bool {
				nextWeek := time.Now().AddDate(0, 0, 7)
				return date.Weekday() == time.Monday && date.Hour() == 0 && date.Year() == nextWeek.Year()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FirstDayOfISOWeek(tt.weekShift)
			if !tt.check(got) {
				t.Errorf("FirstDayOfISOWeek() = %v, want Monday at 00:00", got)
			}
		})
	}
}

func TestGetWorkRange(t *testing.T) {
	// Create a fixed date for testing (Monday)
	beginOfWeek := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		day         int
		startHour   int
		startMinute int
		endHour     int
		endMinute   int
		wantErr     bool
	}{
		{
			name:        "valid range",
			day:         0,
			startHour:   9,
			startMinute: 0,
			endHour:     17,
			endMinute:   0,
			wantErr:     false,
		},
		{
			name:        "invalid day",
			day:         5,
			startHour:   9,
			startMinute: 0,
			endHour:     17,
			endMinute:   0,
			wantErr:     true,
		},
		{
			name:        "invalid hour",
			day:         0,
			startHour:   25,
			startMinute: 0,
			endHour:     17,
			endMinute:   0,
			wantErr:     true,
		},
		{
			name:        "invalid minute",
			day:         0,
			startHour:   9,
			startMinute: 60,
			endHour:     17,
			endMinute:   0,
			wantErr:     true,
		},
		{
			name:        "end before start",
			day:         0,
			startHour:   17,
			startMinute: 0,
			endHour:     9,
			endMinute:   0,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetWorkRange(beginOfWeek, tt.day, tt.startHour, tt.startMinute, tt.endHour, tt.endMinute)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetWorkRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got == nil {
					t.Error("GetWorkRange() returned nil range when no error was expected")
					return
				}
				expectedStart := time.Date(2024, 4, 1+tt.day, tt.startHour, tt.startMinute, 0, 0, time.UTC)
				expectedEnd := time.Date(2024, 4, 1+tt.day, tt.endHour, tt.endMinute, 0, 0, time.UTC)
				if !got.Start.Equal(expectedStart) || !got.End.Equal(expectedEnd) {
					t.Errorf("GetWorkRange() = %v-%v, want %v-%v", got.Start, got.End, expectedStart, expectedEnd)
				}
			}
		})
	}
}

func TestToSlice(t *testing.T) {
	// Create a channel and populate it with some ranges
	c := make(chan *types.Range, 3)
	ranges := []*types.Range{
		{
			Start: time.Now(),
			End:   time.Now().Add(time.Hour),
		},
		{
			Start: time.Now().Add(2 * time.Hour),
			End:   time.Now().Add(3 * time.Hour),
		},
		{
			Start: time.Now().Add(4 * time.Hour),
			End:   time.Now().Add(5 * time.Hour),
		},
	}

	// Send ranges to channel
	for _, r := range ranges {
		c <- r
	}
	close(c)

	// Convert to slice
	got := ToSlice(c)

	// Verify the results
	if len(got) != len(ranges) {
		t.Errorf("ToSlice() length = %v, want %v", len(got), len(ranges))
	}

	for i, r := range got {
		if !r.Start.Equal(ranges[i].Start) || !r.End.Equal(ranges[i].End) {
			t.Errorf("ToSlice()[%d] = %v-%v, want %v-%v", i, r.Start, r.End, ranges[i].Start, ranges[i].End)
		}
	}
}

func TestGetWeekWorkRanges(t *testing.T) {
	// Create a config mock
	configMock := testutils.NewConfigMock()
	configMock.SetupWorkHours()
	defer configMock.Restore()

	// Create a fixed date for testing (Monday)
	beginOfWeek := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		setup   func(*testutils.ConfigMock)
		wantErr bool
	}{
		{
			name: "valid configuration",
			setup: func(m *testutils.ConfigMock) {
				// Use default values
			},
			wantErr: false,
		},
		{
			name: "invalid morning hours",
			setup: func(m *testutils.ConfigMock) {
				m.SetMorningHours(25, 0, 12, 0) // Invalid hour
			},
			wantErr: true,
		},
		{
			name: "invalid afternoon hours",
			setup: func(m *testutils.ConfigMock) {
				m.SetAfternoonHours(13, 0, 25, 0) // Invalid hour
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to default values
			configMock.Reset()

			// Apply test-specific setup
			tt.setup(configMock)

			// Get the ranges
			rangesChan, err := GetWeekWorkRanges(beginOfWeek)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetWeekWorkRanges() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Convert channel to slice for easier testing
				ranges := ToSlice(rangesChan)

				// We expect 10 ranges (5 days × 2 ranges per day)
				if len(ranges) != 10 {
					t.Errorf("GetWeekWorkRanges() returned %d ranges, want 10", len(ranges))
					return
				}

				// Check first range (Monday morning)
				firstRange := ranges[0]
				expectedStart := time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC)
				expectedEnd := time.Date(2024, 4, 1, 12, 0, 0, 0, time.UTC)
				if !firstRange.Start.Equal(expectedStart) || !firstRange.End.Equal(expectedEnd) {
					t.Errorf("First range = %v-%v, want %v-%v", firstRange.Start, firstRange.End, expectedStart, expectedEnd)
				}

				// Check second range (Monday afternoon)
				secondRange := ranges[1]
				expectedStart = time.Date(2024, 4, 1, 13, 0, 0, 0, time.UTC)
				expectedEnd = time.Date(2024, 4, 1, 17, 0, 0, 0, time.UTC)
				if !secondRange.Start.Equal(expectedStart) || !secondRange.End.Equal(expectedEnd) {
					t.Errorf("Second range = %v-%v, want %v-%v", secondRange.Start, secondRange.End, expectedStart, expectedEnd)
				}
			}
		})
	}
}
