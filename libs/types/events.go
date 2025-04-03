package types

// EventBatch represents a collection of events created in a single plan command run
type EventBatch struct {
	ID        string  `json:"id"`
	CreatedAt string  `json:"created_at"`
	Events    []Event `json:"events"`
}

// Event represents a created calendar event
type Event struct {
	ID        string `json:"id"`
	Summary   string `json:"summary"`
	Organizer string `json:"organizer"`
}
