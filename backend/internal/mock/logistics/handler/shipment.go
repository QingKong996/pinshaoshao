package handler

import (
	"context"
	"net/http"

	"pinshaoshao/internal/http/httpx"
	"pinshaoshao/internal/mock/logistics/dto"
	"pinshaoshao/internal/mock/logistics/model"
	"pinshaoshao/internal/mock/logistics/service"

	"github.com/gin-gonic/gin"
)

type ShipmentService interface {
	Create(ctx context.Context, input service.CreateShipmentInput) (*model.Shipment, bool, error)
	Get(ctx context.Context, shipmentNo string) (*model.Shipment, error)
	Outbound(ctx context.Context, shipmentNo string) (*model.Shipment, error)
	Advance(ctx context.Context, shipmentNo string) (*model.Shipment, error)
	Complete(ctx context.Context, shipmentNo string) (*model.Shipment, error)
}

type ShipmentHandler struct {
	service ShipmentService
}

func NewShipmentHandler(shipmentService ShipmentService) *ShipmentHandler {
	return &ShipmentHandler{service: shipmentService}
}

func (h *ShipmentHandler) Create(ctx *gin.Context) {
	var request dto.CreateShipmentRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: httpx.InvalidJSONMessage})
		return
	}

	shipment, created, err := h.service.Create(ctx.Request.Context(), service.CreateShipmentInput{
		MerchantOrderNo: request.MerchantOrderNo,
		ItemSummary:     request.ItemSummary,
		Carrier:         request.Carrier,
		SenderName:      request.SenderName,
		SenderPhone:     request.SenderPhone,
		SenderAddress:   request.SenderAddress,
		ReceiverName:    request.ReceiverName,
		ReceiverPhone:   request.ReceiverPhone,
		ReceiverAddress: request.ReceiverAddress,
		OriginCity:      request.OriginCity,
		DestinationCity: request.DestinationCity,
		NotifyURL:       request.NotifyURL,
	})
	if err != nil {
		writeMockLogisticsError(ctx, "create mock shipment", err)
		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	ctx.JSON(status, newShipmentResponse(shipment))
}

func (h *ShipmentHandler) Get(ctx *gin.Context) {
	shipment, err := h.service.Get(ctx.Request.Context(), ctx.Param("shipmentNo"))
	if err != nil {
		writeMockLogisticsError(ctx, "get mock shipment", err)
		return
	}
	ctx.JSON(http.StatusOK, newShipmentResponse(shipment))
}

func (h *ShipmentHandler) Outbound(ctx *gin.Context) {
	shipment, err := h.service.Outbound(ctx.Request.Context(), ctx.Param("shipmentNo"))
	if err != nil {
		writeMockLogisticsError(ctx, "dispatch mock shipment", err)
		return
	}
	ctx.JSON(http.StatusOK, newShipmentResponse(shipment))
}

func (h *ShipmentHandler) Advance(ctx *gin.Context) {
	shipment, err := h.service.Advance(ctx.Request.Context(), ctx.Param("shipmentNo"))
	if err != nil {
		writeMockLogisticsError(ctx, "advance mock shipment", err)
		return
	}
	ctx.JSON(http.StatusOK, newShipmentResponse(shipment))
}

func (h *ShipmentHandler) Complete(ctx *gin.Context) {
	shipment, err := h.service.Complete(ctx.Request.Context(), ctx.Param("shipmentNo"))
	if err != nil {
		writeMockLogisticsError(ctx, "complete mock shipment", err)
		return
	}
	ctx.JSON(http.StatusOK, newShipmentResponse(shipment))
}

func newShipmentResponse(shipment *model.Shipment) dto.ShipmentResponse {
	events := make([]dto.ShipmentEventResponse, 0, len(shipment.Events))
	for _, event := range shipment.Events {
		events = append(events, dto.ShipmentEventResponse{
			Sequence:    event.Sequence,
			Status:      string(event.Status),
			EventType:   event.EventType,
			Description: event.Description,
			Location:    event.Location,
			OccurredAt:  event.OccurredAt,
		})
	}
	return dto.ShipmentResponse{
		ShipmentNo:      shipment.ShipmentNo,
		TrackingNo:      shipment.TrackingNo,
		MerchantOrderNo: shipment.MerchantOrderNo,
		Status:          string(shipment.Status),
		Carrier:         shipment.Carrier,
		ItemSummary:     shipment.ItemSummary,
		Sender: dto.PartyResponse{
			Name:    shipment.Sender.Name,
			Phone:   shipment.Sender.Phone,
			Address: shipment.Sender.Address,
		},
		Receiver: dto.PartyResponse{
			Name:    shipment.Receiver.Name,
			Phone:   shipment.Receiver.Phone,
			Address: shipment.Receiver.Address,
		},
		OriginCity:      shipment.OriginCity,
		DestinationCity: shipment.DestinationCity,
		NotifyURL:       shipment.NotifyURL,
		CurrentLocation: shipment.CurrentLocation,
		Events:          events,
		DispatchedAt:    shipment.DispatchedAt,
		PickedUpAt:      shipment.PickedUpAt,
		DeliveredAt:     shipment.DeliveredAt,
		CreatedAt:       shipment.CreatedAt,
		UpdatedAt:       shipment.UpdatedAt,
	}
}
