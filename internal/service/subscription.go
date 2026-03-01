package service

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"subscription_service/internal/model"
	"subscription_service/internal/repository"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var (
	ErrNoUpdates = errors.New("no fields to update")
)

type ValidationError struct {
	Field string
	Err   error
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error on field '%s': %v", e.Field, e.Err)
	}
	return fmt.Sprintf("validation error: %v", e.Err)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

type NotFoundError struct {
	ID string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("subscription with id '%s' not found", e.ID)
}

type SubscriptionService interface {
	Create(req *model.CreateSubscriptionRequest) (*model.Subscription, error)
	GetByID(id string) (*model.Subscription, error)
	Update(id string, req *model.UpdateSubscriptionRequest) error
	Delete(id string) error
	List(userID, serviceName *string, startDate, endDate *string, limit, offset int) ([]*model.Subscription, error)
	Aggregate(req *model.AggregateRequest) (*model.AggregateResponse, error)
}

type subscriptionService struct {
	repo repository.SubscriptionRepository
}

func NewSubscriptionService(repo repository.SubscriptionRepository) SubscriptionService {
	return &subscriptionService{repo: repo}
}

func (s *subscriptionService) Create(req *model.CreateSubscriptionRequest) (*model.Subscription, error) {
	if req.Price < 0 {
		return nil, &ValidationError{
			Field: "price",
			Err:   errors.New("price cannot be negative"),
		}
	}

	sub, err := req.ToSubscription()
	if err != nil {
		logrus.WithError(err).Error("Failed to convert request to subscription")
		return nil, &ValidationError{
			Field: "request",
			Err:   err,
		}
	}

	if err := s.repo.Create(sub); err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	return sub, nil
}

func (s *subscriptionService) GetByID(id string) (*model.Subscription, error) {
	uuidID, err := uuid.Parse(id)
	if err != nil {
		logrus.WithError(err).WithField("id", id).Error("Invalid UUID format")
		return nil, &ValidationError{
			Field: "id",
			Err:   fmt.Errorf("invalid UUID format: %w", err),
		}
	}

	sub, err := s.repo.GetByID(uuidID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	if sub == nil {
		return nil, &NotFoundError{ID: id}
	}

	return sub, nil
}

func (s *subscriptionService) Update(id string, req *model.UpdateSubscriptionRequest) error {
	uuidID, err := uuid.Parse(id)
	if err != nil {
		logrus.WithError(err).WithField("id", id).Error("Invalid UUID format")
		return &ValidationError{
			Field: "id",
			Err:   fmt.Errorf("invalid UUID format: %w", err),
		}
	}

	updates := make(map[string]interface{})

	if req.ServiceName != nil {
		updates["service_name"] = *req.ServiceName
	}

	if req.Price != nil {
		if *req.Price < 0 {
			return &ValidationError{
				Field: "price",
				Err:   errors.New("price cannot be negative"),
			}
		}
		updates["price"] = *req.Price
	}

	if req.EndDate != nil {
		if *req.EndDate == "" {
			updates["end_date"] = nil
		} else {
			endDate, err := time.Parse("2006-01-02", *req.EndDate)
			if err != nil {
				logrus.WithError(err).Error("Invalid end date format")
				return &ValidationError{
					Field: "end_date",
					Err:   fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err),
				}
			}
			updates["end_date"] = endDate
		}
	}

	if len(updates) == 0 {
		return ErrNoUpdates
	}

	if err := s.repo.Update(uuidID, updates); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &NotFoundError{ID: id}
		}
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	return nil
}

func (s *subscriptionService) Delete(id string) error {
	uuidID, err := uuid.Parse(id)
	if err != nil {
		logrus.WithError(err).WithField("id", id).Error("Invalid UUID format")
		return &ValidationError{
			Field: "id",
			Err:   fmt.Errorf("invalid UUID format: %w", err),
		}
	}

	if err := s.repo.Delete(uuidID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &NotFoundError{ID: id}
		}
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	return nil
}

func (s *subscriptionService) List(userID, serviceName *string, startDate, endDate *string, limit, offset int) ([]*model.Subscription, error) {
	filter := model.SubscriptionFilter{
		Limit:  limit,
		Offset: offset,
	}

	if userID != nil {
		uuidUserID, err := uuid.Parse(*userID)
		if err != nil {
			logrus.WithError(err).WithField("user_id", *userID).Error("Invalid user_id format")
			return nil, &ValidationError{
				Field: "user_id",
				Err:   fmt.Errorf("invalid UUID format: %w", err),
			}
		}
		filter.UserID = &uuidUserID
	}

	if serviceName != nil {
		filter.ServiceName = serviceName
	}

	if startDate != nil {
		sd, err := time.Parse("2006-01-02", *startDate)
		if err != nil {
			logrus.WithError(err).WithField("start_date", *startDate).Error("Invalid start_date format")
			return nil, &ValidationError{
				Field: "start_date",
				Err:   fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err),
			}
		}
		filter.StartDate = &sd
	}

	if endDate != nil {
		ed, err := time.Parse("2006-01-02", *endDate)
		if err != nil {
			logrus.WithError(err).WithField("end_date", *endDate).Error("Invalid end_date format")
			return nil, &ValidationError{
				Field: "end_date",
				Err:   fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err),
			}
		}
		filter.EndDate = &ed
	}

	subscriptions, err := s.repo.List(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}

	return subscriptions, nil
}

func (s *subscriptionService) Aggregate(req *model.AggregateRequest) (*model.AggregateResponse, error) {
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		logrus.WithError(err).WithField("start_date", req.StartDate).Error("Invalid start_date format")
		return nil, &ValidationError{
			Field: "start_date",
			Err:   fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err),
		}
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		logrus.WithError(err).WithField("end_date", req.EndDate).Error("Invalid end_date format")
		return nil, &ValidationError{
			Field: "end_date",
			Err:   fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err),
		}
	}

	if startDate.After(endDate) {
		return nil, &ValidationError{
			Field: "date_range",
			Err:   errors.New("start_date must be before or equal to end_date"),
		}
	}

	var userIDPtr *uuid.UUID
	if req.UserID != nil {
		uuidUserID, err := uuid.Parse(*req.UserID)
		if err != nil {
			logrus.WithError(err).WithField("user_id", *req.UserID).Error("Invalid user_id format")
			return nil, &ValidationError{
				Field: "user_id",
				Err:   fmt.Errorf("invalid UUID format: %w", err),
			}
		}
		userIDPtr = &uuidUserID
	}

	total, err := s.repo.Aggregate(startDate, endDate, userIDPtr, req.ServiceName)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate subscriptions: %w", err)
	}

	return &model.AggregateResponse{TotalPrice: total}, nil
}
