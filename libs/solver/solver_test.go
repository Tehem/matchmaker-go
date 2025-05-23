package solver

import (
	"matchmaker/libs/testutils"
	"matchmaker/libs/types"
	"testing"
	"time"
)

func TestIsSessionCompatible(t *testing.T) {
	// Create a config mock
	configMock := testutils.NewConfigMock()
	configMock.SetupWorkHours()
	defer configMock.Restore()

	// Create test persons
	person1 := &types.Person{Email: "person1@example.com", MaxSessionsPerWeek: 2}
	person2 := &types.Person{Email: "person2@example.com", MaxSessionsPerWeek: 2}
	person3 := &types.Person{Email: "person3@example.com", MaxSessionsPerWeek: 2}

	// Create test time ranges
	start := time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC)
	range1 := &types.Range{Start: start, End: start.Add(time.Hour)}
	range2 := &types.Range{Start: start.Add(2 * time.Hour), End: start.Add(3 * time.Hour)}

	// Create test squads
	squad1 := &types.Squad{
		People: []*types.Person{person1, person2},
	}
	squad2 := &types.Squad{
		People: []*types.Person{person1, person3},
	}

	// Create test sessions
	session1 := &types.ReviewSession{
		Reviewers: squad1,
		Range:     range1,
	}
	session2 := &types.ReviewSession{
		Reviewers: squad2,
		Range:     range2,
	}

	tests := []struct {
		name     string
		session  *types.ReviewSession
		sessions []*types.ReviewSession
		want     bool
		setup    func()
	}{
		{
			name:     "empty sessions list",
			session:  session1,
			sessions: []*types.ReviewSession{},
			want:     true,
		},
		{
			name:     "same session",
			session:  session1,
			sessions: []*types.ReviewSession{session1},
			want:     false,
		},
		{
			name:    "same squad different time",
			session: session1,
			sessions: []*types.ReviewSession{{
				Reviewers: squad1,
				Range:     range2,
			}},
			want: false,
		},
		{
			name:     "different squad same person",
			session:  session2,
			sessions: []*types.ReviewSession{session1},
			want:     true,
		},
		{
			name:     "session conflicts with busy time",
			session:  session1,
			sessions: []*types.ReviewSession{},
			want:     false,
			setup: func() {
				// Add a busy time that conflicts with the session
				squad1.BusyRanges = []*types.Range{{
					Start: start.Add(30 * time.Minute),
					End:   start.Add(90 * time.Minute),
				}}
			},
		},
		{
			name:     "session overlaps with busy time",
			session:  session1,
			sessions: []*types.ReviewSession{},
			want:     false,
			setup: func() {
				// Add a busy time that overlaps with the session
				squad1.BusyRanges = []*types.Range{{
					Start: start.Add(-30 * time.Minute),
					End:   start.Add(30 * time.Minute),
				}}
			},
		},
		{
			name:     "session contained within busy time",
			session:  session1,
			sessions: []*types.ReviewSession{},
			want:     false,
			setup: func() {
				// Add a busy time that completely contains the session
				squad1.BusyRanges = []*types.Range{{
					Start: start.Add(-30 * time.Minute),
					End:   start.Add(90 * time.Minute),
				}}
			},
		},
		{
			name:     "session contains busy time",
			session:  session1,
			sessions: []*types.ReviewSession{},
			want:     false,
			setup: func() {
				// Add a busy time that is completely contained within the session
				squad1.BusyRanges = []*types.Range{{
					Start: start.Add(15 * time.Minute),
					End:   start.Add(45 * time.Minute),
				}}
			},
		},
		{
			name:     "session adjacent to busy time",
			session:  session1,
			sessions: []*types.ReviewSession{},
			want:     true,
			setup: func() {
				// Add a busy time that is adjacent to the session
				squad1.BusyRanges = []*types.Range{{
					Start: start.Add(time.Hour),
					End:   start.Add(2 * time.Hour),
				}}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset busy ranges before each test
			squad1.BusyRanges = []*types.Range{}
			squad2.BusyRanges = []*types.Range{}

			// Run setup if provided
			if tt.setup != nil {
				tt.setup()
			}

			got := isSessionCompatible(tt.session, tt.sessions)
			if got != tt.want {
				t.Errorf("isSessionCompatible() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsMissingCoverageBetter(t *testing.T) {
	tests := []struct {
		name      string
		coverage1 int
		coverage2 int
		want      bool
	}{
		{
			name:      "coverage1 better",
			coverage1: 1,
			coverage2: 2,
			want:      true,
		},
		{
			name:      "coverage2 better",
			coverage1: 2,
			coverage2: 1,
			want:      false,
		},
		{
			name:      "equal coverage",
			coverage1: 1,
			coverage2: 1,
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isMissingCoverageBetter(tt.coverage1, tt.coverage2); got != tt.want {
				t.Errorf("isMissingCoverageBetter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMissingCoverageToString(t *testing.T) {
	tests := []struct {
		name            string
		missingCoverage int
		want            string
	}{
		{
			name:            "zero coverage",
			missingCoverage: 0,
			want:            "[0]",
		},
		{
			name:            "positive coverage",
			missingCoverage: 5,
			want:            "[5]",
		},
		{
			name:            "negative coverage",
			missingCoverage: -3,
			want:            "[-3]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := missingCoverageToString(tt.missingCoverage); got != tt.want {
				t.Errorf("missingCoverageToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCoveragePeriodId(t *testing.T) {
	// Create test work ranges
	start := time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC)
	workRanges := []*types.Range{
		{Start: start, End: start.Add(8 * time.Hour)},
	}

	tests := []struct {
		name string
		date time.Time
		want int
	}{
		{
			name: "start of range",
			date: start,
			want: 0,
		},
		{
			name: "30 minutes into range",
			date: start.Add(30 * time.Minute),
			want: 1,
		},
		{
			name: "1 hour into range",
			date: start.Add(time.Hour),
			want: 2,
		},
		{
			name: "2 hours into range",
			date: start.Add(2 * time.Hour),
			want: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getCoveragePeriodId(workRanges, tt.date); got != tt.want {
				t.Errorf("getCoveragePeriodId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCoverage(t *testing.T) {
	// Create test work ranges
	start := time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC)
	workRanges := []*types.Range{
		{Start: start, End: start.Add(2 * time.Hour)},
	}

	// Create test persons
	person1 := &types.Person{Email: "person1@example.com", MaxSessionsPerWeek: 2}
	person2 := &types.Person{Email: "person2@example.com", MaxSessionsPerWeek: 2}

	// Create test squads
	squad := &types.Squad{
		People: []*types.Person{person1, person2},
	}

	// Create test sessions
	sessions := []*types.ReviewSession{
		{
			Reviewers: squad,
			Range:     &types.Range{Start: start, End: start.Add(time.Hour)},
		},
		{
			Reviewers: squad,
			Range:     &types.Range{Start: start.Add(30 * time.Minute), End: start.Add(90 * time.Minute)},
		},
	}

	coverage, maxCoverage := getCoverage(workRanges, sessions)

	// Verify coverage map
	// Expected coverage:
	// 9:00-9:30: 1 session
	// 9:30-10:00: 2 sessions
	// 10:00-10:30: 1 session
	// 10:30-11:00: 0 sessions
	expectedCoverage := map[int]int{
		0: 1, // 9:00-9:30
		1: 2, // 9:30-10:00
		2: 1, // 10:00-10:30
		3: 0, // 10:30-11:00
	}

	if len(coverage) != len(expectedCoverage) {
		t.Errorf("getCoverage() returned coverage map with %d entries, want %d", len(coverage), len(expectedCoverage))
	}

	for period, count := range expectedCoverage {
		if coverage[period] != count {
			t.Errorf("getCoverage() coverage[%d] = %d, want %d", period, coverage[period], count)
		}
	}

	// Verify max coverage
	if maxCoverage != 2 {
		t.Errorf("getCoverage() maxCoverage = %d, want 2", maxCoverage)
	}

	// Test with no sessions
	coverage, maxCoverage = getCoverage(workRanges, []*types.ReviewSession{})

	// Verify empty coverage map
	if len(coverage) != 4 {
		t.Errorf("getCoverage() with no sessions returned coverage map with %d entries, want 4", len(coverage))
	}

	// Verify all periods have 0 coverage
	for period := 0; period < 4; period++ {
		if coverage[period] != 0 {
			t.Errorf("getCoverage() with no sessions coverage[%d] = %d, want 0", period, coverage[period])
		}
	}

	// Verify max coverage is 0
	if maxCoverage != 0 {
		t.Errorf("getCoverage() with no sessions maxCoverage = %d, want 0", maxCoverage)
	}
}

func TestGetMissingConverage(t *testing.T) {
	tests := []struct {
		name        string
		coverage    map[int]int
		targetValue int
		want        int
	}{
		{
			name: "all periods meet target",
			coverage: map[int]int{
				0: 3,
				1: 3,
				2: 3,
			},
			targetValue: 2,
			want:        0,
		},
		{
			name: "some periods below target",
			coverage: map[int]int{
				0: 1,
				1: 3,
				2: 0,
			},
			targetValue: 2,
			want:        3, // (2-1) + (2-0) = 3
		},
		{
			name: "all periods below target",
			coverage: map[int]int{
				0: 1,
				1: 0,
				2: 1,
			},
			targetValue: 2,
			want:        4, // (2-1) + (2-0) + (2-1) = 4
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMissingConverage(tt.coverage, tt.targetValue); got != tt.want {
				t.Errorf("getMissingConverage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCoveragePerformance(t *testing.T) {
	// Create test work ranges
	start := time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC)
	workRanges := []*types.Range{
		{Start: start, End: start.Add(2 * time.Hour)},
	}

	// Create test persons
	person1 := &types.Person{Email: "person1@example.com", MaxSessionsPerWeek: 2}
	person2 := &types.Person{Email: "person2@example.com", MaxSessionsPerWeek: 2}

	// Create test squads
	squad := &types.Squad{
		People: []*types.Person{person1, person2},
	}

	// Create test sessions
	sessions := []*types.ReviewSession{
		{
			Reviewers: squad,
			Range:     &types.Range{Start: start, End: start.Add(time.Hour)},
		},
		{
			Reviewers: squad,
			Range:     &types.Range{Start: start.Add(30 * time.Minute), End: start.Add(90 * time.Minute)},
		},
	}

	// Test with target coverage of 1
	missingCoverage, maxCoverage := getCoveragePerformance(sessions, workRanges, 1)

	// Expected coverage:
	// 9:00-9:30: 1 session (meets target)
	// 9:30-10:00: 2 sessions (meets target)
	// 10:00-10:30: 1 session (meets target)
	// 10:30-11:00: 0 sessions (missing 1)
	// Total missing coverage: 1
	if missingCoverage != 1 {
		t.Errorf("getCoveragePerformance() missingCoverage = %d, want 1", missingCoverage)
	}

	// Max coverage should be 2 (during the overlapping period)
	if maxCoverage != 2 {
		t.Errorf("getCoveragePerformance() maxCoverage = %d, want 2", maxCoverage)
	}

	// Test with target coverage of 2
	missingCoverage, maxCoverage = getCoveragePerformance(sessions, workRanges, 2)

	// Expected coverage:
	// 9:00-9:30: 1 session (missing 1)
	// 9:30-10:00: 2 sessions (meets target)
	// 10:00-10:30: 1 session (missing 1)
	// 10:30-11:00: 0 sessions (missing 2)
	// Total missing coverage: 4
	if missingCoverage != 4 {
		t.Errorf("getCoveragePerformance() missingCoverage = %d, want 4", maxCoverage)
	}

	// Max coverage should still be 2
	if maxCoverage != 2 {
		t.Errorf("getCoveragePerformance() maxCoverage = %d, want 2", maxCoverage)
	}

	// Test with no sessions
	missingCoverage, maxCoverage = getCoveragePerformance([]*types.ReviewSession{}, workRanges, 1)

	// Expected coverage:
	// All periods missing 1
	// Total missing coverage: 4 (4 periods * 1 missing)
	if missingCoverage != 4 {
		t.Errorf("getCoveragePerformance() missingCoverage = %d, want 4", missingCoverage)
	}

	// Max coverage should be 0
	if maxCoverage != 0 {
		t.Errorf("getCoveragePerformance() maxCoverage = %d, want 0", maxCoverage)
	}
}

func TestGetSolver(t *testing.T) {
	// Create a config mock
	configMock := testutils.NewConfigMock()
	configMock.SetupWorkHours()
	defer configMock.Restore()

	// Create test persons
	person1 := &types.Person{Email: "person1@example.com", MaxSessionsPerWeek: 2}
	person2 := &types.Person{Email: "person2@example.com", MaxSessionsPerWeek: 2}
	person3 := &types.Person{Email: "person3@example.com", MaxSessionsPerWeek: 2}

	// Create test work ranges
	start := time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC)
	workRanges := []*types.Range{
		{Start: start, End: start.Add(2 * time.Hour)},
	}

	// Create test squads
	squad1 := &types.Squad{
		People: []*types.Person{person1, person2},
	}
	squad2 := &types.Squad{
		People: []*types.Person{person1, person3},
	}
	squad3 := &types.Squad{
		People: []*types.Person{person2, person3},
	}

	// Create test sessions with non-overlapping times
	sessions := []*types.ReviewSession{
		{
			Reviewers: squad1,
			Range:     &types.Range{Start: start, End: start.Add(30 * time.Minute)},
		},
		{
			Reviewers: squad2,
			Range:     &types.Range{Start: start.Add(30 * time.Minute), End: start.Add(time.Hour)},
		},
		{
			Reviewers: squad3,
			Range:     &types.Range{Start: start.Add(time.Hour), End: start.Add(90 * time.Minute)},
		},
		{
			Reviewers: squad1,
			Range:     &types.Range{Start: start.Add(90 * time.Minute), End: start.Add(2 * time.Hour)},
		},
	}

	// Create test problem
	problem := &types.Problem{
		People:           []*types.Person{person1, person2, person3},
		WorkRanges:       workRanges,
		TargetCoverage:   1,
		MaxTotalCoverage: 2,
	}

	// Get the solver function
	solve := getSolver(problem, sessions)

	// Test with empty current sessions
	bestSessions, bestCoverage := solve([]*types.ReviewSession{}, "")

	// Verify that we got a valid solution
	if len(bestSessions) == 0 {
		t.Error("getSolver() returned no sessions")
	}

	// Verify that all sessions in the solution are compatible
	for i := 0; i < len(bestSessions)-1; i++ {
		for j := i + 1; j < len(bestSessions); j++ {
			if !isSessionCompatible(bestSessions[i], []*types.ReviewSession{bestSessions[j]}) {
				t.Errorf("Sessions %d and %d are not compatible", i, j)
			}
		}
	}

	// Verify that the coverage is better than the worst case
	worstCoverage, _ := getCoveragePerformance([]*types.ReviewSession{}, workRanges, problem.TargetCoverage)
	if bestCoverage > worstCoverage {
		t.Errorf("getSolver() returned coverage %d, which is worse than worst case %d", bestCoverage, worstCoverage)
	}

	// Verify that we don't exceed max total coverage
	_, maxCoverage := getCoverage(workRanges, bestSessions)
	if maxCoverage > problem.MaxTotalCoverage {
		t.Errorf("getSolver() returned max coverage %d, which exceeds max total coverage %d", maxCoverage, problem.MaxTotalCoverage)
	}

	// Test with some initial sessions
	initialSessions := []*types.ReviewSession{sessions[0]}
	bestSessions, bestCoverage = solve(initialSessions, "")

	// Verify that initial sessions are included in the solution
	found := false
	for _, session := range bestSessions {
		if session == initialSessions[0] {
			found = true
			break
		}
	}
	if !found {
		t.Error("getSolver() did not include initial sessions in the solution")
	}

	// Verify that the solution with initial sessions is valid
	for i := 0; i < len(bestSessions)-1; i++ {
		for j := i + 1; j < len(bestSessions); j++ {
			if !isSessionCompatible(bestSessions[i], []*types.ReviewSession{bestSessions[j]}) {
				t.Errorf("Sessions %d and %d are not compatible in solution with initial sessions", i, j)
			}
		}
	}

	// Verify that the coverage with initial sessions is better than the worst case
	worstCoverage, _ = getCoveragePerformance(initialSessions, workRanges, problem.TargetCoverage)
	if bestCoverage > worstCoverage {
		t.Errorf("getSolver() with initial sessions returned coverage %d, which is worse than worst case %d", bestCoverage, worstCoverage)
	}
}
