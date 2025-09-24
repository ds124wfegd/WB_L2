package service

import (
	"errors"
	"sync"
	"time"

	"github.com/ds124wfegd/WB_L2/18/internal/entity"
)

type EventService interface {
	CreateEvent(userID int, date time.Time, title string) (entity.Event, error)
	UpdateEvent(id, userID int, date time.Time, title string) (entity.Event, error)
	DeleteEvent(id, userID int) error
	GetEventsForDay(userID int, date time.Time) []entity.Event
	GetEventsForWeek(userID int, date time.Time) []entity.Event
	GetEventsForMonth(userID int, date time.Time) []entity.Event
}

var (
	ErrEventNotFound = errors.New("event not found")
	ErrInvalidDate   = errors.New("invalid date")
	ErrDuplicateID   = errors.New("event with this ID already exists")
)

type Calendar struct {
	mu         sync.RWMutex
	events     map[int]entity.Event
	userEvents map[int][]int
	dateEvents map[time.Time][]int
	nextID     int
}

func NewCalendar() *Calendar {
	return &Calendar{
		events:     make(map[int]entity.Event),
		userEvents: make(map[int][]int),
		dateEvents: make(map[time.Time][]int),
		nextID:     1,
	}
}
