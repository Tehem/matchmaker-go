package libs

import (
	"matchmaker/util"
	"time"

	"github.com/spf13/viper"
)

// WeeklySolve finds a single session for a tuple of people in a specific week
func WeeklySolve(problem *Problem) *Solution {
	// Generate squads for the tuple
	squads := generateSquadsForTuple(problem.People, problem.BusyTimes)

	// Generate time ranges for the week
	ranges := generateTimeRanges(problem.WorkRanges)

	// Generate possible sessions
	sessions := generateSessions(squads, ranges)

	// Find the best session (we only need one)
	var bestSession *ReviewSession
	var bestScore int
	firstValidScore := false

	for _, session := range sessions {
		// Score the session based on how well it fits
		score, isValid := scoreSession(session, problem)

		// Skip invalid sessions
		if !isValid {
			continue
		}

		// Initialize bestScore with the first valid score
		if !firstValidScore {
			bestScore = score
			bestSession = session
			firstValidScore = true
		} else if score > bestScore {
			bestScore = score
			bestSession = session
		}
	}

	// Create a solution with the best session (if found)
	solution := &Solution{
		Sessions: make([]*ReviewSession, 0),
	}

	if bestSession != nil {
		solution.Sessions = append(solution.Sessions, bestSession)
	}

	return solution
}

// generateSquadsForTuple creates squads for a specific tuple of people
func generateSquadsForTuple(people []*Person, busyTimes []*BusyTime) []*Squad {
	// For a tuple, we only need one squad with both people
	if len(people) != 2 {
		util.LogInfo("Warning: WeeklySolve expects exactly 2 people per tuple", map[string]interface{}{
			"peopleCount": len(people),
		})
		return []*Squad{}
	}

	squad := &Squad{
		People:     people,
		BusyRanges: mergeBusyRanges(busyTimes, people),
	}

	return []*Squad{squad}
}

// scoreSession assigns a score to a session based on how well it fits
// Returns (score, isValid) where isValid indicates if the session is valid
func scoreSession(session *ReviewSession, problem *Problem) (int, bool) {
	// Check if the session conflicts with any busy times
	if hasTimeConflict(session, problem.BusyTimes) {
		return 0, false // Invalid session
	}

	// Get session start time
	startTime := session.Start()

	// Get working hours from configuration
	workingHours := getWorkingHours(startTime)

	// Calculate score based on time preferences
	score := calculateTimeScore(startTime, workingHours)

	// Add day of week score
	score += calculateDayScore(startTime)

	return score, true
}

// hasTimeConflict checks if a session conflicts with any busy times
func hasTimeConflict(session *ReviewSession, busyTimes []*BusyTime) bool {
	for _, busyTime := range busyTimes {
		if haveIntersection(session.Range, busyTime.Range) {
			return true
		}
	}
	return false
}

// WorkingHours represents the configured working hours
type WorkingHours struct {
	MorningStart   time.Time
	MorningEnd     time.Time
	AfternoonStart time.Time
	AfternoonEnd   time.Time
	LunchStart     time.Time
	LunchEnd       time.Time
	LastHourStart  time.Time
}

// getWorkingHours retrieves working hours from configuration
func getWorkingHours(referenceTime time.Time) WorkingHours {
	// Get working hours from configuration
	morningStartHour := viper.GetInt("workingHours.morning.start.hour")
	morningStartMinute := viper.GetInt("workingHours.morning.start.minute")
	morningEndHour := viper.GetInt("workingHours.morning.end.hour")
	morningEndMinute := viper.GetInt("workingHours.morning.end.minute")

	afternoonStartHour := viper.GetInt("workingHours.afternoon.start.hour")
	afternoonStartMinute := viper.GetInt("workingHours.afternoon.start.minute")
	afternoonEndHour := viper.GetInt("workingHours.afternoon.end.hour")
	afternoonEndMinute := viper.GetInt("workingHours.afternoon.end.minute")

	// Create time objects for comparison
	morningStart := time.Date(
		referenceTime.Year(), referenceTime.Month(), referenceTime.Day(),
		morningStartHour, morningStartMinute, 0, 0,
		referenceTime.Location(),
	)

	morningEnd := time.Date(
		referenceTime.Year(), referenceTime.Month(), referenceTime.Day(),
		morningEndHour, morningEndMinute, 0, 0,
		referenceTime.Location(),
	)

	afternoonStart := time.Date(
		referenceTime.Year(), referenceTime.Month(), referenceTime.Day(),
		afternoonStartHour, afternoonStartMinute, 0, 0,
		referenceTime.Location(),
	)

	afternoonEnd := time.Date(
		referenceTime.Year(), referenceTime.Month(), referenceTime.Day(),
		afternoonEndHour, afternoonEndMinute, 0, 0,
		referenceTime.Location(),
	)

	// Define lunch break using the end of morning and start of afternoon
	lunchStart := morningEnd
	lunchEnd := afternoonStart

	// Last hour of the day (to avoid)
	lastHourStart := time.Date(
		referenceTime.Year(), referenceTime.Month(), referenceTime.Day(),
		afternoonEndHour-1, afternoonEndMinute, 0, 0,
		referenceTime.Location(),
	)

	return WorkingHours{
		MorningStart:   morningStart,
		MorningEnd:     morningEnd,
		AfternoonStart: afternoonStart,
		AfternoonEnd:   afternoonEnd,
		LunchStart:     lunchStart,
		LunchEnd:       lunchEnd,
		LastHourStart:  lastHourStart,
	}
}

// calculateTimeScore calculates the score based on time of day
func calculateTimeScore(startTime time.Time, hours WorkingHours) int {
	score := 0

	// Check if session is in the last hour of the day (low energy)
	if startTime.After(hours.LastHourStart) {
		score -= 20 // Strong penalty for sessions in the last hour
	}

	// Define preferred time slots
	// Just before lunch (30 minutes before lunch start)
	preferredMorningStart := hours.LunchStart.Add(-30 * time.Minute)

	// Check if session is just before lunch (preferred morning slot)
	if startTime.After(preferredMorningStart) && startTime.Before(hours.LunchStart) {
		score += 25 // Highest score for sessions just before lunch
	}

	// Check if session is just after lunch (preferred afternoon slot)
	if startTime.After(hours.LunchEnd) && startTime.Before(hours.LunchEnd.Add(1*time.Hour)) {
		score += 25 // Highest score for sessions just after lunch
	}

	// Check if session is in the morning but not in the preferred slot
	if startTime.After(hours.MorningStart) && startTime.Before(hours.MorningEnd) {
		score += 10 // Lower score for other morning sessions
	}

	// Check if session is in the afternoon but not in the preferred slot
	if startTime.After(hours.AfternoonStart) && startTime.Before(hours.AfternoonEnd) {
		score += 10 // Lower score for other afternoon sessions
	}

	return score
}

// calculateDayScore calculates the score based on day of week
func calculateDayScore(startTime time.Time) int {
	score := 0

	// Avoid Mondays and Fridays
	weekday := startTime.Weekday()
	if weekday == time.Monday || weekday == time.Friday {
		score -= 15 // Penalty for Mondays and Fridays
	} else {
		score += 5 // Bonus for other weekdays
	}

	// Prefer sessions in the middle of the week (Tuesday-Thursday)
	if weekday == time.Tuesday || weekday == time.Wednesday || weekday == time.Thursday {
		score += 10 // Additional bonus for middle of the week
	}

	return score
}
