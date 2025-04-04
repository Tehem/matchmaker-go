package solver

import (
	"matchmaker/libs/config"
	"matchmaker/libs/types"
	"matchmaker/libs/util"
	"time"

	"github.com/spf13/viper"
)

// WeeklySolveResult contains the result of the weekly solve operation
type WeeklySolveResult struct {
	Solution        *types.Solution
	UnmatchedPeople []*types.Person
	UnmatchedTuples []types.Tuple
}

// WeeklySolve finds a single session for a tuple of people in a specific week
func WeeklySolve(problem *types.Problem) *WeeklySolveResult {
	// Generate squads for the tuple
	squads := generateSquadsForTuple(problem.People, problem.BusyTimes)

	// Generate time ranges for the work ranges
	sessionDuration := config.GetSessionDuration()
	ranges := types.GenerateTimeRanges(problem.WorkRanges, sessionDuration)

	// Generate possible sessions
	sessions := types.GenerateSessions(squads, ranges)

	// Find the best session (we only need one)
	var bestSession *types.ReviewSession
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
	solution := &types.Solution{
		Sessions: make([]*types.ReviewSession, 0),
	}

	if bestSession != nil {
		solution.Sessions = append(solution.Sessions, bestSession)
	}

	// Create the result
	result := &WeeklySolveResult{
		Solution:        solution,
		UnmatchedPeople: make([]*types.Person, 0),
		UnmatchedTuples: make([]types.Tuple, 0),
	}

	// If no session was found, add the tuple to unmatched tuples
	if bestSession == nil && len(problem.People) == 2 {
		// Create a tuple from the two people
		tuple := types.Tuple{
			Person1: problem.People[0],
			Person2: problem.People[1],
		}
		result.UnmatchedTuples = append(result.UnmatchedTuples, tuple)
	}

	return result
}

// generateSquadsForTuple creates squads for a specific tuple of people
func generateSquadsForTuple(people []*types.Person, busyTimes []*types.BusyTime) []*types.Squad {
	// For a tuple, we only need one squad with both people
	if len(people) != 2 {
		util.LogInfo("Warning: WeeklySolve expects exactly 2 people per tuple", map[string]interface{}{
			"peopleCount": len(people),
		})
		return []*types.Squad{}
	}

	squad := &types.Squad{
		People:     people,
		BusyRanges: mergeBusyRanges(busyTimes, people),
	}

	return []*types.Squad{squad}
}

// scoreSession assigns a score to a session based on how well it fits
// Returns (score, isValid) where isValid indicates if the session is valid
func scoreSession(session *types.ReviewSession, problem *types.Problem) (int, bool) {
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
func hasTimeConflict(session *types.ReviewSession, busyTimes []*types.BusyTime) bool {
	for _, busyTime := range busyTimes {
		if session.Range.Overlaps(busyTime.Range) {
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

// isInRange checks if a time is within a range (inclusive)
func isInRange(t, start, end time.Time) bool {
	return !t.Before(start) && !t.After(end)
}

// calculateTimeScore calculates the score based on time of day
func calculateTimeScore(startTime time.Time, hours WorkingHours) int {
	score := 0

	// Check if session is in the last hour of the day (low energy)
	if isInRange(startTime, hours.LastHourStart, hours.AfternoonEnd) {
		score -= 20 // Strong penalty for sessions in the last hour
	}

	// Define preferred time slots
	preferredMorningStart := hours.LunchStart.Add(-1 * time.Hour)
	preferredAfternoonEnd := hours.LunchEnd.Add(1 * time.Hour)

	// Check if session is in preferred morning slot (1 hour before lunch)
	if isInRange(startTime, preferredMorningStart, hours.LunchStart) {
		score += 25 // Highest score for sessions just before lunch
	}

	// Check if session is in preferred afternoon slot (1 hour after lunch)
	if isInRange(startTime, hours.LunchEnd, preferredAfternoonEnd) {
		score += 25 // Highest score for sessions just after lunch
	}

	// Check if session is in regular morning slot
	if isInRange(startTime, hours.MorningStart, hours.MorningEnd) {
		score += 10 // Lower score for other morning sessions
	}

	// Check if session is in regular afternoon slot
	if isInRange(startTime, hours.AfternoonStart, hours.AfternoonEnd) {
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

// FindSessionForTuple finds a session for a tuple of people in a specific week
func FindSessionForTuple(tuple types.Tuple, workRanges []*types.Range, busyTimes []*types.BusyTime) *types.ReviewSession {
	// Create a problem for the tuple
	problem := &types.Problem{
		People:         []*types.Person{tuple.Person1, tuple.Person2},
		WorkRanges:     workRanges,
		BusyTimes:      busyTimes,
		TargetCoverage: 0,
	}

	// Find a session using the weekly solver
	result := WeeklySolve(problem)
	if len(result.Solution.Sessions) > 0 {
		return result.Solution.Sessions[0]
	}
	return nil
}
