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

// CreateSubscription
// @Summary Создать новую подписку
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body model.CreateSubscriptionRequest true "Данные подписки"
// @Success 201 {object} model.Subscription
// @Failure 400 {object} map[string]interface{} "Неверный формат запроса"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/v1/subscriptions [post]
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

// GetSubscription
// @Summary Получить подписку по ID
// @Tags subscriptions
// @Produce json
// @Param id path string true "UUID подписки"
// @Success 200 {object} model.Subscription
// @Failure 400 {object} map[string]interface{} "Неверный формат ID"
// @Failure 404 {object} map[string]interface{} "Подписка не найдена"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/v1/subscriptions/{id} [get]
func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	id := c.Param("id")

	sub, err := h.service.GetByID(id)
	if err != nil {
		logrus.WithError(err).WithField("id", id).Error("Failed to get subscription")

		var validationErr *service.ValidationError
		if errors.As(err, &validationErr) {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		var notFoundErr *service.NotFoundError
		if errors.As(err, &notFoundErr) {
			c.JSON(http.StatusNotFound, gin.H{"error": notFoundErr.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscription"})
		return
	}

	c.JSON(http.StatusOK, sub)
}

// UpdateSubscription
// @Summary Обновить подписку
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "UUID подписки"
// @Param subscription body model.UpdateSubscriptionRequest true "Данные для обновления"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{} "Неверный формат запроса"
// @Failure 404 {object} map[string]interface{} "Подписка не найдена"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/v1/subscriptions/{id} [put]
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

			var notFoundErr *service.NotFoundError
			if errors.As(err, &notFoundErr) {
				c.JSON(http.StatusNotFound, gin.H{"error": notFoundErr.Error()})
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

// DeleteSubscription
// @Summary Удалить подписку
// @Tags subscriptions
// @Produce json
// @Param id path string true "UUID подписки"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{} "Неверный формат ID"
// @Failure 404 {object} map[string]interface{} "Подписка не найдена"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/v1/subscriptions/{id} [delete]
func (h *SubscriptionHandler) DeleteSubscription(c *gin.Context) {
	id := c.Param("id")

	err := h.service.Delete(id)
	if err != nil {
		logrus.WithError(err).WithField("id", id).Error("Failed to delete subscription")

		var validationErr *service.ValidationError
		if errors.As(err, &validationErr) {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		var notFoundErr *service.NotFoundError
		if errors.As(err, &notFoundErr) {
			c.JSON(http.StatusNotFound, gin.H{"error": notFoundErr.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Subscription deleted successfully",
		"id":      id,
	})
}

// ListSubscriptions
// @Summary Список подписок с фильтрацией
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "Фильтр по ID пользователя"
// @Param service_name query string false "Фильтр по названию сервиса"
// @Param start_date query string false "Фильтр по дате начала (подписки, начавшиеся не раньше)"
// @Param end_date query string false "Фильтр по дате начала (подписки, начавшиеся не позже)"
// @Param limit query int false "Лимит записей (по умолчанию 10)"
// @Param offset query int false "Смещение (по умолчанию 0)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{} "Неверные параметры запроса"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/v1/subscriptions [get]
func (h *SubscriptionHandler) ListSubscriptions(c *gin.Context) {
	userID := c.Query("user_id")
	serviceName := c.Query("service_name")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

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

// AggregateSubscriptions
// @Summary Подсчет суммарной стоимости подписок за период
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "Фильтр по ID пользователя"
// @Param service_name query string false "Фильтр по названию сервиса"
// @Param start_date query string true "Начало периода (YYYY-MM-DD)"
// @Param end_date query string true "Конец периода (YYYY-MM-DD)"
// @Success 200 {object} model.AggregateResponse
// @Failure 400 {object} map[string]interface{} "Неверные параметры запроса"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/v1/subscriptions/aggregate [get]
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
