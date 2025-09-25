package service

import (
	"sync"
	"testing"
	"time"
)

func TestCreateEvent(t *testing.T) {
	calendar := NewCalendar()
	userID := 1
	date := time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)
	title := "New Year Party"

	err := calendar.CreateEvent(userID, date, title)
	if err != nil {
		t.Fatalf("CreateEvent failed: %v", err)
	}

	events := calendar.GetEventsForDay(userID, date)
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}

	event := events[0]
	if event.UserID != userID {
		t.Errorf("Expected UserID %d, got %d", userID, event.UserID)
	}
	if event.Title != title {
		t.Errorf("Expected title %s, got %s", title, event.Title)
	}
	if !isSameDay(event.Date, date) {
		t.Errorf("Expected date %v, got %v", date, event.Date)
	}
}

func TestUpdateEvent(t *testing.T) {
	calendar := NewCalendar()
	userID := 1
	date := time.Now()
	title := "Original Event"

	calendar.CreateEvent(userID, date, title)

	events := calendar.GetEventsForDay(userID, date)
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}
	eventID := events[0].ID

	newDate := date.AddDate(0, 0, 1)
	newTitle := "Updated Event"

	updatedEvent, err := calendar.UpdateEvent(eventID, userID, newDate, newTitle)
	if err != nil {
		t.Fatalf("UpdateEvent failed: %v", err)
	}

	if updatedEvent.Title != newTitle {
		t.Errorf("Expected title %s, got %s", newTitle, updatedEvent.Title)
	}
	if !updatedEvent.Date.Equal(newDate) {
		t.Errorf("Expected date %v, got %v", newDate, updatedEvent.Date)
	}

	eventsAfterUpdate := calendar.GetEventsForDay(userID, newDate)
	if len(eventsAfterUpdate) != 1 {
		t.Errorf("Expected 1 event after update, got %d", len(eventsAfterUpdate))
	}
}

func TestUpdateEventNotFound(t *testing.T) {
	calendar := NewCalendar()

	_, err := calendar.UpdateEvent(999, 1, time.Now(), "Test")
	if err != ErrEventNotFound {
		t.Errorf("Expected ErrEventNotFound, got %v", err)
	}
}

func TestDeleteEvent(t *testing.T) {
	calendar := NewCalendar()
	userID := 1
	date := time.Now()

	calendar.CreateEvent(userID, date, "Event to delete")

	events := calendar.GetEventsForDay(userID, date)
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}
	eventID := events[0].ID

	err := calendar.DeleteEvent(eventID, userID)
	if err != nil {
		t.Fatalf("DeleteEvent failed: %v", err)
	}

	eventsAfterDelete := calendar.GetEventsForDay(userID, date)
	if len(eventsAfterDelete) != 0 {
		t.Errorf("Expected 0 events after deletion, got %d", len(eventsAfterDelete))
	}
}

func TestDeleteEventNotFound(t *testing.T) {
	calendar := NewCalendar()
	err := calendar.DeleteEvent(999, 1)
	if err != ErrEventNotFound {
		t.Errorf("Expected ErrEventNotFound for non-existent event, got %v", err)
	}
}

func TestGetEventsForDay(t *testing.T) {
	calendar := NewCalendar()
	userID := 1
	targetDate := time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)
	otherDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	calendar.CreateEvent(userID, targetDate, "Event 1")
	calendar.CreateEvent(userID, targetDate, "Event 2")
	calendar.CreateEvent(userID, otherDate, "Event 3")
	calendar.CreateEvent(2, targetDate, "Event 4")

	events := calendar.GetEventsForDay(userID, targetDate)
	if len(events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events))
	}

	for _, event := range events {
		if event.UserID != userID {
			t.Errorf("Event has wrong user ID: %d", event.UserID)
		}
		if !isSameDay(event.Date, targetDate) {
			t.Errorf("Event has wrong date: %v", event.Date)
		}
	}
}

func TestGetEventsForWeek(t *testing.T) {
	calendar := NewCalendar()
	userID := 1
	monday := time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC)

	calendar.CreateEvent(userID, monday, "Monday Event")
	calendar.CreateEvent(userID, monday.AddDate(0, 0, 3), "Thursday Event")
	calendar.CreateEvent(userID, monday.AddDate(0, 0, 6), "Sunday Event")
	calendar.CreateEvent(userID, monday.AddDate(0, 0, 7), "Next Week Event")

	events := calendar.GetEventsForWeek(userID, monday.AddDate(0, 0, 3))
	if len(events) != 3 {
		t.Errorf("Expected 3 events for the week, got %d", len(events))
	}
}

func TestGetEventsForMonth(t *testing.T) {
	calendar := NewCalendar()
	userID := 1
	dec15 := time.Date(2023, 12, 15, 0, 0, 0, 0, time.UTC)

	calendar.CreateEvent(userID, time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC), "Dec 1")
	calendar.CreateEvent(userID, time.Date(2023, 12, 15, 0, 0, 0, 0, time.UTC), "Dec 15")
	calendar.CreateEvent(userID, time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC), "Dec 31")
	calendar.CreateEvent(userID, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), "Jan 1")

	events := calendar.GetEventsForMonth(userID, dec15)
	if len(events) != 3 {
		t.Errorf("Expected 3 events for the month, got %d", len(events))
	}
}

func TestConcurrentAccess(t *testing.T) {
	calendar := NewCalendar()
	userID := 1
	date := time.Now()

	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			calendar.CreateEvent(userID, date, "Event")
			calendar.GetEventsForDay(userID, date)
		}(i)
	}

	wg.Wait()

	events := calendar.GetEventsForDay(userID, date)
	if len(events) != numGoroutines {
		t.Errorf("Expected %d events, got %d", numGoroutines, len(events))
	}
}

func TestEventIDIncrement(t *testing.T) {
	calendar := NewCalendar()
	userID := 1
	date := time.Now()

	for i := 0; i < 3; i++ {
		calendar.CreateEvent(userID, date, "Event")
	}

	events := calendar.GetEventsForDay(userID, date)
	if len(events) != 3 {
		t.Fatalf("Expected 3 events, got %d", len(events))
	}

	ids := make(map[int]bool)
	for _, event := range events {
		ids[event.ID] = true
	}

	for i := 1; i <= 3; i++ {
		if !ids[i] {
			t.Errorf("Expected ID %d not found", i)
		}
	}
}

func TestEmptyCalendar(t *testing.T) {
	calendar := NewCalendar()
	userID := 1
	date := time.Now()

	dayEvents := calendar.GetEventsForDay(userID, date)
	weekEvents := calendar.GetEventsForWeek(userID, date)
	monthEvents := calendar.GetEventsForMonth(userID, date)

	if len(dayEvents) != 0 {
		t.Errorf("Expected 0 day events, got %d", len(dayEvents))
	}
	if len(weekEvents) != 0 {
		t.Errorf("Expected 0 week events, got %d", len(weekEvents))
	}
	if len(monthEvents) != 0 {
		t.Errorf("Expected 0 month events, got %d", len(monthEvents))
	}
}
