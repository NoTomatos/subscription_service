package model

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	ServiceName string     `json:"service_name" db:"service_name" binding:"required"`
	Price       int        `json:"price" db:"price" binding:"required,min=0"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id" binding:"required"`
	StartDate   time.Time  `json:"start_date" db:"start_date" binding:"required"`
	EndDate     *time.Time `json:"end_date,omitempty" db:"end_date"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateSubscriptionRequest struct {
	ServiceName string `json:"service_name" binding:"required"`
	Price       int    `json:"price" binding:"required,min=0"`
	UserID      string `json:"user_id" binding:"required,uuid"`
	StartDate   string `json:"start_date" binding:"required,datetime=01-2006"`
	EndDate     string `json:"end_date,omitempty" binding:"omitempty,datetime=01-2006"`
}

type UpdateSubscriptionRequest struct {
	ServiceName *string `json:"service_name,omitempty"`
	Price       *int    `json:"price,omitempty" binding:"omitempty,min=0"`
	EndDate     *string `json:"end_date,omitempty" binding:"omitempty,datetime=01-2006"`
}

type SubscriptionFilter struct {
	UserID      *uuid.UUID
	ServiceName *string
	StartDate   *time.Time
	EndDate     *time.Time
	Limit       int
	Offset      int
}

type AggregateRequest struct {
	UserID      *string `form:"user_id" binding:"omitempty,uuid"`
	ServiceName *string `form:"service_name"`
	StartDate   string  `form:"start_date" binding:"required,datetime=01-2006"`
	EndDate     string  `form:"end_date" binding:"required,datetime=01-2006"`
}

type AggregateResponse struct {
	TotalPrice int `json:"total_price"`
}

func (r *CreateSubscriptionRequest) ToSubscription() (*Subscription, error) {
	userID, err := uuid.Parse(r.UserID)
	if err != nil {
		return nil, err
	}

	startDate, err := time.Parse("01-2006", r.StartDate)
	if err != nil {
		return nil, err
	}

	sub := &Subscription{
		ID:          uuid.New(),
		ServiceName: r.ServiceName,
		Price:       r.Price,
		UserID:      userID,
		StartDate:   startDate,
	}

	if r.EndDate != "" {
		endDate, err := time.Parse("01-2006", r.EndDate)
		if err != nil {
			return nil, err
		}
		sub.EndDate = &endDate
	}

	return sub, nil
}
