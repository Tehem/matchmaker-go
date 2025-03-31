package calendar

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// GetFreeSlots retrieves free time slots between busy events
// It splits the day into morning (09:00-12:00) and afternoon (13:00-17:00) slots
// and finds available time slots in each period
func GetFreeSlots(startTime, endTime time.Time, busySlots []TimeSlot) ([]TimeSlot, error) {
	// Validate time range
	if endTime.Before(startTime) {
		return nil, fmt.Errorf("end time %v is before start time %v", endTime, startTime)
	}

	// Get working hours configuration
	loc, err := time.LoadLocation(viper.GetString("workingHours.timezone"))
	if err != nil {
		return nil, err
	}

	// Create morning and afternoon time ranges
	date := startTime.Format("2006-01-02")
	morningStart := parseTime(date, viper.GetString("workingHours.morning.start"), loc)
	morningEnd := parseTime(date, viper.GetString("workingHours.morning.end"), loc)
	afternoonStart := parseTime(date, viper.GetString("workingHours.afternoon.start"), loc)
	afternoonEnd := parseTime(date, viper.GetString("workingHours.afternoon.end"), loc)

	// Find free slots in morning and afternoon
	morningSlots := findFreeSlots(morningStart, morningEnd, busySlots)
	afternoonSlots := findFreeSlots(afternoonStart, afternoonEnd, busySlots)

	// Combine all free slots
	return append(morningSlots, afternoonSlots...), nil
}

// parseTime parses a time string in the format "2006-01-02 15:04" in the given location
func parseTime(date, timeStr string, loc *time.Location) time.Time {
	t, _ := time.ParseInLocation("2006-01-02 15:04", date+" "+timeStr, loc)
	return t
}

// findFreeSlots finds available time slots between busy events in a given time range
func findFreeSlots(start, end time.Time, busySlots []TimeSlot) []TimeSlot {
	// If no busy slots, the entire range is free
	if len(busySlots) == 0 {
		return []TimeSlot{{Start: start, End: end}}
	}

	// Sort busy slots by start time for easier processing
	sortTimeSlots(busySlots)

	freeSlots := make([]TimeSlot, 0)

	// Check for free slot before first busy event
	if start.Before(busySlots[0].Start) {
		freeSlots = append(freeSlots, TimeSlot{
			Start: start,
			End:   minTime(busySlots[0].Start, end),
		})
	}

	// Check for free slots between busy events
	for i := 0; i < len(busySlots)-1; i++ {
		currentEvent := busySlots[i]
		nextEvent := busySlots[i+1]

		// If there's a gap between events, it's a free slot
		if currentEvent.End.Before(end) && nextEvent.Start.After(start) {
			freeSlots = append(freeSlots, TimeSlot{
				Start: maxTime(currentEvent.End, start),
				End:   minTime(nextEvent.Start, end),
			})
		}
	}

	// Check for free slot after last busy event
	lastEvent := busySlots[len(busySlots)-1]
	if lastEvent.End.Before(end) {
		freeSlots = append(freeSlots, TimeSlot{
			Start: maxTime(lastEvent.End, start),
			End:   end,
		})
	}

	return freeSlots
}

// Helper functions for time comparisons
func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

// sortTimeSlots sorts time slots by start time using bubble sort
func sortTimeSlots(slots []TimeSlot) {
	for i := 0; i < len(slots)-1; i++ {
		for j := 0; j < len(slots)-i-1; j++ {
			if slots[j].Start.After(slots[j+1].Start) {
				slots[j], slots[j+1] = slots[j+1], slots[j]
			}
		}
	}
}
