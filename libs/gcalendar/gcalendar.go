package gcalendar

import (
	"fmt"
	"matchmaker/libs/types"
	"matchmaker/libs/util"
	"time"

	"google.golang.org/api/calendar/v3"
)

// GCalendar represents a Google Calendar client
type GCalendar struct {
	service *calendar.Service
}

// NewGCalendar creates a new GCalendar client
func NewGCalendar() (*GCalendar, error) {
	service, err := GetCalendarService()
	if err != nil {
		return nil, err
	}
	return &GCalendar{service: service}, nil
}

// FormatTime formats a time.Time to RFC3339 string
func FormatTime(date time.Time) string {
	return date.Format(time.RFC3339)
}

// GetEvents retrieves events from the calendar
func (g *GCalendar) GetEvents(timeMin, timeMax time.Time) ([]*calendar.Event, error) {
	events, err := g.service.Events.List("primary").ShowDeleted(false).
		SingleEvents(true).TimeMin(FormatTime(timeMin)).TimeMax(FormatTime(timeMax)).
		OrderBy("startTime").Do()
	if err != nil {
		return nil, err
	}
	return events.Items, nil
}

// GetNextEvents retrieves the next N events from the calendar
func (g *GCalendar) GetNextEvents(timeMin time.Time, maxResults int64) ([]*calendar.Event, error) {
	events, err := g.service.Events.List("primary").ShowDeleted(false).
		SingleEvents(true).TimeMin(FormatTime(timeMin)).MaxResults(maxResults).
		OrderBy("startTime").Do()
	if err != nil {
		return nil, err
	}
	return events.Items, nil
}

// CreateEvent creates a new event in the calendar
func (g *GCalendar) CreateEvent(event *calendar.Event) (*calendar.Event, error) {
	return g.service.Events.Insert("primary", event).Do()
}

// UpdateEvent updates an existing event in the calendar
func (g *GCalendar) UpdateEvent(event *calendar.Event) (*calendar.Event, error) {
	return g.service.Events.Update("primary", event.Id, event).Do()
}

// DeleteEvent deletes an event from the calendar
func (g *GCalendar) DeleteEvent(eventID string) error {
	return g.service.Events.Delete("primary", eventID).Do()
}

// GetFreeBusy retrieves free/busy information for a list of calendars
func (g *GCalendar) GetFreeBusy(timeMin, timeMax time.Time, calendars []string) (*calendar.FreeBusyResponse, error) {
	timeMinStr := FormatTime(timeMin)
	timeMaxStr := FormatTime(timeMax)

	freeBusyRequest := &calendar.FreeBusyRequest{
		TimeMin: timeMinStr,
		TimeMax: timeMaxStr,
		Items:   make([]*calendar.FreeBusyRequestItem, len(calendars)),
	}

	for i, cal := range calendars {
		freeBusyRequest.Items[i] = &calendar.FreeBusyRequestItem{
			Id: cal,
		}
	}

	return g.service.Freebusy.Query(freeBusyRequest).Do()
}

// FindAvailableSlots finds available time slots for a list of calendars
func (g *GCalendar) FindAvailableSlots(timeMin, timeMax time.Time, calendars []string, durationMinutes int) ([]types.Range, error) {
	freeBusy, err := g.GetFreeBusy(timeMin, timeMax, calendars)
	if err != nil {
		return nil, err
	}

	// Convert free/busy information to ranges
	var busyRanges []types.Range
	for _, cal := range calendars {
		if calBusy, ok := freeBusy.Calendars[cal]; ok {
			for _, busy := range calBusy.Busy {
				start, err := time.Parse(time.RFC3339, busy.Start)
				if err != nil {
					return nil, err
				}
				end, err := time.Parse(time.RFC3339, busy.End)
				if err != nil {
					return nil, err
				}
				busyRanges = append(busyRanges, types.Range{
					Start: start,
					End:   end,
				})
			}
		}
	}

	// Find available slots
	var availableSlots []types.Range
	current := timeMin
	duration := time.Duration(durationMinutes) * time.Minute

	for current.Add(duration).Before(timeMax) || current.Add(duration).Equal(timeMax) {
		slotEnd := current.Add(duration)
		isAvailable := true

		for _, busy := range busyRanges {
			if (current.After(busy.Start) && current.Before(busy.End)) ||
				(slotEnd.After(busy.Start) && slotEnd.Before(busy.End)) ||
				(current.Before(busy.Start) && slotEnd.After(busy.End)) {
				isAvailable = false
				break
			}
		}

		if isAvailable {
			availableSlots = append(availableSlots, types.Range{
				Start: current,
				End:   slotEnd,
			})
		}

		current = current.Add(time.Hour)
	}

	return availableSlots, nil
}

// CreateSessionEvent creates a session event in the calendar
func (g *GCalendar) CreateSessionEvent(session *types.ReviewSession) (*calendar.Event, error) {
	event := &calendar.Event{
		Summary: session.GetEventSummary(),
		Start: &calendar.EventDateTime{
			DateTime: session.Range.Start.Format(time.RFC3339),
			TimeZone: "UTC",
		},
		End: &calendar.EventDateTime{
			DateTime: session.Range.End.Format(time.RFC3339),
			TimeZone: "UTC",
		},
		Attendees: []*calendar.EventAttendee{
			{Email: session.Reviewers.People[0].Email},
			{Email: session.Reviewers.People[1].Email},
		},
		Reminders: &calendar.EventReminders{
			UseDefault: false,
			Overrides: []*calendar.EventReminder{
				{Method: "email", Minutes: 24 * 60},
				{Method: "popup", Minutes: 10},
			},
		},
	}

	return g.CreateEvent(event)
}

// GetBusyTimesForPeople retrieves busy times for multiple people across work ranges
func (g *GCalendar) GetBusyTimesForPeople(people []*types.Person, workRanges []*types.Range) []*types.BusyTime {
	busyTimes := []*types.BusyTime{}
	for _, person := range people {
		if !person.CanParticipateInSession() {
			continue
		}
		for _, workRange := range workRanges {
			personBusyTimes, err := g.GetBusyTimes(person, workRange)
			util.PanicOnError(err, "Cannot load busy times for person")
			busyTimes = append(busyTimes, personBusyTimes...)
		}
	}
	return busyTimes
}

// GetBusyTimes retrieves busy time slots for a person within a given time range
func (g *GCalendar) GetBusyTimes(person *types.Person, timeRange *types.Range) ([]*types.BusyTime, error) {
	util.LogInfo("Loading busy detail", map[string]interface{}{
		"person": person.Email,
	})
	util.LogRange("Time range", timeRange)

	result, err := g.service.Freebusy.Query(&calendar.FreeBusyRequest{
		TimeMin: FormatTime(timeRange.Start),
		TimeMax: FormatTime(timeRange.End),
		Items: []*calendar.FreeBusyRequestItem{
			{
				Id: person.Email,
			},
		},
	}).Do()
	if err != nil {
		return nil, fmt.Errorf("can't retrieve free/busy data for %s: %w", person.Email, err)
	}

	busyTimes := make([]*types.BusyTime, 0)
	busyTimePeriods := result.Calendars[person.Email].Busy
	util.LogInfo("Person busy times", map[string]interface{}{
		"person": person.Email,
	})

	for _, busyTimePeriod := range busyTimePeriods {
		busyTime := &types.BusyTime{
			Person: person,
			Range: &types.Range{
				Start: parseTime(busyTimePeriod.Start),
				End:   parseTime(busyTimePeriod.End),
			},
		}
		util.LogRange("Busy time period", busyTime.Range)
		busyTimes = append(busyTimes, busyTime)
	}

	return busyTimes, nil
}

// parseTime parses a time string in RFC3339 format
func parseTime(dateStr string) time.Time {
	result, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		util.LogError(err, "Impossible to parse date "+dateStr)
		return time.Time{}
	}
	return result
}
