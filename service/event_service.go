package service

import (
	"errors"
	"ticketing-system/entity"
	"ticketing-system/repository"
	"time"

	"gorm.io/gorm"
)

type EventService interface {
	CreateEvent(req *entity.CreateEventRequest) (*entity.Event, error)
	GetEventByID(id string) (*entity.Event, error)
	UpdateEvent(id string, req *entity.UpdateEventRequest) (*entity.Event, error)
	DeleteEvent(id string) error
	GetAllEvents(pagination *entity.Pagination, search *entity.Search, filter *entity.EventFilter) ([]entity.Event, *entity.PaginationMeta, error)
	GetActiveEvents() ([]entity.Event, error)
	GetUpcomingEvents(limit int) ([]entity.Event, error)
}

type eventService struct {
	eventRepo repository.EventRepository
}

func NewEventService(eventRepo repository.EventRepository) EventService {
	return &eventService{
		eventRepo: eventRepo,
	}
}

func (s *eventService) CreateEvent(req *entity.CreateEventRequest) (*entity.Event, error) {
	// Validate event date
	if req.EventDate.Before(time.Now()) {
		return nil, errors.New("event date cannot be in the past")
	}

	// Check if event name already exists
	existingEvent, err := s.eventRepo.GetByName(req.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existingEvent != nil {
		return nil, errors.New("event name already exists")
	}

	// Create event
	event := &entity.Event{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Capacity:    req.Capacity,
		Available:   req.Capacity,
		Price:       req.Price,
		Location:    req.Location,
		EventDate:   req.EventDate,
		Status:      entity.EventStatusActive,
	}

	if err := s.eventRepo.Create(event); err != nil {
		return nil, err
	}

	return event, nil
}

func (s *eventService) GetEventByID(id string) (*entity.Event, error) {
	return s.eventRepo.GetByID(id)
}

func (s *eventService) UpdateEvent(id string, req *entity.UpdateEventRequest) (*entity.Event, error) {
	event, err := s.eventRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Check if event can be modified
	if !event.CanBeModified() {
		return nil, errors.New("cannot modify event that is not active")
	}

	// Update fields if provided
	if req.Name != nil {
		// Check if new name is already taken
		if *req.Name != event.Name {
			existingEvent, err := s.eventRepo.GetByName(*req.Name)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}
			if existingEvent != nil {
				return nil, errors.New("event name already exists")
			}
		}
		event.Name = *req.Name
	}

	if req.Description != nil {
		event.Description = *req.Description
	}

	if req.Category != nil {
		event.Category = *req.Category
	}

	if req.Capacity != nil {
		if *req.Capacity < 0 {
			return nil, errors.New("capacity cannot be negative")
		}
		// Calculate new available tickets
		soldTickets := event.Capacity - event.Available
		if *req.Capacity < soldTickets {
			return nil, errors.New("cannot reduce capacity below sold tickets")
		}
		event.Available = *req.Capacity - soldTickets
		event.Capacity = *req.Capacity
	}

	if req.Price != nil {
		if *req.Price < 0 {
			return nil, errors.New("price cannot be negative")
		}
		event.Price = *req.Price
	}

	if req.Location != nil {
		event.Location = *req.Location
	}

	if req.EventDate != nil {
		if req.EventDate.Before(time.Now()) {
			return nil, errors.New("event date cannot be in the past")
		}
		event.EventDate = *req.EventDate
	}

	if err := s.eventRepo.Update(event); err != nil {
		return nil, err
	}

	return event, nil
}

func (s *eventService) DeleteEvent(id string) error {
	event, err := s.eventRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Check if event has sold tickets
	soldTickets := event.Capacity - event.Available
	if soldTickets > 0 {
		return errors.New("cannot delete event with sold tickets")
	}

	return s.eventRepo.Delete(id)
}

func (s *eventService) GetAllEvents(pagination *entity.Pagination, search *entity.Search, filter *entity.EventFilter) ([]entity.Event, *entity.PaginationMeta, error) {
	events, total, err := s.eventRepo.GetAll(pagination, search, filter)
	if err != nil {
		return nil, nil, err
	}

	meta := &entity.PaginationMeta{
		CurrentPage: pagination.Page,
		TotalItems:  total,
		Limit:       pagination.GetLimit(),
		TotalPages:  int((total + int64(pagination.GetLimit()) - 1) / int64(pagination.GetLimit())),
	}

	return events, meta, nil
}

func (s *eventService) GetActiveEvents() ([]entity.Event, error) {
	return s.eventRepo.GetActiveEvents()
}

func (s *eventService) GetUpcomingEvents(limit int) ([]entity.Event, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.eventRepo.GetUpcomingEvents(limit)
} 