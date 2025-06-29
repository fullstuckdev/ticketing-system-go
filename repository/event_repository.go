package repository

import (
	"ticketing-system/entity"
	"time"

	"gorm.io/gorm"
)

type EventRepository interface {
	Create(event *entity.Event) error
	GetByID(id string) (*entity.Event, error)
	GetByName(name string) (*entity.Event, error)
	Update(event *entity.Event) error
	Delete(id string) error
	GetAll(pagination *entity.Pagination, search *entity.Search, filter *entity.EventFilter) ([]entity.Event, int64, error)
	GetActiveEvents() ([]entity.Event, error)
	UpdateAvailableTickets(eventID string, quantity int) error
	GetUpcomingEvents(limit int) ([]entity.Event, error)
}

type eventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) EventRepository {
	return &eventRepository{db: db}
}

func (r *eventRepository) Create(event *entity.Event) error {
	return r.db.Create(event).Error
}

func (r *eventRepository) GetByID(id string) (*entity.Event, error) {
	var event entity.Event
	err := r.db.Where("id = ?", id).First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *eventRepository) GetByName(name string) (*entity.Event, error) {
	var event entity.Event
	err := r.db.Where("name = ?", name).First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *eventRepository) Update(event *entity.Event) error {
	return r.db.Save(event).Error
}

func (r *eventRepository) Delete(id string) error {
	return r.db.Delete(&entity.Event{}, "id = ?", id).Error
}

func (r *eventRepository) GetAll(pagination *entity.Pagination, search *entity.Search, filter *entity.EventFilter) ([]entity.Event, int64, error) {
	var events []entity.Event
	var total int64

	query := r.db.Model(&entity.Event{})

	// Apply search filter
	if search != nil && search.Query != "" {
		searchQuery := "%" + search.Query + "%"
		query = query.Where("name LIKE ? OR description LIKE ? OR location LIKE ?", 
			searchQuery, searchQuery, searchQuery)
	}

	// Apply filters
	if filter != nil {
		if filter.Category != "" {
			query = query.Where("category = ?", filter.Category)
		}
		if filter.Status != "" {
			query = query.Where("status = ?", filter.Status)
		}
		if filter.Location != "" {
			query = query.Where("location LIKE ?", "%"+filter.Location+"%")
		}
		if filter.MinPrice != nil {
			query = query.Where("price >= ?", *filter.MinPrice)
		}
		if filter.MaxPrice != nil {
			query = query.Where("price <= ?", *filter.MaxPrice)
		}
		if filter.StartDate != nil {
			query = query.Where("event_date >= ?", *filter.StartDate)
		}
		if filter.EndDate != nil {
			query = query.Where("event_date <= ?", *filter.EndDate)
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

	err := query.Find(&events).Error
	return events, total, err
}

func (r *eventRepository) GetActiveEvents() ([]entity.Event, error) {
	var events []entity.Event
	err := r.db.Where("status = ? AND available > 0", entity.EventStatusActive).
		Order("event_date ASC").
		Find(&events).Error
	return events, err
}

func (r *eventRepository) UpdateAvailableTickets(eventID string, quantity int) error {
	return r.db.Model(&entity.Event{}).
		Where("id = ?", eventID).
		UpdateColumn("available", gorm.Expr("available - ?", quantity)).Error
}

func (r *eventRepository) GetUpcomingEvents(limit int) ([]entity.Event, error) {
	var events []entity.Event
	err := r.db.Where("status = ? AND event_date > ?", entity.EventStatusActive, time.Now()).
		Order("event_date ASC").
		Limit(limit).
		Find(&events).Error
	return events, err
} 