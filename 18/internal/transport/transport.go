package transport

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ds124wfegd/WB_L2/18/internal/entity"
	"github.com/ds124wfegd/WB_L2/18/internal/service"

	"github.com/gin-gonic/gin"
)

type EventHandler struct {
	calendar *service.Calendar
}

func NewEventHandlers(calendar *service.Calendar) *EventHandler {
	return &EventHandler{calendar: calendar}
}

func (h *EventHandler) InitRoutes() *gin.Engine {
	router := gin.New()

	router.POST("/create_event", h.CreateEvent)
	router.POST("/update_event", h.UpdateEvent)
	router.POST("/delete_event", h.DeleteEvent)
	router.GET("/events_for_day", h.GetEventsForDay)
	router.GET("/events_for_week", h.GetEventsForWeek)
	router.GET("/events_for_month", h.GetEventsForMonth)

	return router
}

func (h *EventHandler) CreateEvent(c *gin.Context) {
	var req entity.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, entity.ErrorResponse{Error: err.Error()})
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, entity.ErrorResponse{Error: "invalid date format, use YYYY-MM-DD"})
		return
	}

	err = h.calendar.CreateEvent(req.UserID, date, req.Title)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, entity.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, entity.SuccessResponse{Result: "Event created successfully"})
}

func (h *EventHandler) UpdateEvent(c *gin.Context) {
	var req entity.UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, entity.ErrorResponse{Error: err.Error()})
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, entity.ErrorResponse{Error: "invalid date format, use YYYY-MM-DD"})
		return
	}

	_, err = h.calendar.UpdateEvent(req.ID, req.UserID, date, req.Title)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, entity.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, entity.SuccessResponse{Result: "Event updated successfully"})
}

func (h *EventHandler) DeleteEvent(c *gin.Context) {
	var req entity.DeleteEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, entity.ErrorResponse{Error: err.Error()})
		return
	}

	err := h.calendar.DeleteEvent(req.ID, req.UserID)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, entity.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, entity.SuccessResponse{Result: "Event deleted successfully"})
}

func (h *EventHandler) GetEventsForDay(c *gin.Context) {
	h.getEventsForPeriod(c, "day")
}

func (h *EventHandler) GetEventsForWeek(c *gin.Context) {
	h.getEventsForPeriod(c, "week")
}

func (h *EventHandler) GetEventsForMonth(c *gin.Context) {
	h.getEventsForPeriod(c, "month")
}

func (h *EventHandler) getEventsForPeriod(c *gin.Context, period string) {
	userID, err := parseIntParam(c, "user_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, entity.ErrorResponse{Error: "invalid user_id"})
		return
	}

	dateStr := c.Query("date")
	if dateStr == "" {
		c.JSON(http.StatusBadRequest, entity.ErrorResponse{Error: "date parameter is required"})
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, entity.ErrorResponse{Error: "invalid date format, use YYYY-MM-DD"})
		return
	}

	var events []entity.Event
	switch period {
	case "day":
		events = h.calendar.GetEventsForDay(userID, date)
	case "week":
		events = h.calendar.GetEventsForWeek(userID, date)
	case "month":
		events = h.calendar.GetEventsForMonth(userID, date)
	}

	c.JSON(http.StatusOK, entity.EventsResponse{Result: events})
}

func parseIntParam(c *gin.Context, param string) (int, error) {
	value := c.Query(param)
	if value == "" {
		return 0, fmt.Errorf("%s is required", param)
	}

	var result int
	_, err := fmt.Sscanf(value, "%d", &result)
	return result, err
}
