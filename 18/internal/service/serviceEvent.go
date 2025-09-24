package service

import (
	"sort"
	"time"

	"github.com/ds124wfegd/WB_L2/18/internal/entity"
)

func (c *Calendar) CreateEvent(userID int, date time.Time, title string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	event := entity.Event{
		ID:     c.nextID,
		UserID: userID,
		Date:   date,
		Title:  title,
	}

	c.events[c.nextID] = event
	c.nextID++

	return nil
}

func (c *Calendar) UpdateEvent(id, userID int, date time.Time, title string) (entity.Event, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	event, exists := c.events[id]
	if !exists {
		return entity.Event{}, ErrEventNotFound
	}

	if event.UserID != userID {
		return entity.Event{}, ErrEventNotFound
	}

	event.Date = date
	event.Title = title
	c.events[id] = event

	return event, nil
}

func (c *Calendar) DeleteEvent(id, userID int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	event, exists := c.events[id]
	if !exists {
		return ErrEventNotFound
	}

	if event.UserID != userID {
		return ErrEventNotFound
	}

	delete(c.events, id)
	return nil
}

func (c *Calendar) GetEventsForDay(userID int, date time.Time) []entity.Event {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var events []entity.Event
	for _, event := range c.events {
		if event.UserID == userID && isSameDay(event.Date, date) {
			events = append(events, event)
		}
	}

	sortEventsByDate(events)
	return events
}

func (c *Calendar) GetEventsForWeek(userID int, date time.Time) []entity.Event {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var events []entity.Event
	year, week := date.ISOWeek()

	for _, event := range c.events {
		if event.UserID == userID {
			eventYear, eventWeek := event.Date.ISOWeek()
			if eventYear == year && eventWeek == week {
				events = append(events, event)
			}
		}
	}

	sortEventsByDate(events)
	return events
}

func (c *Calendar) GetEventsForMonth(userID int, date time.Time) []entity.Event {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var events []entity.Event
	year, month := date.Year(), date.Month()

	for _, event := range c.events {
		if event.UserID == userID {
			eventYear, eventMonth := event.Date.Year(), event.Date.Month()
			if eventYear == year && eventMonth == month {
				events = append(events, event)
			}
		}
	}

	sortEventsByDate(events)
	return events
}

func isSameDay(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func sortEventsByDate(events []entity.Event) {
	sort.Slice(events, func(i, j int) bool {
		return events[i].Date.Before(events[j].Date)
	})
}
