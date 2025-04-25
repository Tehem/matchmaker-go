package solver

import (
	"fmt"
	"matchmaker/libs/config"
	"matchmaker/libs/types"
	"matchmaker/libs/util"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// These constants control the search behavior of the matching algorithm.
// They implement a form of "beam search" or "limited breadth-first search"
// to balance between finding good solutions and keeping computation time reasonable.
const (
	// maxWidthExploration limits the "width" of the search tree during exploration.
	// It controls how many alternative solutions the algorithm will explore at each decision point.
	// When the algorithm finds multiple possible matches, it will only explore up to this many options.
	//
	// Current value: 2 (explores only the 2 best alternatives at each step)
	//
	// Recommendations:
	// - For small groups (10-20 people): 2-3 is sufficient
	// - For medium groups (20-50 people): 3-5 provides better results
	// - For large groups (50+ people): 5-10 may be needed for optimal results
	// - Higher values improve solution quality but increase computation time exponentially
	maxWidthExploration = 3

	// maxExplorationPathLength limits the "depth" of the search tree during exploration.
	// It controls how many sequential decisions the algorithm will make before stopping a particular path.
	// The algorithm stops exploring any path that exceeds this length after the first decision.
	//
	// Current value: 10 (stops exploring paths longer than 10 steps)
	//
	// Recommendations:
	// - For weekly planning: 5-10 is typically sufficient
	// - For monthly planning: 10-15 may be needed
	// - For quarterly planning: 15-20 could be beneficial
	// - Higher values allow for more complex solution paths but increase computation time
	maxExplorationPathLength = 10
)

type Solution struct {
	Sessions []*types.ReviewSession
}

func Solve(problem *types.Problem) *Solution {
	squads := generateSquads(problem.People, problem.BusyTimes)
	sessionDuration := config.GetSessionDuration()
	ranges := types.GenerateTimeRanges(problem.WorkRanges, sessionDuration)
	sessions := types.GenerateSessions(squads, ranges)

	printSquads(squads)
	printRanges(ranges)

	bestSessions, _ := getSolver(problem, sessions)([]*types.ReviewSession{}, "")
	solution := &Solution{
		Sessions: bestSessions,
	}

	coverage, maxCoverage := getCoverage(problem.WorkRanges, bestSessions)

	missingCoverage := getMissingConverage(coverage, problem.TargetCoverage)

	worstMissingCoverage, _ := getCoveragePerformance([]*types.ReviewSession{}, problem.WorkRanges, problem.TargetCoverage)

	util.LogInfo("Coverage information", map[string]interface{}{
		"missingCoverage":      missingCoverageToString(missingCoverage),
		"worstMissingCoverage": missingCoverageToString(worstMissingCoverage),
		"maxCoverage":          maxCoverage,
	})

	sort.Sort(types.ByStart(solution.Sessions))

	printSessions(solution.Sessions)

	return solution
}

func printSessions(sessions []*types.ReviewSession) {
	util.LogInfo("Generated sessions", map[string]interface{}{
		"count": len(sessions),
	})
	for _, session := range sessions {
		printSession(session)
	}
}

func printSession(session *types.ReviewSession) {
	util.LogSession("Session", session)
}

type solver func([]*types.ReviewSession, string) ([]*types.ReviewSession, int)

type partialSolution struct {
	sessions []*types.ReviewSession
	coverage int
}

type byCoverage []*partialSolution

func (a byCoverage) Len() int      { return len(a) }
func (a byCoverage) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byCoverage) Less(i, j int) bool {
	return isMissingCoverageBetter(a[i].coverage, a[j].coverage)
}

func getSolver(problem *types.Problem, allSessions []*types.ReviewSession) solver {
	var solve solver

	workRanges := problem.WorkRanges
	targetCoverage := problem.TargetCoverage

	bestSessions := []*types.ReviewSession{}
	bestCoveragePerformance, _ := getCoveragePerformance(bestSessions, workRanges, targetCoverage)

	var iterations int64 = 0

	interrupted := false
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)
	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		interrupted = true
		signal.Reset(syscall.SIGINT)
	}()

	solve = func(currentSessions []*types.ReviewSession, path string) ([]*types.ReviewSession, int) {
		derivedSolutions := []*partialSolution{}

		for _, session := range allSessions {
			if interrupted {
				break
			}

			sessionCompatible := isSessionCompatible(session, currentSessions)
			if !sessionCompatible {
				continue
			}

			newSessions := append(currentSessions, session)
			newCoveragePerformance, newMaxCoverage := getCoveragePerformance(newSessions, workRanges, targetCoverage)
			if newMaxCoverage > problem.MaxTotalCoverage {
				continue
			}

			derivedSolution := &partialSolution{
				sessions: make([]*types.ReviewSession, len(newSessions)),
				coverage: newCoveragePerformance,
			}
			copy(derivedSolution.sessions, newSessions)
			derivedSolutions = append(derivedSolutions, derivedSolution)
		}

		logrus.WithFields(logrus.Fields{
			"iterations": iterations,
			"best":       missingCoverageToString(bestCoveragePerformance),
			"path":       path,
			"children":   len(derivedSolutions),
		}).Info("Exploring children")

		if len(derivedSolutions) > 0 {
			sort.Sort(byCoverage(derivedSolutions))

			newCoveragePerformance := derivedSolutions[0].coverage
			if isMissingCoverageBetter(newCoveragePerformance, bestCoveragePerformance) {
				newSessions := derivedSolutions[0].sessions
				bestSessions = make([]*types.ReviewSession, len(newSessions))
				copy(bestSessions, newSessions)
				bestCoveragePerformance = newCoveragePerformance
			}

			for i, derivedSolution := range derivedSolutions {
				if interrupted || i >= maxWidthExploration || i > 0 && len(path) > maxExplorationPathLength {
					break
				}

				iterations += 1

				subPath := path + "/" + strconv.Itoa(i)

				solve(derivedSolution.sessions, subPath)
			}
		}

		return bestSessions, bestCoveragePerformance
	}
	return solve
}

func missingCoverageToString(missingCoverage int) string {
	return "[" + strconv.Itoa(missingCoverage) + "]"
}

func isMissingCoverageBetter(coverage1 int, coverage2 int) bool {
	return coverage1 <= coverage2
}

func isSessionCompatible(session *types.ReviewSession, sessions []*types.ReviewSession) bool {
	// store the number of sessions to cap it
	reviewers := session.Reviewers
	people := reviewers.People
	minSessionSpacingHours := config.GetMinSessionSpacing()

	person0 := people[0]
	person0.ResetSessionCount()
	person1 := people[1]
	person1.ResetSessionCount()

	for _, otherSession := range sessions {
		// not the same session two times
		if session == otherSession {
			return false
		}

		otherReviewers := otherSession.Reviewers
		// not the same squad
		if reviewers == otherReviewers {
			return false
		}

		// not the same skills (if no skills specified, the reviewer can be paired with any other reviewer)
		if len(person0.Skills) != 0 && len(person1.Skills) != 0 && len(util.Intersection(person0.Skills, person1.Skills)) == 0 {
			return false
		}

		otherPeople := otherReviewers.People

		otherPerson0 := otherPeople[0]
		otherPerson1 := otherPeople[1]
		if otherPerson0 == person0 || otherPerson0 == person1 || otherPerson1 == person0 || otherPerson1 == person1 {
			range1 := session.Range.Pad(minSessionSpacingHours)
			range2 := otherSession.Range
			if range1.Overlaps(range2) {
				return false
			}
		}
		// every reviewer must be able to attempt all the sessions
		otherPerson0.IncrementSessionCount()
		otherPerson1.IncrementSessionCount()
	}

	// check the max reviews per person
	maxSessionsForPerson0 := person0.MaxSessionsPerWeek
	maxSessionsForPerson1 := person1.MaxSessionsPerWeek
	if maxSessionsForPerson0 == 0 && maxSessionsForPerson1 == 0 {
		return false
	}
	return person0.GetSessionCount() < maxSessionsForPerson0 &&
		person1.GetSessionCount() < maxSessionsForPerson1
}

func printRanges(ranges []*types.Range) {
	for _, currentRange := range ranges {
		util.LogRange("Range", currentRange)
	}
}

func printSquads(squads []*types.Squad) {
	for _, squad := range squads {
		util.LogInfo("Squad", map[string]interface{}{
			"person1": squad.People[0].Email,
			"person2": squad.People[1].Email,
		})
	}
}

func getNameFromEmail(email string) string {
	beforeA := strings.Split(email, "@")[0]
	firstName := strings.Split(beforeA, ".")[0]
	return cases.Title(language.English).String(firstName)
}

type Score struct {
	Hours    int
	Coverage float32
}

var coveragePeriodSpan = 30 * time.Minute

func getCoveragePerformance(sessions []*types.ReviewSession, workRanges []*types.Range, target int) (int, int) {
	coverage, maxCoverage := getCoverage(workRanges, sessions)

	missingCoverage := getMissingConverage(coverage, target)

	return missingCoverage, maxCoverage
}

func getMissingConverage(exclusivityCoverage map[int]int, targetValue int) int {
	missingCoverage := 0
	for _, value := range exclusivityCoverage {
		if value < targetValue {
			missingCoverage += targetValue - value
		}
	}
	return missingCoverage
}

func getCoverage(workRanges []*types.Range, sessions []*types.ReviewSession) (map[int]int, int) {
	coverage := map[int]int{}
	for _, workRange := range workRanges {
		date := workRange.Start
		for date.Before(workRange.End) {
			coveragePeriodId := getCoveragePeriodId(workRanges, date)
			coverage[coveragePeriodId] = 0
			date = date.Add(coveragePeriodSpan)
		}
	}
	maxCoverage := 0
	for _, session := range sessions {
		date := session.Start()
		for date.Before(session.End()) {
			coveragePeriodId := getCoveragePeriodId(workRanges, date)
			coverage[coveragePeriodId] += 1
			if coverage[coveragePeriodId] > maxCoverage {
				maxCoverage = coverage[coveragePeriodId]
			}
			date = date.Add(coveragePeriodSpan)
		}
	}
	return coverage, maxCoverage
}

func getCoveragePeriodId(workRanges []*types.Range, date time.Time) int {
	elapsedNanoseconds := date.Sub(workRanges[0].Start).Nanoseconds()
	elapsedCoveragePeriods := elapsedNanoseconds / (30 * 60 * 1000 * 1000 * 1000)
	return int(elapsedCoveragePeriods)
}
