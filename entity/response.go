package entity

import "time"

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

type PaginationMeta struct {
	CurrentPage int   `json:"current_page"`
	TotalPages  int   `json:"total_pages"`
	TotalItems  int64 `json:"total_items"`
	Limit       int   `json:"limit"`
}

type PaginatedResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Data    interface{}    `json:"data"`
	Meta    PaginationMeta `json:"meta"`
}

type Pagination struct {
	Page  int `form:"page" json:"page"`
	Limit int `form:"limit" json:"limit"`
}

func (p *Pagination) GetOffset() int {
	if p.Page <= 0 {
		p.Page = 1
	}
	return (p.Page - 1) * p.GetLimit()
}

func (p *Pagination) GetLimit() int {
	if p.Limit <= 0 {
		p.Limit = 10
	}
	if p.Limit > 100 {
		p.Limit = 100
	}
	return p.Limit
}

type Search struct {
	Query string `form:"q" json:"query"`
}

// Report structures
type ReportSummary struct {
	TotalTicketsSold int     `json:"total_tickets_sold"`
	TotalRevenue     float64 `json:"total_revenue"`
	TotalEvents      int     `json:"total_events"`
	ActiveEvents     int     `json:"active_events"`
	TotalUsers       int     `json:"total_users"`
	GeneratedAt      time.Time `json:"generated_at"`
}

type EventReport struct {
	EventID       string  `json:"event_id"`
	EventName     string  `json:"event_name"`
	TicketsSold   int     `json:"tickets_sold"`
	Revenue       float64 `json:"revenue"`
	Capacity      int     `json:"capacity"`
	Available     int     `json:"available"`
	SalesRate     float64 `json:"sales_rate"` // Percentage of tickets sold
}

type DateRangeFilter struct {
	StartDate *time.Time `form:"start_date" json:"start_date"`
	EndDate   *time.Time `form:"end_date" json:"end_date"`
} 