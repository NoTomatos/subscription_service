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
	sub, err := req.ToSubscription()
	if err != nil {
		logrus.WithError(err).Error("Failed to convert request to subscription")
		return nil, fmt.Errorf("invalid request data: %w", err)
	}

	if err := s.repo.Create(sub); err != nil {
		return nil, err
	}

	return sub, nil
}

func (s *subscriptionService) GetByID(id string) (*model.Subscription, error) {
	uuidID, err := uuid.Parse(id)
	if err != nil {
		logrus.WithError(err).WithField("id", id).Error("Invalid UUID format")
		return nil, fmt.Errorf("invalid id format: %w", err)
	}

	sub, err := s.repo.GetByID(uuidID)
	if err != nil {
		return nil, err
	}

	return sub, nil
}

func (s *subscriptionService) Update(id string, req *model.UpdateSubscriptionRequest) error {
	uuidID, err := uuid.Parse(id)
	if err != nil {
		logrus.WithError(err).WithField("id", id).Error("Invalid UUID format")
		return fmt.Errorf("invalid id format: %w", err)
	}

	updates := make(map[string]interface{})

	if req.ServiceName != nil {
		updates["service_name"] = *req.ServiceName
	}

	if req.Price != nil {
		updates["price"] = *req.Price
	}

	if req.EndDate != nil {
		if *req.EndDate == "" {
			updates["end_date"] = nil
		} else {
			endDate, err := time.Parse("01-2006", *req.EndDate)
			if err != nil {
				logrus.WithError(err).Error("Invalid end date format")
				return fmt.Errorf("invalid end date format: %w", err)
			}
			updates["end_date"] = endDate
		}
	}

	if err := s.repo.Update(uuidID, updates); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("subscription not found")
		}
		return err
	}

	return nil
}

func (s *subscriptionService) Delete(id string) error {
	uuidID, err := uuid.Parse(id)
	if err != nil {
		logrus.WithError(err).WithField("id", id).Error("Invalid UUID format")
		return fmt.Errorf("invalid id format: %w", err)
	}

	if err := s.repo.Delete(uuidID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("subscription not found")
		}
		return err
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
			return nil, fmt.Errorf("invalid user_id format: %w", err)
		}
		filter.UserID = &uuidUserID
	}

	if serviceName != nil {
		filter.ServiceName = serviceName
	}

	if startDate != nil {
		sd, err := time.Parse("01-2006", *startDate)
		if err != nil {
			logrus.WithError(err).WithField("start_date", *startDate).Error("Invalid start_date format")
			return nil, fmt.Errorf("invalid start_date format: %w", err)
		}
		filter.StartDate = &sd
	}

	if endDate != nil {
		ed, err := time.Parse("01-2006", *endDate)
		if err != nil {
			logrus.WithError(err).WithField("end_date", *endDate).Error("Invalid end_date format")
			return nil, fmt.Errorf("invalid end_date format: %w", err)
		}
		filter.EndDate = &ed
	}

	subscriptions, err := s.repo.List(filter)
	if err != nil {
		return nil, err
	}

	return subscriptions, nil
}

func (s *subscriptionService) Aggregate(req *model.AggregateRequest) (*model.AggregateResponse, error) {
	startDate, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		logrus.WithError(err).WithField("start_date", req.StartDate).Error("Invalid start_date format")
		return nil, fmt.Errorf("invalid start_date format: %w", err)
	}

	endDate, err := time.Parse("01-2006", req.EndDate)
	if err != nil {
		logrus.WithError(err).WithField("end_date", req.EndDate).Error("Invalid end_date format")
		return nil, fmt.Errorf("invalid end_date format: %w", err)
	}

	if startDate.After(endDate) {
		return nil, fmt.Errorf("start_date must be before or equal to end_date")
	}

	var userIDPtr *uuid.UUID
	if req.UserID != nil {
		uuidUserID, err := uuid.Parse(*req.UserID)
		if err != nil {
			logrus.WithError(err).WithField("user_id", *req.UserID).Error("Invalid user_id format")
			return nil, fmt.Errorf("invalid user_id format: %w", err)
		}
		userIDPtr = &uuidUserID
	}

	total, err := s.repo.Aggregate(startDate, endDate, userIDPtr, req.ServiceName)
	if err != nil {
		return nil, err
	}

	return &model.AggregateResponse{TotalPrice: total}, nil
}
