package handler

import (
	"context"
	"net/http"

	"pinshaoshao/internal/mock/payment/dto"
	"pinshaoshao/internal/mock/payment/model"

	"github.com/gin-gonic/gin"
)

type NotificationLookupService interface {
	ListByPaymentNo(ctx context.Context, paymentNo string) ([]*model.Notification, error)
	Retry(ctx context.Context, paymentNo string) (*model.Notification, error)
}

type WebhookHandler struct {
	payments      PaymentService
	notifications NotificationLookupService
}

func NewWebhookHandler(
	paymentService PaymentService,
	notificationService NotificationLookupService,
) *WebhookHandler {
	return &WebhookHandler{
		payments:      paymentService,
		notifications: notificationService,
	}
}

func (h *WebhookHandler) List(ctx *gin.Context) {
	paymentNo := ctx.Param("paymentNo")
	if _, err := h.payments.Get(ctx.Request.Context(), paymentNo); err != nil {
		writeMockPaymentError(ctx, "get mock payment notifications", err)
		return
	}

	notifications, err := h.notifications.ListByPaymentNo(ctx.Request.Context(), paymentNo)
	if err != nil {
		writeMockPaymentError(ctx, "list mock payment notifications", err)
		return
	}
	items := make([]dto.NotificationResponse, 0, len(notifications))
	for _, notification := range notifications {
		items = append(items, newNotificationResponse(notification))
	}
	ctx.JSON(http.StatusOK, dto.NotificationListResponse{Items: items})
}

func (h *WebhookHandler) Retry(ctx *gin.Context) {
	paymentNo := ctx.Param("paymentNo")
	if _, err := h.payments.Get(ctx.Request.Context(), paymentNo); err != nil {
		writeMockPaymentError(ctx, "get mock payment notification for retry", err)
		return
	}

	notification, err := h.notifications.Retry(ctx.Request.Context(), paymentNo)
	if err != nil {
		writeMockPaymentError(ctx, "retry mock payment notification", err)
		return
	}
	ctx.JSON(http.StatusOK, newNotificationResponse(notification))
}

func newNotificationResponse(notification *model.Notification) dto.NotificationResponse {
	return dto.NotificationResponse{
		ID:            notification.ID,
		EventID:       notification.EventID,
		EventType:     notification.EventType,
		TargetURL:     notification.TargetURL,
		Status:        string(notification.Status),
		AttemptCount:  notification.AttemptCount,
		MaxAttempts:   notification.MaxAttempts,
		NextAttemptAt: notification.NextAttemptAt,
		LastAttemptAt: notification.LastAttemptAt,
		DeliveredAt:   notification.DeliveredAt,
		LastError:     notification.LastError,
		CreatedAt:     notification.CreatedAt,
		UpdatedAt:     notification.UpdatedAt,
	}
}
