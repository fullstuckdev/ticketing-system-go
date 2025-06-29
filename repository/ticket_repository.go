package repository

import (
	"ticketing-system/entity"
	"time"

	"gorm.io/gorm"
)

type TicketRepository interface {
	Create(ticket *entity.Ticket) error
	CreateWithTx(tx *gorm.DB, ticket *entity.Ticket) error
	GetByID(id string) (*entity.Ticket, error)
	GetByIDWithTx(tx *gorm.DB, id string) (*entity.Ticket, error)
	Update(ticket *entity.Ticket) error
	UpdateWithTx(tx *gorm.DB, ticket *entity.Ticket) error
	Delete(id string) error
	GetAll(pagination *entity.Pagination, search *entity.Search, filter *entity.TicketFilter) ([]entity.Ticket, int64, error)
	GetByUserID(userID string, pagination *entity.Pagination) ([]entity.Ticket, int64, error)
	GetByEventID(eventID string, pagination *entity.Pagination) ([]entity.Ticket, int64, error)
	GetTicketStats() (*entity.ReportSummary, error)
	GetEventReport(eventID string) (*entity.EventReport, error)
	GetRevenueByDateRange(startDate, endDate time.Time) (float64, error)
	GetTicketsSoldByDateRange(startDate, endDate time.Time) (int, error)
}

type ticketRepository struct {
	db *gorm.DB
}

func NewTicketRepository(db *gorm.DB) TicketRepository {
	return &ticketRepository{db: db}
}

func (r *ticketRepository) Create(ticket *entity.Ticket) error {
	return r.db.Create(ticket).Error
}

func (r *ticketRepository) CreateWithTx(tx *gorm.DB, ticket *entity.Ticket) error {
	return tx.Create(ticket).Error
}

func (r *ticketRepository) GetByID(id string) (*entity.Ticket, error) {
	var ticket entity.Ticket
	err := r.db.Preload("User").Preload("Event").Where("id = ?", id).First(&ticket).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *ticketRepository) GetByIDWithTx(tx *gorm.DB, id string) (*entity.Ticket, error) {
	var ticket entity.Ticket
	err := tx.Preload("User").Preload("Event").Where("id = ?", id).First(&ticket).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *ticketRepository) Update(ticket *entity.Ticket) error {
	return r.db.Save(ticket).Error
}

func (r *ticketRepository) UpdateWithTx(tx *gorm.DB, ticket *entity.Ticket) error {
	return tx.Save(ticket).Error
}

func (r *ticketRepository) Delete(id string) error {
	return r.db.Delete(&entity.Ticket{}, "id = ?", id).Error
}

func (r *ticketRepository) GetAll(pagination *entity.Pagination, search *entity.Search, filter *entity.TicketFilter) ([]entity.Ticket, int64, error) {
	var tickets []entity.Ticket
	var total int64

	query := r.db.Model(&entity.Ticket{}).Preload("User").Preload("Event")

	// Apply search filter
	if search != nil && search.Query != "" {
		searchQuery := "%" + search.Query + "%"
		query = query.Joins("LEFT JOIN users ON tickets.user_id = users.id").
			Joins("LEFT JOIN events ON tickets.event_id = events.id").
			Where("users.name LIKE ? OR users.email LIKE ? OR events.name LIKE ?", 
				searchQuery, searchQuery, searchQuery)
	}

	// Apply filters
	if filter != nil {
		if filter.UserID != "" {
			query = query.Where("user_id = ?", filter.UserID)
		}
		if filter.EventID != "" {
			query = query.Where("event_id = ?", filter.EventID)
		}
		if filter.Status != "" {
			query = query.Where("status = ?", filter.Status)
		}
		if filter.StartDate != nil {
			query = query.Where("purchase_date >= ?", *filter.StartDate)
		}
		if filter.EndDate != nil {
			query = query.Where("purchase_date <= ?", *filter.EndDate)
		}
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and ordering
	if pagination != nil {
		query = query.Offset(pagination.GetOffset()).Limit(pagination.GetLimit())
	}
	
	query = query.Order("created_at DESC")

	err := query.Find(&tickets).Error
	return tickets, total, err
}

func (r *ticketRepository) GetByUserID(userID string, pagination *entity.Pagination) ([]entity.Ticket, int64, error) {
	var tickets []entity.Ticket
	var total int64

	query := r.db.Model(&entity.Ticket{}).Preload("Event").Where("user_id = ?", userID)

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if pagination != nil {
		query = query.Offset(pagination.GetOffset()).Limit(pagination.GetLimit())
	}

	err := query.Order("created_at DESC").Find(&tickets).Error
	return tickets, total, err
}

func (r *ticketRepository) GetByEventID(eventID string, pagination *entity.Pagination) ([]entity.Ticket, int64, error) {
	var tickets []entity.Ticket
	var total int64

	query := r.db.Model(&entity.Ticket{}).Preload("User").Where("event_id = ?", eventID)

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if pagination != nil {
		query = query.Offset(pagination.GetOffset()).Limit(pagination.GetLimit())
	}

	err := query.Order("created_at DESC").Find(&tickets).Error
	return tickets, total, err
}

func (r *ticketRepository) GetTicketStats() (*entity.ReportSummary, error) {
	var summary entity.ReportSummary

	// Get total tickets sold
	var totalTickets int64
	if err := r.db.Model(&entity.Ticket{}).Where("status != ?", entity.TicketStatusCancelled).Count(&totalTickets).Error; err != nil {
		return nil, err
	}
	summary.TotalTicketsSold = int(totalTickets)

	// Get total revenue
	var totalRevenue float64
	if err := r.db.Model(&entity.Ticket{}).Where("status != ?", entity.TicketStatusCancelled).
		Select("COALESCE(SUM(total_price), 0)").Row().Scan(&totalRevenue); err != nil {
		return nil, err
	}
	summary.TotalRevenue = totalRevenue

	// Get total events
	var totalEvents int64
	if err := r.db.Model(&entity.Event{}).Count(&totalEvents).Error; err != nil {
		return nil, err
	}
	summary.TotalEvents = int(totalEvents)

	// Get active events
	var activeEvents int64
	if err := r.db.Model(&entity.Event{}).Where("status = ?", entity.EventStatusActive).Count(&activeEvents).Error; err != nil {
		return nil, err
	}
	summary.ActiveEvents = int(activeEvents)

	// Get total users
	var totalUsers int64
	if err := r.db.Model(&entity.User{}).Count(&totalUsers).Error; err != nil {
		return nil, err
	}
	summary.TotalUsers = int(totalUsers)

	summary.GeneratedAt = time.Now()

	return &summary, nil
}

func (r *ticketRepository) GetEventReport(eventID string) (*entity.EventReport, error) {
	var report entity.EventReport

	// Get event details
	var event entity.Event
	if err := r.db.Where("id = ?", eventID).First(&event).Error; err != nil {
		return nil, err
	}

	// Get tickets sold count
	var ticketsSold int64
	if err := r.db.Model(&entity.Ticket{}).Where("event_id = ? AND status != ?", eventID, entity.TicketStatusCancelled).Count(&ticketsSold).Error; err != nil {
		return nil, err
	}

	// Get total revenue
	var revenue float64
	if err := r.db.Model(&entity.Ticket{}).Where("event_id = ? AND status != ?", eventID, entity.TicketStatusCancelled).
		Select("COALESCE(SUM(total_price), 0)").Row().Scan(&revenue); err != nil {
		return nil, err
	}

	// Calculate sales rate
	salesRate := float64(0)
	if event.Capacity > 0 {
		salesRate = (float64(ticketsSold) / float64(event.Capacity)) * 100
	}

	report = entity.EventReport{
		EventID:     event.ID,
		EventName:   event.Name,
		TicketsSold: int(ticketsSold),
		Revenue:     revenue,
		Capacity:    event.Capacity,
		Available:   event.Available,
		SalesRate:   salesRate,
	}

	return &report, nil
}

func (r *ticketRepository) GetRevenueByDateRange(startDate, endDate time.Time) (float64, error) {
	var revenue float64
	err := r.db.Model(&entity.Ticket{}).
		Where("purchase_date BETWEEN ? AND ? AND status != ?", startDate, endDate, entity.TicketStatusCancelled).
		Select("COALESCE(SUM(total_price), 0)").Row().Scan(&revenue)
	return revenue, err
}

func (r *ticketRepository) GetTicketsSoldByDateRange(startDate, endDate time.Time) (int, error) {
	var count int64
	err := r.db.Model(&entity.Ticket{}).
		Where("purchase_date BETWEEN ? AND ? AND status != ?", startDate, endDate, entity.TicketStatusCancelled).
		Count(&count).Error
	return int(count), err
} 