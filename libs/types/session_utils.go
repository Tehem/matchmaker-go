package types

// GenerateSessions generates all possible sessions for the given squads and time ranges
func GenerateSessions(squads []*Squad, ranges []*Range) []*ReviewSession {
	sessions := make([]*ReviewSession, 0)
	for _, squad := range squads {
		for _, timeRange := range ranges {
			sessions = append(sessions, &ReviewSession{
				Reviewers: squad,
				Range:     timeRange,
			})
		}
	}
	return sessions
}

// ByStart is a type for sorting sessions by start time
type ByStart []*ReviewSession

func (a ByStart) Len() int      { return len(a) }
func (a ByStart) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByStart) Less(i, j int) bool {
	return a[i].Start().Before(a[j].Start())
}
