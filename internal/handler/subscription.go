package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"subscription_service/internal/model"
	"subscription_service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type SubscriptionHandler struct {
	service service.SubscriptionService
}

func NewSubscriptionHandler(service service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{service: service}
}

func (h *SubscriptionHandler) CreateSubscription(c *gin.Context) {
	var req model.CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Warn("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	sub, err := h.service.Create(&req)
	if err != nil {
		logrus.WithError(err).Error("Failed to create subscription")

		// Проверяем тип ошибки
		var validationErr *service.ValidationError
		if errors.As(err, &validationErr) {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subscription"})
		return
	}

	c.JSON(http.StatusCreated, sub)
}

func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	id := c.Param("id")

	sub, err := h.service.GetByID(id)
	if err != nil {
		logrus.WithError(err).WithField("id", id).Error("Failed to get subscription")

		// Проверяем, является ли ошибка ошибкой валидации UUID
		var validationErr *service.ValidationError
		if errors.As(err, &validationErr) {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscription"})
		return
	}

	if sub == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	c.JSON(http.StatusOK, sub)
}

func (h *SubscriptionHandler) UpdateSubscription(c *gin.Context) {
	id := c.Param("id")

	var req model.UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Warn("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	err := h.service.Update(id, &req)
	if err != nil {
		logrus.WithError(err).WithField("id", id).Error("Failed to update subscription")

		// Проверяем различные типы ошибок
		switch {
		case errors.Is(err, sql.ErrNoRows):
			c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
			return

		case errors.Is(err, service.ErrNoUpdates):
			c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
			return

		default:
			var validationErr *service.ValidationError
			if errors.As(err, &validationErr) {
				c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Subscription updated successfully",
		"id":      id,
	})
}

func (h *SubscriptionHandler) DeleteSubscription(c *gin.Context) {
	id := c.Param("id")

	err := h.service.Delete(id)
	if err != nil {
		logrus.WithError(err).WithField("id", id).Error("Failed to delete subscription")

		// Проверяем различные типы ошибок
		switch {
		case errors.Is(err, sql.ErrNoRows):
			c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
			return

		default:
			var validationErr *service.ValidationError
			if errors.As(err, &validationErr) {
				c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete subscription"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Subscription deleted successfully",
		"id":      id,
	})
}

func (h *SubscriptionHandler) ListSubscriptions(c *gin.Context) {
	userID := c.Query("user_id")
	serviceName := c.Query("service_name")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// Парсим limit с проверкой
	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		} else if err != nil {
			logrus.WithField("limit", l).Warn("Invalid limit parameter")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
			return
		}
	}

	// Парсим offset с проверкой
	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		} else if err != nil {
			logrus.WithField("offset", o).Warn("Invalid offset parameter")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset parameter"})
			return
		}
	}

	// Преобразуем пустые строки в nil
	var userIDPtr, serviceNamePtr, startDatePtr, endDatePtr *string
	if userID != "" {
		userIDPtr = &userID
	}
	if serviceName != "" {
		serviceNamePtr = &serviceName
	}
	if startDate != "" {
		startDatePtr = &startDate
	}
	if endDate != "" {
		endDatePtr = &endDate
	}

	subscriptions, err := h.service.List(userIDPtr, serviceNamePtr, startDatePtr, endDatePtr, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to list subscriptions")

		var validationErr *service.ValidationError
		if errors.As(err, &validationErr) {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list subscriptions"})
		return
	}

	// Возвращаем пустой массив вместо null, если нет результатов
	if subscriptions == nil {
		subscriptions = []*model.Subscription{}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   subscriptions,
		"limit":  limit,
		"offset": offset,
		"total":  len(subscriptions),
	})
}

func (h *SubscriptionHandler) AggregateSubscriptions(c *gin.Context) {
	var req model.AggregateRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logrus.WithError(err).Warn("Invalid query parameters")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters: " + err.Error()})
		return
	}

	result, err := h.service.Aggregate(&req)
	if err != nil {
		logrus.WithError(err).Error("Failed to aggregate subscriptions")

		var validationErr *service.ValidationError
		if errors.As(err, &validationErr) {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to aggregate subscriptions"})
		return
	}

	c.JSON(http.StatusOK, result)
}
