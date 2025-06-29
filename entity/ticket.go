package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TicketStatus string

const (
	TicketStatusActive    TicketStatus = "active"
	TicketStatusUsed      TicketStatus = "used"
	TicketStatusCancelled TicketStatus = "cancelled"
	TicketStatusExpired   TicketStatus = "expired"
)

type Ticket struct {
	ID           string         `json:"id" gorm:"type:varchar(36);primary_key"`
	UserID       string         `json:"user_id" gorm:"type:varchar(36);not null"`
	EventID      string         `json:"event_id" gorm:"type:varchar(36);not null"`
	Quantity     int            `json:"quantity" gorm:"not null;default:1" validate:"required,min=1"`
	TotalPrice   float64        `json:"total_price" gorm:"not null"`
	Status       TicketStatus   `json:"status" gorm:"type:enum('active','used','cancelled','expired');default:'active'"`
	PurchaseDate time.Time      `json:"purchase_date" gorm:"not null"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relationships
	User  User  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Event Event `json:"event,omitempty" gorm:"foreignKey:EventID"`
}

func (t *Ticket) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	if t.PurchaseDate.IsZero() {
		t.PurchaseDate = time.Now()
	}
	return nil
}

func (t *Ticket) CanBeCancelled() bool {
	return t.Status == TicketStatusActive
}

type BuyTicketRequest struct {
	EventID  string `json:"event_id" validate:"required"`
	Quantity int    `json:"quantity" validate:"required,min=1"`
}

type TicketFilter struct {
	UserID    string `form:"user_id"`
	EventID   string `form:"event_id"`
	Status    string `form:"status"`
	StartDate *time.Time `form:"start_date"`
	EndDate   *time.Time `form:"end_date"`
}

type UpdateTicketStatusRequest struct {
	Status TicketStatus `json:"status" validate:"required,oneof=cancelled used"`
} 