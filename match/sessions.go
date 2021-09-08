package match

import (
	"time"
)

type ReviewSession struct {
	Reviewers *Squad
	Range     *Range
}

func (session *ReviewSession) End() time.Time {
	return session.Range.End
}

func (session *ReviewSession) Start() time.Time {
	return session.Range.Start
}

func (session *ReviewSession) GetDisplayName() string {
	return "Review " + session.Reviewers.GetDisplayName()
}

func generateSessions(squads []*Squad, ranges []*Range) []*ReviewSession {
	sessions := []*ReviewSession{}
	for _, currentRange := range ranges {
		for _, squad := range squads {
			sessionPossible := true

			for _, busyRange := range squad.BusyRanges {
				if haveIntersection(currentRange, busyRange) {
					sessionPossible = false
					break
				}
			}

			if sessionPossible {
				sessions = append(sessions, &ReviewSession{
					Reviewers: squad,
					Range:     currentRange,
				})
			}
		}
	}
	return sessions
}

type ByStart []*ReviewSession

func (a ByStart) Len() int      { return len(a) }
func (a ByStart) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByStart) Less(i, j int) bool {
	iStart := a[i].Start()
	jStart := a[j].Start()
	return iStart.Before(jStart)
}
