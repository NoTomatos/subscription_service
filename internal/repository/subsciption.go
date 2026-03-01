package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"subscription_service/internal/model"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type SubscriptionRepository interface {
	Create(sub *model.Subscription) error
	GetByID(id uuid.UUID) (*model.Subscription, error)
	Update(id uuid.UUID, updates map[string]interface{}) error
	Delete(id uuid.UUID) error
	List(filter model.SubscriptionFilter) ([]*model.Subscription, error)
	Aggregate(startDate, endDate time.Time, userID *uuid.UUID, serviceName *string) (int, error)
}

type subscriptionRepository struct {
	db *sql.DB
}

func NewSubscriptionRepository(db *sql.DB) SubscriptionRepository {
	return &subscriptionRepository{db: db}
}

func (r *subscriptionRepository) Create(sub *model.Subscription) error {
	query := `
        INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `

	now := time.Now()
	sub.CreatedAt = now
	sub.UpdatedAt = now

	_, err := r.db.Exec(query,
		sub.ID, sub.ServiceName, sub.Price, sub.UserID,
		sub.StartDate, sub.EndDate, sub.CreatedAt, sub.UpdatedAt,
	)

	if err != nil {
		logrus.WithError(err).Error("Failed to create subscription")
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"id":           sub.ID,
		"service_name": sub.ServiceName,
		"user_id":      sub.UserID,
	}).Info("Subscription created successfully")

	return nil
}

func (r *subscriptionRepository) GetByID(id uuid.UUID) (*model.Subscription, error) {
	query := `
        SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
        FROM subscriptions
        WHERE id = $1
    `

	var sub model.Subscription
	err := r.db.QueryRow(query, id).Scan(
		&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID,
		&sub.StartDate, &sub.EndDate, &sub.CreatedAt, &sub.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		logrus.WithError(err).WithField("id", id).Error("Failed to get subscription")
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return &sub, nil
}

func (r *subscriptionRepository) Update(id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	setClauses := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	i := 1

	for field, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, i))
		args = append(args, value)
		i++
	}

	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", i))
	args = append(args, time.Now())
	i++

	args = append(args, id)

	query := fmt.Sprintf(`
        UPDATE subscriptions
        SET %s
        WHERE id = $%d
    `, strings.Join(setClauses, ", "), i)

	result, err := r.db.Exec(query, args...)
	if err != nil {
		logrus.WithError(err).WithField("id", id).Error("Failed to update subscription")
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	logrus.WithFields(logrus.Fields{
		"id":     id,
		"fields": updates,
	}).Info("Subscription updated successfully")

	return nil
}

func (r *subscriptionRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM subscriptions WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		logrus.WithError(err).WithField("id", id).Error("Failed to delete subscription")
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	logrus.WithField("id", id).Info("Subscription deleted successfully")
	return nil
}

func (r *subscriptionRepository) List(filter model.SubscriptionFilter) ([]*model.Subscription, error) {
	query := `
        SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
        FROM subscriptions
        WHERE 1=1
    `
	args := make([]interface{}, 0)
	i := 1

	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", i)
		args = append(args, *filter.UserID)
		i++
	}

	if filter.ServiceName != nil {
		query += fmt.Sprintf(" AND service_name ILIKE $%d", i)
		args = append(args, "%"+*filter.ServiceName+"%")
		i++
	}

	if filter.StartDate != nil {
		query += fmt.Sprintf(" AND start_date >= $%d", i)
		args = append(args, *filter.StartDate)
		i++
	}

	if filter.EndDate != nil {
		query += fmt.Sprintf(" AND (end_date IS NULL OR end_date <= $%d)", i)
		args = append(args, *filter.EndDate)
		i++
	}

	query += " ORDER BY start_date DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", i)
		args = append(args, filter.Limit)
		i++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", i)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		logrus.WithError(err).Error("Failed to list subscriptions")
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []*model.Subscription
	for rows.Next() {
		var sub model.Subscription
		err := rows.Scan(
			&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID,
			&sub.StartDate, &sub.EndDate, &sub.CreatedAt, &sub.UpdatedAt,
		)
		if err != nil {
			logrus.WithError(err).Error("Failed to scan subscription")
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}
		subscriptions = append(subscriptions, &sub)
	}

	return subscriptions, nil
}

func (r *subscriptionRepository) Aggregate(startDate, endDate time.Time, userID *uuid.UUID, serviceName *string) (int, error) {
	query := `
        SELECT COALESCE(SUM(price), 0)
        FROM subscriptions
        WHERE start_date <= $2
        AND (end_date IS NULL OR end_date >= $1)
    `
	args := []interface{}{startDate, endDate}
	i := 3

	if userID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", i)
		args = append(args, *userID)
		i++
	}

	if serviceName != nil {
		query += fmt.Sprintf(" AND service_name ILIKE $%d", i)
		args = append(args, *serviceName)
	}

	var total int
	err := r.db.QueryRow(query, args...).Scan(&total)
	if err != nil {
		logrus.WithError(err).Error("Failed to aggregate subscriptions")
		return 0, fmt.Errorf("failed to aggregate subscriptions: %w", err)
	}

	return total, nil
}
