package entity

import "time"

type Event struct {
	ID     int       `json:"id"`
	UserID int       `json:"user_id" binding:"required"`
	Date   time.Time `json:"date" binding:"required"`
	Title  string    `json:"title" binding:"required"`
}

type CreateEventRequest struct {
	UserID int    `json:"user_id" binding:"required"`
	Date   string `json:"date" binding:"required"`
	Title  string `json:"title" binding:"required"`
}

type UpdateEventRequest struct {
	ID     int    `json:"id" binding:"required"`
	UserID int    `json:"user_id" binding:"required"`
	Date   string `json:"date" binding:"required"`
	Title  string `json:"title" binding:"required"`
}

type DeleteEventRequest struct {
	ID     int `json:"id" binding:"required"`
	UserID int `json:"user_id" binding:"required"`
}

type EventsResponse struct {
	Result []Event `json:"result"`
}

type SuccessResponse struct {
	Result string `json:"result"`
	Event  *Event `json:"event,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
