package solver

import (
	"matchmaker/libs/testutils"
	"matchmaker/libs/types"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestWeeklySolve(t *testing.T) {
	// Create a config mock
	configMock := testutils.NewConfigMock()
	configMock.SetupWorkHours()
	viper.SetDefault("sessions.sessionDurationMinutes", 60)
	defer configMock.Restore()

	// Create test persons
	person1 := &types.Person{Email: "person1@example.com", MaxSessionsPerWeek: 2}
	person2 := &types.Person{Email: "person2@example.com", MaxSessionsPerWeek: 2}

	// Create test work ranges
	start := time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC)
	workRanges := []*types.Range{
		{Start: start, End: start.Add(8 * time.Hour)},
	}

	// Create test problem
	problem := &types.Problem{
		People:     []*types.Person{person1, person2},
		WorkRanges: workRanges,
	}

	// Test with no busy times
	result := WeeklySolve(problem)

	// Verify that we got a solution
	if result.Solution == nil {
		t.Error("WeeklySolve() returned nil solution")
	}

	// Verify that we have exactly one session
	if len(result.Solution.Sessions) != 1 {
		t.Errorf("WeeklySolve() returned %d sessions, want 1", len(result.Solution.Sessions))
	}

	// Verify that the session contains both people
	session := result.Solution.Sessions[0]
	if len(session.Reviewers.People) != 2 {
		t.Errorf("WeeklySolve() session has %d people, want 2", len(session.Reviewers.People))
	}

	// Test with busy times
	busyTime := &types.BusyTime{
		Person: person1,
		Range: &types.Range{
			Start: start.Add(2 * time.Hour),
			End:   start.Add(4 * time.Hour),
		},
	}
	problem.BusyTimes = []*types.BusyTime{busyTime}

	result = WeeklySolve(problem)

	// Verify that the session doesn't conflict with busy time
	session = result.Solution.Sessions[0]
	if session.Range.Overlaps(busyTime.Range) {
		t.Error("WeeklySolve() returned session that conflicts with busy time")
	}
}

func TestGenerateSquadsForTuple(t *testing.T) {
	// Create test persons
	person1 := &types.Person{Email: "person1@example.com", MaxSessionsPerWeek: 2}
	person2 := &types.Person{Email: "person2@example.com", MaxSessionsPerWeek: 2}

	// Create test busy times
	start := time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC)
	busyTime := &types.BusyTime{
		Person: person1,
		Range: &types.Range{
			Start: start.Add(2 * time.Hour),
			End:   start.Add(4 * time.Hour),
		},
	}

	// Test with two people
	squads := generateSquadsForTuple([]*types.Person{person1, person2}, []*types.BusyTime{busyTime})

	// Verify that we got exactly one squad
	if len(squads) != 1 {
		t.Errorf("generateSquadsForTuple() returned %d squads, want 1", len(squads))
	}

	// Verify that the squad contains both people
	squad := squads[0]
	if len(squad.People) != 2 {
		t.Errorf("generateSquadsForTuple() squad has %d people, want 2", len(squad.People))
	}

	// Verify that the busy ranges are merged
	if len(squad.BusyRanges) != 1 {
		t.Errorf("generateSquadsForTuple() squad has %d busy ranges, want 1", len(squad.BusyRanges))
	}

	// Test with wrong number of people
	squads = generateSquadsForTuple([]*types.Person{person1}, []*types.BusyTime{})
	if len(squads) != 0 {
		t.Errorf("generateSquadsForTuple() with wrong number of people returned %d squads, want 0", len(squads))
	}
}

func TestScoreSession(t *testing.T) {
	// Create a config mock
	configMock := testutils.NewConfigMock()
	configMock.SetupWorkHours()
	defer configMock.Restore()

	// Create test persons
	person1 := &types.Person{Email: "person1@example.com", MaxSessionsPerWeek: 2}
	person2 := &types.Person{Email: "person2@example.com", MaxSessionsPerWeek: 2}

	// Create test squad
	squad := &types.Squad{
		People: []*types.Person{person1, person2},
	}

	// Create test problem
	problem := &types.Problem{
		People: []*types.Person{person1, person2},
	}

	// Test with session in preferred morning slot on Wednesday (best case)
	// Default morning hours: 9:00-12:00, preferred slot is 1 hour before lunch (11:00-12:00)
	// This time matches both preferred morning slot (+25) and morning slot (+10)
	start := time.Date(2024, 4, 3, 11, 0, 0, 0, time.UTC) // Wednesday, 11:00 AM
	session := &types.ReviewSession{
		Reviewers: squad,
		Range:     &types.Range{Start: start, End: start.Add(time.Hour)},
	}

	score, isValid := scoreSession(session, problem)
	if !isValid {
		t.Error("scoreSession() returned invalid for valid session")
	}
	// Expected score: 35 (time score) + 5 (weekday bonus) + 10 (middle of week bonus) = 50
	if score != 50 {
		t.Errorf("scoreSession() returned score %d for preferred morning slot on Wednesday, want 50", score)
	}

	// Test with session in preferred afternoon slot on Tuesday
	// Default afternoon hours: 13:00-17:00, preferred slot is 1 hour after lunch (13:00-14:00)
	// This time matches both preferred afternoon slot (+25) and afternoon slot (+10)
	start = time.Date(2024, 4, 2, 13, 0, 0, 0, time.UTC) // Tuesday, 1:00 PM
	session = &types.ReviewSession{
		Reviewers: squad,
		Range:     &types.Range{Start: start, End: start.Add(time.Hour)},
	}

	score, isValid = scoreSession(session, problem)
	if !isValid {
		t.Error("scoreSession() returned invalid for valid session")
	}
	// Expected score: 35 (time score) + 5 (weekday bonus) + 10 (middle of week bonus) = 50
	if score != 50 {
		t.Errorf("scoreSession() returned score %d for preferred afternoon slot on Tuesday, want 50", score)
	}

	// Test with session in last hour of the day on Monday
	// Default afternoon hours: 13:00-17:00, last hour starts at 16:00
	// This time matches both last hour (-20) and afternoon slot (+10)
	start = time.Date(2024, 4, 1, 16, 0, 0, 0, time.UTC) // Monday, 4:00 PM
	session = &types.ReviewSession{
		Reviewers: squad,
		Range:     &types.Range{Start: start, End: start.Add(time.Hour)},
	}

	score, isValid = scoreSession(session, problem)
	if !isValid {
		t.Error("scoreSession() returned invalid for valid session")
	}
	// Expected score: -10 (time score) - 15 (Monday penalty) = -25
	if score != -25 {
		t.Errorf("scoreSession() returned score %d for last hour slot on Monday, want -25", score)
	}

	// Test with session in regular morning slot on Friday
	// Default morning hours: 9:00-12:00
	// This time only matches morning slot (+10)
	start = time.Date(2024, 4, 5, 10, 0, 0, 0, time.UTC) // Friday, 10:00 AM
	session = &types.ReviewSession{
		Reviewers: squad,
		Range:     &types.Range{Start: start, End: start.Add(time.Hour)},
	}

	score, isValid = scoreSession(session, problem)
	if !isValid {
		t.Error("scoreSession() returned invalid for valid session")
	}
	// Expected score: 10 (time score) - 15 (Friday penalty) = -5
	if score != -5 {
		t.Errorf("scoreSession() returned score %d for regular morning slot on Friday, want -5", score)
	}

	// Test with session in regular afternoon slot on Thursday
	// Default afternoon hours: 13:00-17:00
	// This time only matches afternoon slot (+10)
	start = time.Date(2024, 4, 4, 15, 0, 0, 0, time.UTC) // Thursday, 3:00 PM
	session = &types.ReviewSession{
		Reviewers: squad,
		Range:     &types.Range{Start: start, End: start.Add(time.Hour)},
	}

	score, isValid = scoreSession(session, problem)
	if !isValid {
		t.Error("scoreSession() returned invalid for valid session")
	}
	// Expected score: 10 (time score) + 5 (weekday bonus) + 10 (middle of week bonus) = 25
	if score != 25 {
		t.Errorf("scoreSession() returned score %d for regular afternoon slot on Thursday, want 25", score)
	}

	// Test with session conflicting with busy time
	busyTime := &types.BusyTime{
		Person: person1,
		Range: &types.Range{
			Start: start,
			End:   start.Add(time.Hour),
		},
	}
	problem.BusyTimes = []*types.BusyTime{busyTime}

	score, isValid = scoreSession(session, problem)
	if isValid {
		t.Error("scoreSession() returned valid for session with time conflict")
	}
}

func TestHasTimeConflict(t *testing.T) {
	// Create test session
	start := time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC)
	session := &types.ReviewSession{
		Range: &types.Range{Start: start, End: start.Add(time.Hour)},
	}

	// Test with no busy times
	if hasTimeConflict(session, []*types.BusyTime{}) {
		t.Error("hasTimeConflict() returned true for session with no busy times")
	}

	// Test with non-conflicting busy time
	busyTime := &types.BusyTime{
		Range: &types.Range{
			Start: start.Add(2 * time.Hour),
			End:   start.Add(3 * time.Hour),
		},
	}
	if hasTimeConflict(session, []*types.BusyTime{busyTime}) {
		t.Error("hasTimeConflict() returned true for session with non-conflicting busy time")
	}

	// Test with conflicting busy time
	busyTime = &types.BusyTime{
		Range: &types.Range{
			Start: start.Add(30 * time.Minute),
			End:   start.Add(90 * time.Minute),
		},
	}
	if !hasTimeConflict(session, []*types.BusyTime{busyTime}) {
		t.Error("hasTimeConflict() returned false for session with conflicting busy time")
	}
}

func TestCalculateTimeScore(t *testing.T) {
	// Create a config mock
	configMock := testutils.NewConfigMock()
	configMock.SetupWorkHours()
	defer configMock.Restore()

	// Get working hours for a reference time
	referenceTime := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)
	hours := getWorkingHours(referenceTime)

	// Print working hours for debugging
	t.Logf("Working hours: Morning %s-%s, Afternoon %s-%s, Lunch %s-%s, Last hour starts at %s",
		hours.MorningStart.Format("15:04"),
		hours.MorningEnd.Format("15:04"),
		hours.AfternoonStart.Format("15:04"),
		hours.AfternoonEnd.Format("15:04"),
		hours.LunchStart.Format("15:04"),
		hours.LunchEnd.Format("15:04"),
		hours.LastHourStart.Format("15:04"),
	)

	// Test with session at the start of preferred morning slot (1 hour before lunch)
	// Default morning hours: 9:00-12:00, lunch at 12:00
	// This time matches both preferred morning slot (+25) and morning slot (+10)
	start := hours.LunchStart.Add(-1 * time.Hour) // 11:00
	score := calculateTimeScore(start, hours)
	t.Logf("Test 1: Time %s, Score %d", start.Format("15:04"), score)
	if score != 35 {
		t.Errorf("calculateTimeScore() returned score %d for start of preferred morning slot, want 35 (25 + 10)", score)
	}

	// Test with session at the end of preferred morning slot (lunch start)
	start = hours.LunchStart // 12:00
	score = calculateTimeScore(start, hours)
	t.Logf("Test 2: Time %s, Score %d", start.Format("15:04"), score)
	if score != 35 {
		t.Errorf("calculateTimeScore() returned score %d for end of preferred morning slot, want 35 (25 + 10)", score)
	}

	// Test with session at the start of preferred afternoon slot (lunch end)
	// Default afternoon hours: 13:00-17:00, lunch ends at 13:00
	// This time matches both preferred afternoon slot (+25) and afternoon slot (+10)
	start = hours.LunchEnd // 13:00
	score = calculateTimeScore(start, hours)
	t.Logf("Test 3: Time %s, Score %d", start.Format("15:04"), score)
	if score != 35 {
		t.Errorf("calculateTimeScore() returned score %d for start of preferred afternoon slot, want 35 (25 + 10)", score)
	}

	// Test with session at the end of preferred afternoon slot
	start = hours.LunchEnd.Add(1 * time.Hour) // 14:00
	score = calculateTimeScore(start, hours)
	t.Logf("Test 4: Time %s, Score %d", start.Format("15:04"), score)
	if score != 35 {
		t.Errorf("calculateTimeScore() returned score %d for end of preferred afternoon slot, want 35 (25 + 10)", score)
	}

	// Test with session at the start of last hour
	// Default afternoon hours: 13:00-17:00, last hour starts at 16:00
	// This time matches both last hour (-20) and afternoon slot (+10)
	start = hours.LastHourStart // 16:00
	score = calculateTimeScore(start, hours)
	t.Logf("Test 5: Time %s, Score %d", start.Format("15:04"), score)
	if score != -10 {
		t.Errorf("calculateTimeScore() returned score %d for start of last hour, want -10 (-20 + 10)", score)
	}

	// Test with session at the end of last hour
	start = hours.AfternoonEnd // 17:00
	score = calculateTimeScore(start, hours)
	t.Logf("Test 6: Time %s, Score %d", start.Format("15:04"), score)
	if score != -10 {
		t.Errorf("calculateTimeScore() returned score %d for end of last hour, want -10 (-20 + 10)", score)
	}

	// Test with session at the start of regular morning slot
	// Default morning hours: 9:00-12:00
	// This time only matches morning slot (+10)
	start = hours.MorningStart.Add(30 * time.Minute) // 9:30
	score = calculateTimeScore(start, hours)
	t.Logf("Test 7: Time %s, Score %d", start.Format("15:04"), score)
	if score != 10 {
		t.Errorf("calculateTimeScore() returned score %d for regular morning slot, want 10", score)
	}

	// Test with session at the end of regular morning slot
	// Use 11:00 which is before the preferred slot starts
	start = hours.LunchStart.Add(-2 * time.Hour) // 10:00
	score = calculateTimeScore(start, hours)
	t.Logf("Test 8: Time %s, Score %d", start.Format("15:04"), score)
	if score != 10 {
		t.Errorf("calculateTimeScore() returned score %d for regular morning slot, want 10", score)
	}

	// Test with session at the start of regular afternoon slot
	// Use 15:00 which is after the preferred slot ends
	start = hours.LunchEnd.Add(2 * time.Hour) // 15:00
	score = calculateTimeScore(start, hours)
	t.Logf("Test 9: Time %s, Score %d", start.Format("15:04"), score)
	if score != 10 {
		t.Errorf("calculateTimeScore() returned score %d for regular afternoon slot, want 10", score)
	}

	// Test with session at the end of regular afternoon slot
	// Use 15:30 which is before the last hour starts
	start = hours.LastHourStart.Add(-30 * time.Minute) // 15:30
	score = calculateTimeScore(start, hours)
	t.Logf("Test 10: Time %s, Score %d", start.Format("15:04"), score)
	if score != 10 {
		t.Errorf("calculateTimeScore() returned score %d for regular afternoon slot, want 10", score)
	}

	// Test with session outside working hours (should get no score)
	start = hours.MorningStart.Add(-time.Hour) // 8:00
	score = calculateTimeScore(start, hours)
	t.Logf("Test 11: Time %s, Score %d", start.Format("15:04"), score)
	if score != 0 {
		t.Errorf("calculateTimeScore() returned score %d for outside working hours, want 0", score)
	}
}

func TestCalculateDayScore(t *testing.T) {
	// Test with Monday
	start := time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC) // Monday
	score := calculateDayScore(start)
	if score > -15 {
		t.Errorf("calculateDayScore() returned score %d for Monday, want <= -15", score)
	}

	// Test with Friday
	start = time.Date(2024, 4, 5, 9, 0, 0, 0, time.UTC) // Friday
	score = calculateDayScore(start)
	if score > -15 {
		t.Errorf("calculateDayScore() returned score %d for Friday, want <= -15", score)
	}

	// Test with Wednesday
	start = time.Date(2024, 4, 3, 9, 0, 0, 0, time.UTC) // Wednesday
	score = calculateDayScore(start)
	if score < 5 {
		t.Errorf("calculateDayScore() returned score %d for Wednesday, want >= 5", score)
	}
}
