package service

import (
	"errors"
	"ticketing-system/entity"
	"ticketing-system/repository"
	"time"
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
}

func NewTicketService(
	ticketRepo repository.TicketRepository,
	eventRepo repository.EventRepository,
	userRepo repository.UserRepository,
) TicketService {
	return &ticketService{
		ticketRepo: ticketRepo,
		eventRepo:  eventRepo,
		userRepo:   userRepo,
	}
}

func (s *ticketService) BuyTicket(userID string, req *entity.BuyTicketRequest) (*entity.Ticket, error) {
	// Validate user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if !user.IsActive {
		return nil, errors.New("user account is not active")
	}

	// Validate event
	event, err := s.eventRepo.GetByID(req.EventID)
	if err != nil {
		return nil, err
	}

	// Check event availability
	if !event.IsAvailable() {
		return nil, errors.New("event is not available for booking")
	}

	// Check capacity
	if event.Available < req.Quantity {
		return nil, errors.New("insufficient tickets available")
	}

	// Check if event date is in the future
	if event.EventDate.Before(time.Now().Add(time.Hour)) {
		return nil, errors.New("cannot purchase tickets for events starting within an hour")
	}

	// Calculate total price
	totalPrice := event.Price * float64(req.Quantity)

	// Create ticket
	ticket := &entity.Ticket{
		UserID:       userID,
		EventID:      req.EventID,
		Quantity:     req.Quantity,
		TotalPrice:   totalPrice,
		Status:       entity.TicketStatusActive,
		PurchaseDate: time.Now(),
	}

	// Create ticket record
	if err := s.ticketRepo.Create(ticket); err != nil {
		return nil, err
	}

	// Update event available tickets
	if err := s.eventRepo.UpdateAvailableTickets(req.EventID, req.Quantity); err != nil {
		// If updating availability fails, we should delete the ticket
		s.ticketRepo.Delete(ticket.ID)
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
	ticket, err := s.ticketRepo.GetByID(ticketID)
	if err != nil {
		return nil, err
	}

	// Check ownership (users can only cancel their own tickets)
	if ticket.UserID != userID {
		return nil, errors.New("you can only cancel your own tickets")
	}

	// Check if ticket can be cancelled
	if !ticket.CanBeCancelled() {
		return nil, errors.New("ticket cannot be cancelled")
	}

	// Check if event is still in the future (allow cancellation up to 2 hours before event)
	event, err := s.eventRepo.GetByID(ticket.EventID)
	if err != nil {
		return nil, err
	}

	if event.EventDate.Before(time.Now().Add(2 * time.Hour)) {
		return nil, errors.New("cannot cancel tickets within 2 hours of event start")
	}

	// Update ticket status
	ticket.Status = entity.TicketStatusCancelled
	if err := s.ticketRepo.Update(ticket); err != nil {
		return nil, err
	}

	// Return tickets to event availability (negative quantity to add back)
	if err := s.eventRepo.UpdateAvailableTickets(ticket.EventID, -ticket.Quantity); err != nil {
		// If updating availability fails, revert ticket status
		ticket.Status = entity.TicketStatusActive
		s.ticketRepo.Update(ticket)
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