package controller

import (
	"net/http"
	"ticketing-system/entity"
	"ticketing-system/middleware"
	"ticketing-system/service"

	"github.com/gin-gonic/gin"
)

type TicketController struct {
	ticketService service.TicketService
}

func NewTicketController(ticketService service.TicketService) *TicketController {
	return &TicketController{ticketService: ticketService}
}

// BuyTicket godoc
// @Summary Buy tickets
// @Description Purchase tickets for an event
// @Tags Tickets
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body entity.BuyTicketRequest true "Ticket purchase data"
// @Success 201 {object} entity.Response{data=entity.Ticket}
// @Failure 400 {object} entity.Response
// @Failure 401 {object} entity.Response
// @Router /tickets [post]
func (tc *TicketController) BuyTicket(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, entity.Response{
			Success: false,
			Message: "Authentication required",
		})
		return
	}

	var req entity.BuyTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, entity.Response{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	ticket, err := tc.ticketService.BuyTicket(userID, &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "user account is not active" ||
			err.Error() == "event is not available for booking" ||
			err.Error() == "insufficient tickets available" ||
			err.Error() == "cannot purchase tickets for events starting within an hour" {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, entity.Response{
			Success: false,
			Message: "Failed to purchase ticket",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, entity.Response{
		Success: true,
		Message: "Ticket purchased successfully",
		Data:    ticket,
	})
}

// GetAllTickets godoc
// @Summary Get all tickets (Admin only)
// @Description Get list of all tickets with pagination, search, and filtering
// @Tags Tickets
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param q query string false "Search query"
// @Param user_id query string false "Filter by user ID"
// @Param event_id query string false "Filter by event ID"
// @Param status query string false "Filter by status"
// @Param start_date query string false "Start date filter (RFC3339)"
// @Param end_date query string false "End date filter (RFC3339)"
// @Success 200 {object} entity.PaginatedResponse{data=[]entity.Ticket}
// @Failure 400 {object} entity.Response
// @Failure 401 {object} entity.Response
// @Failure 403 {object} entity.Response
// @Router /tickets [get]
func (tc *TicketController) GetAllTickets(c *gin.Context) {
	var pagination entity.Pagination
	var search entity.Search
	var filter entity.TicketFilter

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

	tickets, meta, err := tc.ticketService.GetAllTickets(&pagination, &search, &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, entity.Response{
			Success: false,
			Message: "Failed to retrieve tickets",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, entity.PaginatedResponse{
		Success: true,
		Message: "Tickets retrieved successfully",
		Data:    tickets,
		Meta:    *meta,
	})
}

// GetUserTickets godoc
// @Summary Get user's tickets
// @Description Get current user's tickets
// @Tags Tickets
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} entity.PaginatedResponse{data=[]entity.Ticket}
// @Failure 400 {object} entity.Response
// @Failure 401 {object} entity.Response
// @Router /tickets/my [get]
func (tc *TicketController) GetUserTickets(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, entity.Response{
			Success: false,
			Message: "Authentication required",
		})
		return
	}

	var pagination entity.Pagination
	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.JSON(http.StatusBadRequest, entity.Response{
			Success: false,
			Message: "Invalid pagination parameters",
			Error:   err.Error(),
		})
		return
	}

	tickets, meta, err := tc.ticketService.GetUserTickets(userID, &pagination)
	if err != nil {
		c.JSON(http.StatusInternalServerError, entity.Response{
			Success: false,
			Message: "Failed to retrieve tickets",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, entity.PaginatedResponse{
		Success: true,
		Message: "User tickets retrieved successfully",
		Data:    tickets,
		Meta:    *meta,
	})
}

// GetTicketByID godoc
// @Summary Get ticket by ID
// @Description Get a single ticket by its ID
// @Tags Tickets
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Ticket ID"
// @Success 200 {object} entity.Response{data=entity.Ticket}
// @Failure 401 {object} entity.Response
// @Failure 403 {object} entity.Response
// @Failure 404 {object} entity.Response
// @Router /tickets/{id} [get]
func (tc *TicketController) GetTicketByID(c *gin.Context) {
	ticketID := c.Param("id")
	if ticketID == "" {
		c.JSON(http.StatusBadRequest, entity.Response{
			Success: false,
			Message: "Ticket ID is required",
		})
		return
	}

	ticket, err := tc.ticketService.GetTicketByID(ticketID)
	if err != nil {
		c.JSON(http.StatusNotFound, entity.Response{
			Success: false,
			Message: "Ticket not found",
			Error:   err.Error(),
		})
		return
	}

	// Check if user can access this ticket (own ticket or admin)
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, entity.Response{
			Success: false,
			Message: "Authentication required",
		})
		return
	}

	if !currentUser.IsAdmin() && ticket.UserID != currentUser.ID {
		c.JSON(http.StatusForbidden, entity.Response{
			Success: false,
			Message: "Access denied: You can only view your own tickets",
		})
		return
	}

	c.JSON(http.StatusOK, entity.Response{
		Success: true,
		Message: "Ticket retrieved successfully",
		Data:    ticket,
	})
}

// UpdateTicketStatus godoc
// @Summary Update ticket status (Admin only)
// @Description Update the status of a ticket
// @Tags Tickets
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Ticket ID"
// @Param request body entity.UpdateTicketStatusRequest true "Status update data"
// @Success 200 {object} entity.Response{data=entity.Ticket}
// @Failure 400 {object} entity.Response
// @Failure 401 {object} entity.Response
// @Failure 403 {object} entity.Response
// @Failure 404 {object} entity.Response
// @Router /tickets/{id} [patch]
func (tc *TicketController) UpdateTicketStatus(c *gin.Context) {
	ticketID := c.Param("id")
	if ticketID == "" {
		c.JSON(http.StatusBadRequest, entity.Response{
			Success: false,
			Message: "Ticket ID is required",
		})
		return
	}

	var req entity.UpdateTicketStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, entity.Response{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
		return
	}

	ticket, err := tc.ticketService.UpdateTicketStatus(ticketID, &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "cannot update cancelled ticket" ||
			err.Error() == "can only mark active tickets as used" {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, entity.Response{
			Success: false,
			Message: "Failed to update ticket status",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, entity.Response{
		Success: true,
		Message: "Ticket status updated successfully",
		Data:    ticket,
	})
}

// CancelTicket godoc
// @Summary Cancel ticket
// @Description Cancel a user's ticket
// @Tags Tickets
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Ticket ID"
// @Success 200 {object} entity.Response{data=entity.Ticket}
// @Failure 400 {object} entity.Response
// @Failure 401 {object} entity.Response
// @Failure 403 {object} entity.Response
// @Failure 404 {object} entity.Response
// @Router /tickets/{id}/cancel [patch]
func (tc *TicketController) CancelTicket(c *gin.Context) {
	ticketID := c.Param("id")
	if ticketID == "" {
		c.JSON(http.StatusBadRequest, entity.Response{
			Success: false,
			Message: "Ticket ID is required",
		})
		return
	}

	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, entity.Response{
			Success: false,
			Message: "Authentication required",
		})
		return
	}

	ticket, err := tc.ticketService.CancelTicket(ticketID, userID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "you can only cancel your own tickets" {
			statusCode = http.StatusForbidden
		} else if err.Error() == "ticket cannot be cancelled" ||
			err.Error() == "cannot cancel tickets within 2 hours of event start" {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, entity.Response{
			Success: false,
			Message: "Failed to cancel ticket",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, entity.Response{
		Success: true,
		Message: "Ticket cancelled successfully",
		Data:    ticket,
	})
} 