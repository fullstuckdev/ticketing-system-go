package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventStatus string

const (
	EventStatusActive    EventStatus = "active"
	EventStatusOngoing   EventStatus = "ongoing"
	EventStatusCompleted EventStatus = "completed"
	EventStatusCancelled EventStatus = "cancelled"
)

type Event struct {
	ID          string         `json:"id" gorm:"type:varchar(36);primary_key"`
	Name        string         `json:"name" gorm:"uniqueIndex;not null" validate:"required,min=3"`
	Description string         `json:"description" gorm:"type:text"`
	Category    string         `json:"category" gorm:"not null" validate:"required"`
	Capacity    int            `json:"capacity" gorm:"not null" validate:"required,min=1"`
	Available   int            `json:"available" gorm:"not null"`
	Price       float64        `json:"price" gorm:"not null" validate:"required,min=0"`
	Location    string         `json:"location" gorm:"not null" validate:"required"`
	EventDate   time.Time      `json:"event_date" gorm:"not null" validate:"required"`
	Status      EventStatus    `json:"status" gorm:"type:enum('active','ongoing','completed','cancelled');default:'active'"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relationships
	Tickets []Ticket `json:"tickets,omitempty" gorm:"foreignKey:EventID"`
}

func (e *Event) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	if e.Available == 0 {
		e.Available = e.Capacity
	}
	return nil
}

func (e *Event) IsAvailable() bool {
	return e.Available > 0 && e.Status == EventStatusActive
}

func (e *Event) CanBeModified() bool {
	return e.Status == EventStatusActive
}

type CreateEventRequest struct {
	Name        string    `json:"name" validate:"required,min=3"`
	Description string    `json:"description"`
	Category    string    `json:"category" validate:"required"`
	Capacity    int       `json:"capacity" validate:"required,min=1"`
	Price       float64   `json:"price" validate:"required,min=0"`
	Location    string    `json:"location" validate:"required"`
	EventDate   time.Time `json:"event_date" validate:"required"`
}

type UpdateEventRequest struct {
	Name        *string    `json:"name,omitempty" validate:"omitempty,min=3"`
	Description *string    `json:"description,omitempty"`
	Category    *string    `json:"category,omitempty"`
	Capacity    *int       `json:"capacity,omitempty" validate:"omitempty,min=1"`
	Price       *float64   `json:"price,omitempty" validate:"omitempty,min=0"`
	Location    *string    `json:"location,omitempty"`
	EventDate   *time.Time `json:"event_date,omitempty"`
}

type EventFilter struct {
	Category  string `form:"category"`
	Status    string `form:"status"`
	Location  string `form:"location"`
	MinPrice  *float64 `form:"min_price"`
	MaxPrice  *float64 `form:"max_price"`
	StartDate *time.Time `form:"start_date"`
	EndDate   *time.Time `form:"end_date"`
} 