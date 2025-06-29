package controller

import (
	"net/http"
	"ticketing-system/entity"
	"ticketing-system/service"

	"github.com/gin-gonic/gin"
)

type EventController struct {
	eventService service.EventService
}

func NewEventController(eventService service.EventService) *EventController {
	return &EventController{eventService: eventService}
}

// GetAllEvents godoc
// @Summary Get all events
// @Description Get list of events with pagination, search, and filtering
// @Tags Events
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param q query string false "Search query"
// @Param category query string false "Filter by category"
// @Param status query string false "Filter by status"
// @Param location query string false "Filter by location"
// @Param min_price query number false "Minimum price filter"
// @Param max_price query number false "Maximum price filter"
// @Param start_date query string false "Start date filter (RFC3339)"
// @Param end_date query string false "End date filter (RFC3339)"
// @Success 200 {object} entity.PaginatedResponse{data=[]entity.Event}
// @Failure 400 {object} entity.Response
// @Router /events [get]
func (ec *EventController) GetAllEvents(c *gin.Context) {
	var pagination entity.Pagination
	var search entity.Search
	var filter entity.EventFilter

	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.JSON(http.StatusBadRequest, entity.Response{
			Success: false,
			Message: "Invalid pagination parameters",
			Error:   err.Error(),
		})
		return
	}

	if err := c.ShouldBindQuery(&search); err != nil {
		c.JSON(http.StatusBadRequest, entity.Response{
			Success: false,
			Message: "Invalid search parameters",
			Error:   err.Error(),
		})
		return
	}

	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, entity.Response{
			Success: false,
			Message: "Invalid filter parameters",
			Error:   err.Error(),
		})
		return
	}

	events, meta, err := ec.eventService.GetAllEvents(&pagination, &search, &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, entity.Response{
			Success: false,
			Message: "Failed to retrieve events",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, entity.PaginatedResponse{
		Success: true,
		Message: "Events retrieved successfully",
		Data:    events,
		Meta:    *meta,
	})
}

// GetEventByID godoc
// @Summary Get event by ID
// @Description Get a single event by its ID
// @Tags Events
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {object} entity.Response{data=entity.Event}
// @Failure 404 {object} entity.Response
// @Router /events/{id} [get]
func (ec *EventController) GetEventByID(c *gin.Context) {
	eventID := c.Param("id")
	if eventID == "" {
		c.JSON(http.StatusBadRequest, entity.Response{
			Success: false,
			Message: "Event ID is required",
		})
		return
	}

	event, err := ec.eventService.GetEventByID(eventID)
	if err != nil {
		c.JSON(http.StatusNotFound, entity.Response{
			Success: false,
			Message: "Event not found",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, entity.Response{
		Success: true,
		Message: "Event retrieved successfully",
		Data:    event,
	})
}

// CreateEvent godoc
// @Summary Create new event (Admin only)
// @Description Create a new event
// @Tags Events
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body entity.CreateEventRequest true "Event data"
// @Success 201 {object} entity.Response{data=entity.Event}
// @Failure 400 {object} entity.Response
// @Failure 401 {object} entity.Response
// @Failure 403 {object} entity.Response
// @Failure 409 {object} entity.Response
// @Router /events [post]
func (ec *EventController) CreateEvent(c *gin.Context) {
	var req entity.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, entity.Response{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	event, err := ec.eventService.CreateEvent(&req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "event name already exists" {
			statusCode = http.StatusConflict
		} else if err.Error() == "event date cannot be in the past" {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, entity.Response{
			Success: false,
			Message: "Failed to create event",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, entity.Response{
		Success: true,
		Message: "Event created successfully",
		Data:    event,
	})
}

// UpdateEvent godoc
// @Summary Update event (Admin only)
// @Description Update an existing event
// @Tags Events
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Event ID"
// @Param request body entity.UpdateEventRequest true "Event update data"
// @Success 200 {object} entity.Response{data=entity.Event}
// @Failure 400 {object} entity.Response
// @Failure 401 {object} entity.Response
// @Failure 403 {object} entity.Response
// @Failure 404 {object} entity.Response
// @Failure 409 {object} entity.Response
// @Router /events/{id} [put]
func (ec *EventController) UpdateEvent(c *gin.Context) {
	eventID := c.Param("id")
	if eventID == "" {
		c.JSON(http.StatusBadRequest, entity.Response{
			Success: false,
			Message: "Event ID is required",
		})
		return
	}

	var req entity.UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, entity.Response{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	event, err := ec.eventService.UpdateEvent(eventID, &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "event name already exists" {
			statusCode = http.StatusConflict
		} else if err.Error() == "cannot modify event that is not active" ||
			err.Error() == "capacity cannot be negative" ||
			err.Error() == "price cannot be negative" ||
			err.Error() == "cannot reduce capacity below sold tickets" ||
			err.Error() == "event date cannot be in the past" {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, entity.Response{
			Success: false,
			Message: "Failed to update event",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, entity.Response{
		Success: true,
		Message: "Event updated successfully",
		Data:    event,
	})
}

// DeleteEvent godoc
// @Summary Delete event (Admin only)
// @Description Delete an event
// @Tags Events
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Event ID"
// @Success 200 {object} entity.Response
// @Failure 400 {object} entity.Response
// @Failure 401 {object} entity.Response
// @Failure 403 {object} entity.Response
// @Failure 404 {object} entity.Response
// @Router /events/{id} [delete]
func (ec *EventController) DeleteEvent(c *gin.Context) {
	eventID := c.Param("id")
	if eventID == "" {
		c.JSON(http.StatusBadRequest, entity.Response{
			Success: false,
			Message: "Event ID is required",
		})
		return
	}

	err := ec.eventService.DeleteEvent(eventID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "cannot delete event with sold tickets" {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, entity.Response{
			Success: false,
			Message: "Failed to delete event",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, entity.Response{
		Success: true,
		Message: "Event deleted successfully",
	})
}

// GetActiveEvents godoc
// @Summary Get active events
// @Description Get list of active events available for booking
// @Tags Events
// @Accept json
// @Produce json
// @Success 200 {object} entity.Response{data=[]entity.Event}
// @Failure 500 {object} entity.Response
// @Router /events/active [get]
func (ec *EventController) GetActiveEvents(c *gin.Context) {
	events, err := ec.eventService.GetActiveEvents()
	if err != nil {
		c.JSON(http.StatusInternalServerError, entity.Response{
			Success: false,
			Message: "Failed to retrieve active events",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, entity.Response{
		Success: true,
		Message: "Active events retrieved successfully",
		Data:    events,
	})
}

// GetUpcomingEvents godoc
// @Summary Get upcoming events
// @Description Get list of upcoming events
// @Tags Events
// @Accept json
// @Produce json
// @Param limit query int false "Number of events to return" default(10)
// @Success 200 {object} entity.Response{data=[]entity.Event}
// @Failure 500 {object} entity.Response
// @Router /events/upcoming [get]
func (ec *EventController) GetUpcomingEvents(c *gin.Context) {
	limit := 10
	if limitParam := c.Query("limit"); limitParam != "" {
		// Parse limit parameter
		// For simplicity, using default if parsing fails
	}

	events, err := ec.eventService.GetUpcomingEvents(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, entity.Response{
			Success: false,
			Message: "Failed to retrieve upcoming events",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, entity.Response{
		Success: true,
		Message: "Upcoming events retrieved successfully",
		Data:    events,
	})
} 