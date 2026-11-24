package controller

import (
	"context"
	"errors"
	"net/http"

	"pinshaoshao/internal/feature/logistics/apperr"
	logisticsdto "pinshaoshao/internal/feature/logistics/dto"
	"pinshaoshao/internal/feature/logistics/model"
	"pinshaoshao/internal/http/authctx"
	"pinshaoshao/internal/http/httpx"

	"github.com/gin-gonic/gin"
)

type ShipmentService interface {
	Ship(ctx context.Context, shopID int, subOrderID int, carrier string) (*model.Shipment, bool, error)
	GetForMerchant(ctx context.Context, shopID int, shipmentNo string) (*model.Shipment, error)
	ListForMerchant(ctx context.Context, shopID int, filter model.ListFilter) (*model.ListResult, error)
	GetForBuyer(ctx context.Context, userID int, shipmentNo string) (*model.Shipment, error)
	ListForBuyerOrder(ctx context.Context, userID int, orderID int) ([]model.Summary, error)
}

type ShipmentController struct {
	service ShipmentService
}

func NewShipmentController(service ShipmentService) *ShipmentController {
	return &ShipmentController{service: service}
}

// Ship creates and dispatches the single shipment for a paid sub-order.
//
// @Summary     Ship a merchant sub-order
// @Description Create or return the single shipment for a paid sub-order and mark it shipped.
// @Tags        Merchant Logistics
// @Accept      json
// @Produce     json
// @Param       request body logisticsdto.CreateShipmentRequest true "Shipment request"
// @Success     201 {object} logisticsdto.ShipmentResponse
// @Success     200 {object} logisticsdto.ShipmentResponse
// @Failure     400 {object} httpx.ErrorResponse
// @Failure     401 {object} httpx.ErrorResponse
// @Failure     404 {object} httpx.ErrorResponse
// @Failure     409 {object} httpx.ErrorResponse
// @Failure     502 {object} httpx.ErrorResponse
// @Failure     500 {object} httpx.ErrorResponse
// @Security    BearerAuth
// @Router      /merchant/shipments [post]
func (c *ShipmentController) Ship(ctx *gin.Context) {
	var request logisticsdto.CreateShipmentRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: apperr.ErrInvalidShipmentInput.Error()})
		return
	}

	shipment, created, err := c.service.Ship(
		ctx.Request.Context(),
		authctx.MustShopID(ctx),
		request.SubOrderID,
		request.Carrier,
	)
	if err != nil {
		writeLogisticsError(ctx, "ship sub-order", err)
		return
	}
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	ctx.JSON(status, newShipmentResponse(shipment))
}

// ListForMerchant returns shipments owned by the authenticated shop.
//
// @Summary     List merchant shipments
// @Description List shipments for the authenticated shop newest first.
// @Tags        Merchant Logistics
// @Produce     json
// @Param       page query int false "Page number" default(1)
// @Param       page_size query int false "Page size" default(10)
// @Param       status query string false "Shipment status" Enums(pending_pickup,picked_up,in_transit,at_transit_hub,out_for_delivery,delivered)
// @Success     200 {object} logisticsdto.ShipmentListResponse
// @Failure     400 {object} httpx.ErrorResponse
// @Failure     401 {object} httpx.ErrorResponse
// @Failure     500 {object} httpx.ErrorResponse
// @Security    BearerAuth
// @Router      /merchant/shipments [get]
func (c *ShipmentController) ListForMerchant(ctx *gin.Context) {
	var request logisticsdto.ListShipmentsQuery
	if err := ctx.ShouldBindQuery(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: apperr.ErrInvalidShipmentList.Error()})
		return
	}
	page := 0
	if request.Page != nil {
		page = *request.Page
	}
	pageSize := 0
	if request.PageSize != nil {
		pageSize = *request.PageSize
	}

	result, err := c.service.ListForMerchant(
		ctx.Request.Context(),
		authctx.MustShopID(ctx),
		model.ListFilter{
			Status:   model.Status(request.Status),
			Page:     page,
			PageSize: pageSize,
		},
	)
	if err != nil {
		writeLogisticsError(ctx, "list merchant shipments", err)
		return
	}

	items := make([]logisticsdto.ShipmentSummaryResponse, 0, len(result.Shipments))
	for i := range result.Shipments {
		items = append(items, newShipmentSummaryResponse(&result.Shipments[i]))
	}
	totalPages := 0
	if result.PageSize > 0 {
		totalPages = int((result.Total + int64(result.PageSize) - 1) / int64(result.PageSize))
	}
	ctx.JSON(http.StatusOK, logisticsdto.ShipmentListResponse{
		Items:      items,
		Page:       result.Page,
		PageSize:   result.PageSize,
		Total:      result.Total,
		TotalPages: totalPages,
	})
}

// GetForMerchant returns one shipment owned by the authenticated shop.
//
// @Summary     Get a merchant shipment
// @Description Get a shipment and its full tracking timeline for the authenticated shop.
// @Tags        Merchant Logistics
// @Produce     json
// @Param       shipmentNo path string true "Shipment number"
// @Success     200 {object} logisticsdto.ShipmentResponse
// @Failure     400 {object} httpx.ErrorResponse
// @Failure     401 {object} httpx.ErrorResponse
// @Failure     404 {object} httpx.ErrorResponse
// @Failure     500 {object} httpx.ErrorResponse
// @Security    BearerAuth
// @Router      /merchant/shipments/{shipmentNo} [get]
func (c *ShipmentController) GetForMerchant(ctx *gin.Context) {
	shipment, err := c.service.GetForMerchant(
		ctx.Request.Context(),
		authctx.MustShopID(ctx),
		ctx.Param("shipmentNo"),
	)
	if err != nil {
		writeLogisticsError(ctx, "get merchant shipment", err)
		return
	}
	ctx.JSON(http.StatusOK, newShipmentResponse(shipment))
}

// ListForBuyer returns the authenticated buyer's shipments for an order.
//
// @Summary     List order shipments
// @Description List shipments belonging to the authenticated buyer's order.
// @Tags        Logistics
// @Produce     json
// @Param       order_id query int true "Order ID"
// @Success     200 {object} logisticsdto.BuyerShipmentListResponse
// @Failure     400 {object} httpx.ErrorResponse
// @Failure     401 {object} httpx.ErrorResponse
// @Failure     500 {object} httpx.ErrorResponse
// @Security    BearerAuth
// @Router      /shipments [get]
func (c *ShipmentController) ListForBuyer(ctx *gin.Context) {
	var request logisticsdto.ListBuyerShipmentsQuery
	if err := ctx.ShouldBindQuery(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: apperr.ErrInvalidShipmentList.Error()})
		return
	}

	shipments, err := c.service.ListForBuyerOrder(
		ctx.Request.Context(),
		authctx.MustUserID(ctx),
		request.OrderID,
	)
	if err != nil {
		if errors.Is(err, apperr.ErrInvalidShipmentList) {
			ctx.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: apperr.ErrInvalidShipmentList.Error()})
			return
		}
		writeLogisticsError(ctx, "list buyer shipments", err)
		return
	}

	items := make([]logisticsdto.ShipmentSummaryResponse, 0, len(shipments))
	for i := range shipments {
		items = append(items, newShipmentSummaryResponse(&shipments[i]))
	}
	ctx.JSON(http.StatusOK, logisticsdto.BuyerShipmentListResponse{Items: items})
}

// GetForBuyer returns one shipment owned by the authenticated buyer.
//
// @Summary     Get a shipment
// @Description Get an owned shipment and its full tracking timeline.
// @Tags        Logistics
// @Produce     json
// @Param       shipmentNo path string true "Shipment number"
// @Success     200 {object} logisticsdto.ShipmentResponse
// @Failure     400 {object} httpx.ErrorResponse
// @Failure     401 {object} httpx.ErrorResponse
// @Failure     404 {object} httpx.ErrorResponse
// @Failure     500 {object} httpx.ErrorResponse
// @Security    BearerAuth
// @Router      /shipments/{shipmentNo} [get]
func (c *ShipmentController) GetForBuyer(ctx *gin.Context) {
	shipment, err := c.service.GetForBuyer(
		ctx.Request.Context(),
		authctx.MustUserID(ctx),
		ctx.Param("shipmentNo"),
	)
	if err != nil {
		writeLogisticsError(ctx, "get buyer shipment", err)
		return
	}
	ctx.JSON(http.StatusOK, newShipmentResponse(shipment))
}

func newShipmentResponse(shipment *model.Shipment) logisticsdto.ShipmentResponse {
	events := make([]logisticsdto.ShipmentEventResponse, 0, len(shipment.Events))
	for _, event := range shipment.Events {
		events = append(events, logisticsdto.ShipmentEventResponse{
			Sequence:    event.Sequence,
			Status:      string(event.Status),
			EventType:   event.EventType,
			Description: event.Description,
			Location:    event.Location,
			OccurredAt:  event.OccurredAt,
		})
	}
	return logisticsdto.ShipmentResponse{
		ShipmentNo:  shipment.ShipmentNo,
		SubOrderID:  shipment.SubOrderID,
		OrderID:     shipment.OrderID,
		Status:      string(shipment.Status),
		Carrier:     shipment.Carrier,
		TrackingNo:  shipment.ProviderTrackingNo,
		ItemSummary: shipment.ItemSummary,
		Receiver: logisticsdto.PartyResponse{
			Name:    shipment.ReceiverName,
			Phone:   shipment.ReceiverPhone,
			Address: shipment.ReceiverAddress,
		},
		OriginCity:      shipment.OriginCity,
		DestinationCity: shipment.DestinationCity,
		CurrentLocation: shipment.CurrentLocation,
		Events:          events,
		ShippedAt:       shipment.ShippedAt,
		DeliveredAt:     shipment.DeliveredAt,
		CreatedAt:       shipment.CreatedAt,
		UpdatedAt:       shipment.UpdatedAt,
	}
}

func newShipmentSummaryResponse(shipment *model.Summary) logisticsdto.ShipmentSummaryResponse {
	return logisticsdto.ShipmentSummaryResponse{
		ShipmentNo:      shipment.ShipmentNo,
		SubOrderID:      shipment.SubOrderID,
		OrderID:         shipment.OrderID,
		Status:          string(shipment.Status),
		Carrier:         shipment.Carrier,
		TrackingNo:      shipment.ProviderTrackingNo,
		ItemSummary:     shipment.ItemSummary,
		ReceiverName:    shipment.ReceiverName,
		ReceiverPhone:   shipment.ReceiverPhone,
		CurrentLocation: shipment.CurrentLocation,
		ShippedAt:       shipment.ShippedAt,
		DeliveredAt:     shipment.DeliveredAt,
		CreatedAt:       shipment.CreatedAt,
		UpdatedAt:       shipment.UpdatedAt,
	}
}
