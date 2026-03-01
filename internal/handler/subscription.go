package handler

import (
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sub, err := h.service.Create(&req)
	if err != nil {
		logrus.WithError(err).Error("Failed to create subscription")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, sub)
}

func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	id := c.Param("id")

	sub, err := h.service.GetByID(id)
	if err != nil {
		logrus.WithError(err).WithField("id", id).Error("Failed to get subscription")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if sub == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	c.JSON(http.StatusOK, sub)
}

func (h *SubscriptionHandler) UpdateSubscription(c *gin.Context) {
	id := c.Param("id")

	var req model.UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Warn("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.Update(id, &req)
	if err != nil {
		if err.Error() == "subscription not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		logrus.WithError(err).WithField("id", id).Error("Failed to update subscription")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscription updated successfully"})
}

func (h *SubscriptionHandler) DeleteSubscription(c *gin.Context) {
	id := c.Param("id")

	err := h.service.Delete(id)
	if err != nil {
		if err.Error() == "subscription not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		logrus.WithError(err).WithField("id", id).Error("Failed to delete subscription")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscription deleted successfully"})
}

func (h *SubscriptionHandler) ListSubscriptions(c *gin.Context) {
	userID := c.Query("user_id")
	serviceName := c.Query("service_name")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, subscriptions)
}

func (h *SubscriptionHandler) AggregateSubscriptions(c *gin.Context) {
	var req model.AggregateRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logrus.WithError(err).Warn("Invalid query parameters")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.Aggregate(&req)
	if err != nil {
		logrus.WithError(err).Error("Failed to aggregate subscriptions")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
