package controller

import (
	"net/http"
	"ticketing-system/entity"
	"ticketing-system/service"

	"github.com/gin-gonic/gin"
)

type ReportController struct {
	ticketService service.TicketService
}

func NewReportController(ticketService service.TicketService) *ReportController {
	return &ReportController{ticketService: ticketService}
}

// GetSummaryReport godoc
// @Summary Get summary report (Admin only)
// @Description Get overall system summary including tickets sold, revenue, events, and users
// @Tags Reports
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} entity.Response{data=entity.ReportSummary}
// @Failure 401 {object} entity.Response
// @Failure 403 {object} entity.Response
// @Failure 500 {object} entity.Response
// @Router /reports/summary [get]
func (rc *ReportController) GetSummaryReport(c *gin.Context) {
	summary, err := rc.ticketService.GetTicketStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, entity.Response{
			Success: false,
			Message: "Failed to generate summary report",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, entity.Response{
		Success: true,
		Message: "Summary report generated successfully",
		Data:    summary,
	})
}

// GetEventReport godoc
// @Summary Get event report (Admin only)
// @Description Get detailed report for a specific event including sales metrics
// @Tags Reports
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Event ID"
// @Success 200 {object} entity.Response{data=entity.EventReport}
// @Failure 400 {object} entity.Response
// @Failure 401 {object} entity.Response
// @Failure 403 {object} entity.Response
// @Failure 404 {object} entity.Response
// @Failure 500 {object} entity.Response
// @Router /reports/event/{id} [get]
func (rc *ReportController) GetEventReport(c *gin.Context) {
	eventID := c.Param("id")
	if eventID == "" {
		c.JSON(http.StatusBadRequest, entity.Response{
			Success: false,
			Message: "Event ID is required",
		})
		return
	}

	report, err := rc.ticketService.GetEventReport(eventID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "record not found" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, entity.Response{
			Success: false,
			Message: "Failed to generate event report",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, entity.Response{
		Success: true,
		Message: "Event report generated successfully",
		Data:    report,
	})
} 