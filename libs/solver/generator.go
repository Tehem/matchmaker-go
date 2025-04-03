package solver

import (
	"matchmaker/libs/types"
	"time"
)

// GenerateTimeRanges generates all possible time ranges for scheduling
func GenerateTimeRanges(workRanges []*types.Range) []*types.Range {
	ranges := make([]*types.Range, 0)
	for _, workRange := range workRanges {
		start := workRange.Start
		end := workRange.End
		for start.Before(end) {
			rangeEnd := start.Add(30 * time.Minute)
			if rangeEnd.After(end) {
				break
			}
			ranges = append(ranges, &types.Range{
				Start: start,
				End:   rangeEnd,
			})
			start = rangeEnd
		}
	}
	return ranges
}

// GenerateSessions generates all possible sessions for the given squads and time ranges
func GenerateSessions(squads []*types.Squad, ranges []*types.Range) []*types.ReviewSession {
	sessions := make([]*types.ReviewSession, 0)
	for _, squad := range squads {
		for _, timeRange := range ranges {
			sessions = append(sessions, &types.ReviewSession{
				Reviewers: squad,
				Range:     timeRange,
			})
		}
	}
	return sessions
}

// ByStart is a type for sorting sessions by start time
type ByStart []*types.ReviewSession

func (a ByStart) Len() int      { return len(a) }
func (a ByStart) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByStart) Less(i, j int) bool {
	return a[i].Start().Before(a[j].Start())
}
