package libs

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	logger "github.com/transcovo/go-chpr-logger"
	"matchmaker/util"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Solution struct {
	Sessions []*ReviewSession
}

func Solve(problem *Problem) *Solution {
	squads := generateSquads(problem.People, problem.BusyTimes)
	ranges := generateTimeRanges(problem.WorkRanges)
	sessions := generateSessions(squads, ranges)

	printSquads(squads)
	printRanges(ranges)

	bestSessions, _ := getSolver(problem, sessions)([]*ReviewSession{}, "")
	solution := &Solution{
		Sessions: bestSessions,
	}

	coverage, maxCoverage := getCoverage(problem.WorkRanges, bestSessions)

	missingCoverage := getMissingConverage(coverage, problem.TargetCoverage)

	worstMissingCoverage, _ := getCoveragePerformance([]*ReviewSession{}, problem.WorkRanges, problem.TargetCoverage)

	println(missingCoverageToString(missingCoverage))
	println(missingCoverageToString(worstMissingCoverage))

	println(maxCoverage)

	sort.Sort(ByStart(solution.Sessions))

	printSessions(solution.Sessions)

	//println("Coverage:")
	//for i, value := range coverage {
	//	println("  " + strconv.Itoa(i) + " -> " + strconv.Itoa(value))
	//}

	return solution
}

func printSessions(sessions []*ReviewSession) {
	print(len(sessions), " session(s):")
	println()
	for _, session := range sessions {
		printSession(session)
	}
}
func printSession(session *ReviewSession) {
	println(session.Reviewers.People[0].Email + "\t" +
		session.Reviewers.People[1].Email + "\t" +
		session.Range.Start.Format(time.Stamp) + "\t->\t" +
		session.Range.End.Format(time.Stamp) + "\t")
}

type solver func([]*ReviewSession, string) ([]*ReviewSession, int)

type partialSolution struct {
	sessions []*ReviewSession
	coverage int
}

type byCoverage []*partialSolution

func (a byCoverage) Len() int      { return len(a) }
func (a byCoverage) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byCoverage) Less(i, j int) bool {
	return isMissingCoverageBetter(a[i].coverage, a[j].coverage)
}

func getSolver(problem *Problem, allSessions []*ReviewSession) solver {
	var solve solver

	workRanges := problem.WorkRanges
	targetCoverage := problem.TargetCoverage

	bestSessions := []*ReviewSession{}
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

	solve = func(currentSessions []*ReviewSession, path string) ([]*ReviewSession, int) {
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
				sessions: make([]*ReviewSession, len(newSessions)),
				coverage: newCoveragePerformance,
			}
			copy(derivedSolution.sessions, newSessions)
			derivedSolutions = append(derivedSolutions, derivedSolution)
		}

		logger.WithFields(logrus.Fields{
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
				bestSessions = make([]*ReviewSession, len(newSessions))
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

func isSessionCompatible(session *ReviewSession, sessions []*ReviewSession) bool {
	// store the number of sessions to cap it
	reviewers := session.Reviewers
	people := reviewers.People
	minSessionSpacingHours := viper.GetDuration("sessions.minSessionSpacingHours")

	person0 := people[0]
	person0.isSessionCompatibleSessionCount = 0
	person1 := people[1]
	person1.isSessionCompatibleSessionCount = 0

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
			if haveIntersection(range1, range2) {
				return false
			}
		}
		// every reviewer must be able to attempt all the sessions
		otherPerson0.isSessionCompatibleSessionCount += 1
		otherPerson1.isSessionCompatibleSessionCount += 1
	}

	// check the max reviews per person
	maxSessionsForPerson0 := viper.GetInt("sessions.maxPerPersonPerWeek")
	maxSessionsForPerson1 := viper.GetInt("sessions.maxPerPersonPerWeek")
	if person0.MaxSessionsPerWeek != 0 {
		maxSessionsForPerson0 = person0.MaxSessionsPerWeek
	}
	if person1.MaxSessionsPerWeek != 0 {
		maxSessionsForPerson1 = person1.MaxSessionsPerWeek
	}
	return person0.isSessionCompatibleSessionCount < maxSessionsForPerson0 &&
		person1.isSessionCompatibleSessionCount < maxSessionsForPerson1
}

func printRanges(ranges []*Range) {
	for _, currentRange := range ranges {
		println(currentRange.Start.Format(time.RFC3339) + " -> " + currentRange.End.Format(time.RFC3339))
	}
}

func printSquads(squads []*Squad) {
	for _, squad := range squads {
		println(squad.People[0].Email + " + " + squad.People[1].Email)
	}
}

type Squad struct {
	People     []*Person
	BusyRanges []*Range
}

func (squad *Squad) GetDisplayName() string {
	result := ""
	for _, person := range squad.People {
		if result != "" {
			result = result + " / "
		}
		result = result + getNameFromEmail(person.Email)
	}
	return result
}

func getNameFromEmail(email string) string {
	beforeA := strings.Split(email, "@")[0]
	firstName := strings.Split(beforeA, ".")[0]
	return strings.Title(firstName)
}

type Score struct {
	Hours    int
	Coverage float32
}

var coveragePeriodSpan = 30 * time.Minute

func getCoveragePerformance(sessions []*ReviewSession, workRanges []*Range, target int) (int, int) {
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

func getCoverage(workRanges []*Range, sessions []*ReviewSession) (map[int]int, int) {
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

func getCoveragePeriodId(workRanges []*Range, date time.Time) int {
	elapsedNanoseconds := date.Sub(workRanges[0].Start).Nanoseconds()
	elapsedCoveragePeriods := elapsedNanoseconds / (30 * 60 * 1000 * 1000 * 1000)
	return int(elapsedCoveragePeriods)
}
