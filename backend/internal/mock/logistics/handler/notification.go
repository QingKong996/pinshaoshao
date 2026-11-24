package handler

import (
	"context"
	"net/http"
	"strconv"

	"pinshaoshao/internal/mock/logistics/dto"
	"pinshaoshao/internal/mock/logistics/model"

	"github.com/gin-gonic/gin"
)

type NotificationService interface {
	ListByShipmentNo(ctx context.Context, shipmentNo string) ([]*model.Notification, error)
	Retry(ctx context.Context, shipmentNo string, notificationID int64) (*model.Notification, error)
}

type NotificationHandler struct {
	shipments     ShipmentService
	notifications NotificationService
}

func NewNotificationHandler(
	shipmentService ShipmentService,
	notificationService NotificationService,
) *NotificationHandler {
	return &NotificationHandler{
		shipments:     shipmentService,
		notifications: notificationService,
	}
}

func (h *NotificationHandler) List(ctx *gin.Context) {
	shipmentNo := ctx.Param("shipmentNo")
	if _, err := h.shipments.Get(ctx.Request.Context(), shipmentNo); err != nil {
		writeMockLogisticsError(ctx, "get mock shipment notifications", err)
		return
	}
	notifications, err := h.notifications.ListByShipmentNo(ctx.Request.Context(), shipmentNo)
	if err != nil {
		writeMockLogisticsError(ctx, "list mock shipment notifications", err)
		return
	}
	items := make([]dto.NotificationResponse, 0, len(notifications))
	for _, notification := range notifications {
		items = append(items, newNotificationResponse(notification))
	}
	ctx.JSON(http.StatusOK, dto.NotificationListResponse{Items: items})
}

func (h *NotificationHandler) Retry(ctx *gin.Context) {
	shipmentNo := ctx.Param("shipmentNo")
	if _, err := h.shipments.Get(ctx.Request.Context(), shipmentNo); err != nil {
		writeMockLogisticsError(ctx, "get mock shipment notification for retry", err)
		return
	}
	notificationID, err := strconv.ParseInt(ctx.Param("notificationID"), 10, 64)
	if err != nil || notificationID <= 0 {
		writeMockLogisticsError(ctx, "retry mock shipment notification", model.ErrInvalidNotificationID)
		return
	}

	notification, err := h.notifications.Retry(ctx.Request.Context(), shipmentNo, notificationID)
	if err != nil {
		writeMockLogisticsError(ctx, "retry mock shipment notification", err)
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

func (h *NotificationHandler) WebhookSink(ctx *gin.Context) {
	ctx.Status(http.StatusNoContent)
}
