package service

import (
	"errors"
	"ticketing-system/entity"
	"ticketing-system/repository"
	"time"

	"gorm.io/gorm"
)

type TicketService interface {
	BuyTicket(userID string, req *entity.BuyTicketRequest) (*entity.Ticket, error)
	GetTicketByID(id string) (*entity.Ticket, error)
	GetUserTickets(userID string, pagination *entity.Pagination) ([]entity.Ticket, *entity.PaginationMeta, error)
	GetAllTickets(pagination *entity.Pagination, search *entity.Search, filter *entity.TicketFilter) ([]entity.Ticket, *entity.PaginationMeta, error)
	UpdateTicketStatus(ticketID string, req *entity.UpdateTicketStatusRequest) (*entity.Ticket, error)
	CancelTicket(ticketID, userID string) (*entity.Ticket, error)
	GetTicketStats() (*entity.ReportSummary, error)
	GetEventReport(eventID string) (*entity.EventReport, error)
}

type ticketService struct {
	ticketRepo repository.TicketRepository
	eventRepo  repository.EventRepository
	userRepo   repository.UserRepository
	db         *gorm.DB
}

func NewTicketService(
	ticketRepo repository.TicketRepository,
	eventRepo repository.EventRepository,
	userRepo repository.UserRepository,
	db *gorm.DB,
) TicketService {
	return &ticketService{
		ticketRepo: ticketRepo,
		eventRepo:  eventRepo,
		userRepo:   userRepo,
		db:         db,
	}
}

func (s *ticketService) BuyTicket(userID string, req *entity.BuyTicketRequest) (*entity.Ticket, error) {
	var ticket *entity.Ticket
	var err error

	// Start transaction
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Validate user
		user, err := s.userRepo.GetByID(userID)
		if err != nil {
			return err
		}
		if !user.IsActive {
			return errors.New("user account is not active")
		}

		// Validate event with SELECT FOR UPDATE to prevent race conditions
		var event entity.Event
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", req.EventID).First(&event).Error; err != nil {
			return err
		}

		// Check event availability
		if !event.IsAvailable() {
			return errors.New("event is not available for booking")
		}

		// Check capacity
		if event.Available < req.Quantity {
			return errors.New("insufficient tickets available")
		}

		// Check if event date is in the future
		if event.EventDate.Before(time.Now().Add(time.Hour)) {
			return errors.New("cannot purchase tickets for events starting within an hour")
		}

		// Calculate total price
		totalPrice := event.Price * float64(req.Quantity)

		// Create ticket
		ticket = &entity.Ticket{
			UserID:       userID,
			EventID:      req.EventID,
			Quantity:     req.Quantity,
			TotalPrice:   totalPrice,
			Status:       entity.TicketStatusActive,
			PurchaseDate: time.Now(),
		}

		// Create ticket record within transaction
		if err := tx.Create(ticket).Error; err != nil {
			return err
		}

		// Update event available tickets within transaction
		if err := tx.Model(&entity.Event{}).
			Where("id = ?", req.EventID).
			UpdateColumn("available", gorm.Expr("available - ?", req.Quantity)).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Return ticket with relations
	return s.GetTicketByID(ticket.ID)
}

func (s *ticketService) GetTicketByID(id string) (*entity.Ticket, error) {
	return s.ticketRepo.GetByID(id)
}

func (s *ticketService) GetUserTickets(userID string, pagination *entity.Pagination) ([]entity.Ticket, *entity.PaginationMeta, error) {
	tickets, total, err := s.ticketRepo.GetByUserID(userID, pagination)
	if err != nil {
		return nil, nil, err
	}

	meta := &entity.PaginationMeta{
		CurrentPage: pagination.Page,
		TotalItems:  total,
		Limit:       pagination.GetLimit(),
		TotalPages:  int((total + int64(pagination.GetLimit()) - 1) / int64(pagination.GetLimit())),
	}

	return tickets, meta, nil
}

func (s *ticketService) GetAllTickets(pagination *entity.Pagination, search *entity.Search, filter *entity.TicketFilter) ([]entity.Ticket, *entity.PaginationMeta, error) {
	tickets, total, err := s.ticketRepo.GetAll(pagination, search, filter)
	if err != nil {
		return nil, nil, err
	}

	meta := &entity.PaginationMeta{
		CurrentPage: pagination.Page,
		TotalItems:  total,
		Limit:       pagination.GetLimit(),
		TotalPages:  int((total + int64(pagination.GetLimit()) - 1) / int64(pagination.GetLimit())),
	}

	return tickets, meta, nil
}

func (s *ticketService) UpdateTicketStatus(ticketID string, req *entity.UpdateTicketStatusRequest) (*entity.Ticket, error) {
	ticket, err := s.ticketRepo.GetByID(ticketID)
	if err != nil {
		return nil, err
	}

	// Validate status transition
	if ticket.Status == entity.TicketStatusCancelled {
		return nil, errors.New("cannot update cancelled ticket")
	}

	if req.Status == entity.TicketStatusUsed && ticket.Status != entity.TicketStatusActive {
		return nil, errors.New("can only mark active tickets as used")
	}

	// Update status
	ticket.Status = req.Status
	if err := s.ticketRepo.Update(ticket); err != nil {
		return nil, err
	}

	return ticket, nil
}

func (s *ticketService) CancelTicket(ticketID, userID string) (*entity.Ticket, error) {
	var ticket *entity.Ticket
	var err error

	// Start transaction
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Get ticket with SELECT FOR UPDATE
		var ticketEntity entity.Ticket
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", ticketID).First(&ticketEntity).Error; err != nil {
			return err
		}
		ticket = &ticketEntity

		// Check ownership (users can only cancel their own tickets)
		if ticket.UserID != userID {
			return errors.New("you can only cancel your own tickets")
		}

		// Check if ticket can be cancelled
		if !ticket.CanBeCancelled() {
			return errors.New("ticket cannot be cancelled")
		}

		// Get event to check timing
		var event entity.Event
		if err := tx.Where("id = ?", ticket.EventID).First(&event).Error; err != nil {
			return err
		}

		if event.EventDate.Before(time.Now().Add(2 * time.Hour)) {
			return errors.New("cannot cancel tickets within 2 hours of event start")
		}

		// Update ticket status within transaction
		ticket.Status = entity.TicketStatusCancelled
		if err := tx.Save(ticket).Error; err != nil {
			return err
		}

		// Return tickets to event availability within transaction (negative quantity to add back)
		if err := tx.Model(&entity.Event{}).
			Where("id = ?", ticket.EventID).
			UpdateColumn("available", gorm.Expr("available + ?", ticket.Quantity)).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return ticket, nil
}

func (s *ticketService) GetTicketStats() (*entity.ReportSummary, error) {
	return s.ticketRepo.GetTicketStats()
}

func (s *ticketService) GetEventReport(eventID string) (*entity.EventReport, error) {
	// Validate event exists
	_, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		return nil, err
	}

	return s.ticketRepo.GetEventReport(eventID)
} 