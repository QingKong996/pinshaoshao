package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"pinshaoshao/internal/feature/logistics/apperr"
	"pinshaoshao/internal/feature/logistics/model"
	"pinshaoshao/internal/feature/order/shippingport"

	"github.com/google/uuid"
)

const (
	providerMockLogistics = "mocklogistics"
	maxCarrierLength      = 64
	defaultPage           = 1
	defaultPageSize       = 10
	maxPage               = 1_000_000
	maxPageSize           = 100
)

var shipmentNamespace = uuid.MustParse("6f1b9b76-6f19-4f2b-9816-94c1e7d0063d")

type ShipmentRepository interface {
	GetBySubOrder(ctx context.Context, shopID int, subOrderID int) (*model.Shipment, bool, error)
	Create(
		ctx context.Context,
		shipment *model.Shipment,
		target *shippingport.Target,
	) (*model.Shipment, bool, error)
	GetForMerchant(ctx context.Context, shopID int, shipmentNo string) (*model.Shipment, error)
	ListForMerchant(ctx context.Context, shopID int, filter model.ListFilter) (*model.ListResult, error)
	GetForBuyer(ctx context.Context, userID int, shipmentNo string) (*model.Shipment, error)
	ListForBuyerOrder(ctx context.Context, userID int, orderID int) ([]model.Summary, error)
	ApplyEvent(ctx context.Context, callback model.Callback) (*model.Shipment, error)
}

type ShippingGateway interface {
	Dispatch(ctx context.Context, input model.ProviderShipmentRequest) (*model.ProviderShipmentResult, error)
}

type Config struct {
	NotifyURL string
	Gateway   ShippingGateway
}

type ShipmentService struct {
	repository ShipmentRepository
	orders     shippingport.Lifecycle
	gateway    ShippingGateway
	notifyURL  string
}

func NewShipmentService(
	repository ShipmentRepository,
	orders shippingport.Lifecycle,
	config Config,
) *ShipmentService {
	return &ShipmentService{
		repository: repository,
		orders:     orders,
		gateway:    config.Gateway,
		notifyURL:  strings.TrimSpace(config.NotifyURL),
	}
}

func (s *ShipmentService) Ship(
	ctx context.Context,
	shopID int,
	subOrderID int,
	carrier string,
) (*model.Shipment, bool, error) {
	if shopID <= 0 || subOrderID <= 0 {
		return nil, false, apperr.ErrInvalidSubOrderID
	}
	carrier = strings.TrimSpace(carrier)
	if utf8.RuneCountInString(carrier) > maxCarrierLength {
		return nil, false, apperr.ErrInvalidShipmentInput
	}

	existing, found, err := s.repository.GetBySubOrder(ctx, shopID, subOrderID)
	if err != nil {
		return nil, false, fmt.Errorf("service.Ship: %w", err)
	}
	if found {
		return existing, false, nil
	}

	target, err := s.orders.GetShipmentTarget(ctx, shopID, subOrderID)
	if err != nil {
		return nil, false, fmt.Errorf("service.Ship: %w", mapOrderError(err))
	}
	switch target.Status {
	case "awaiting_shipment":
	case "shipped", "completed":
		return nil, false, apperr.ErrShipmentConflict
	default:
		return nil, false, apperr.ErrOrderNotShippable
	}

	shipmentNo := uuid.NewSHA1(shipmentNamespace, []byte(fmt.Sprintf("sub-order:%d", subOrderID)))
	result, err := s.gateway.Dispatch(ctx, model.ProviderShipmentRequest{
		MerchantOrderNo: shipmentNo.String(),
		ItemSummary:     target.ItemSummary,
		Carrier:         carrier,
		ReceiverName:    target.ReceiverName,
		ReceiverPhone:   target.ReceiverPhone,
		ReceiverAddress: target.ShippingAddress,
		NotifyURL:       s.notifyURL,
	})
	if err != nil {
		return nil, false, fmt.Errorf("service.Ship: %w", err)
	}
	if err := validateProviderResult(shipmentNo.String(), result); err != nil {
		return nil, false, fmt.Errorf("service.Ship: %w", err)
	}

	shippedAt := time.Now().UTC()
	if result.DispatchedAt != nil {
		shippedAt = result.DispatchedAt.UTC()
	}
	shipment := &model.Shipment{
		ShipmentNo:         shipmentNo.String(),
		SubOrderID:         target.SubOrderID,
		OrderID:            target.OrderID,
		UserID:             target.UserID,
		ShopID:             target.ShopID,
		Provider:           providerMockLogistics,
		ProviderShipmentNo: result.ShipmentNo,
		ProviderTrackingNo: result.TrackingNo,
		Carrier:            result.Carrier,
		Status:             result.Status,
		ItemSummary:        target.ItemSummary,
		ReceiverName:       target.ReceiverName,
		ReceiverPhone:      target.ReceiverPhone,
		ReceiverAddress:    target.ShippingAddress,
		OriginCity:         result.OriginCity,
		DestinationCity:    result.DestinationCity,
		CurrentLocation:    result.CurrentLocation,
		ShippedAt:          shippedAt,
	}

	stored, created, err := s.repository.Create(ctx, shipment, target)
	if err != nil {
		return nil, false, fmt.Errorf("service.Ship: %w", err)
	}
	return stored, created, nil
}

func (s *ShipmentService) GetForMerchant(
	ctx context.Context,
	shopID int,
	shipmentNo string,
) (*model.Shipment, error) {
	if shopID <= 0 {
		return nil, apperr.ErrShipmentNotFound
	}
	normalized, err := normalizeShipmentNo(shipmentNo)
	if err != nil {
		return nil, err
	}
	shipment, err := s.repository.GetForMerchant(ctx, shopID, normalized)
	if err != nil {
		return nil, fmt.Errorf("service.GetForMerchant: %w", err)
	}
	return shipment, nil
}

func (s *ShipmentService) ListForMerchant(
	ctx context.Context,
	shopID int,
	filter model.ListFilter,
) (*model.ListResult, error) {
	normalized, err := normalizeListFilter(filter)
	if err != nil {
		return nil, err
	}
	if shopID <= 0 {
		return nil, apperr.ErrShipmentNotFound
	}
	result, err := s.repository.ListForMerchant(ctx, shopID, normalized)
	if err != nil {
		return nil, fmt.Errorf("service.ListForMerchant: %w", err)
	}
	return result, nil
}

func (s *ShipmentService) GetForBuyer(
	ctx context.Context,
	userID int,
	shipmentNo string,
) (*model.Shipment, error) {
	if userID <= 0 {
		return nil, apperr.ErrShipmentNotFound
	}
	normalized, err := normalizeShipmentNo(shipmentNo)
	if err != nil {
		return nil, err
	}
	shipment, err := s.repository.GetForBuyer(ctx, userID, normalized)
	if err != nil {
		return nil, fmt.Errorf("service.GetForBuyer: %w", err)
	}
	return shipment, nil
}

func (s *ShipmentService) ListForBuyerOrder(
	ctx context.Context,
	userID int,
	orderID int,
) ([]model.Summary, error) {
	if userID <= 0 || orderID <= 0 {
		return nil, apperr.ErrInvalidShipmentList
	}
	shipments, err := s.repository.ListForBuyerOrder(ctx, userID, orderID)
	if err != nil {
		return nil, fmt.Errorf("service.ListForBuyerOrder: %w", err)
	}
	return shipments, nil
}

func (s *ShipmentService) HandleCallback(
	ctx context.Context,
	callback model.Callback,
) (*model.Shipment, error) {
	eventID, err := uuid.Parse(strings.TrimSpace(callback.EventID))
	if err != nil || eventID == uuid.Nil {
		return nil, fmt.Errorf("service.HandleCallback: %w", apperr.ErrInvalidCallback)
	}
	shipmentNo, err := uuid.Parse(strings.TrimSpace(callback.ShipmentNo))
	if err != nil || shipmentNo == uuid.Nil {
		return nil, fmt.Errorf("service.HandleCallback: %w", apperr.ErrInvalidCallback)
	}
	providerShipmentNo, err := uuid.Parse(strings.TrimSpace(callback.ProviderShipmentNo))
	if err != nil || providerShipmentNo == uuid.Nil {
		return nil, fmt.Errorf("service.HandleCallback: %w", apperr.ErrInvalidCallback)
	}
	callback.EventType = strings.TrimSpace(callback.EventType)
	if !model.ValidEventType(callback.EventType) ||
		!callback.Status.Valid() ||
		model.EventTypeForStatus(callback.Status) != callback.EventType ||
		callback.Sequence <= 0 ||
		callback.OccurredAt.IsZero() ||
		strings.TrimSpace(callback.Description) == "" ||
		strings.TrimSpace(callback.Location) == "" {
		return nil, fmt.Errorf("service.HandleCallback: %w", apperr.ErrInvalidCallback)
	}

	callback.EventID = eventID.String()
	callback.ShipmentNo = shipmentNo.String()
	callback.ProviderShipmentNo = providerShipmentNo.String()
	callback.EventType = strings.TrimSpace(callback.EventType)
	callback.ProviderTrackingNo = strings.TrimSpace(callback.ProviderTrackingNo)
	callback.Description = strings.TrimSpace(callback.Description)
	callback.Location = strings.TrimSpace(callback.Location)
	callback.OccurredAt = callback.OccurredAt.UTC()

	shipment, err := s.repository.ApplyEvent(ctx, callback)
	if err != nil {
		return nil, fmt.Errorf("service.HandleCallback: %w", err)
	}
	return shipment, nil
}

func normalizeShipmentNo(value string) (string, error) {
	parsed, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil || parsed == uuid.Nil {
		return "", apperr.ErrInvalidShipmentNo
	}
	return parsed.String(), nil
}

func normalizeListFilter(filter model.ListFilter) (model.ListFilter, error) {
	if filter.Page == 0 {
		filter.Page = defaultPage
	}
	if filter.PageSize == 0 {
		filter.PageSize = defaultPageSize
	}
	if filter.Page < 1 ||
		filter.Page > maxPage ||
		filter.PageSize < 1 ||
		filter.PageSize > maxPageSize {
		return model.ListFilter{}, apperr.ErrInvalidShipmentList
	}
	if filter.Status != "" && !filter.Status.Valid() {
		return model.ListFilter{}, apperr.ErrInvalidShipmentList
	}
	return filter, nil
}

func validateProviderResult(
	merchantNo string,
	result *model.ProviderShipmentResult,
) error {
	if result == nil {
		return apperr.ErrProviderInvalid
	}
	providerNo, err := uuid.Parse(strings.TrimSpace(result.ShipmentNo))
	if err != nil ||
		providerNo == uuid.Nil ||
		result.MerchantNo != merchantNo ||
		strings.TrimSpace(result.TrackingNo) == "" ||
		strings.TrimSpace(result.Carrier) == "" ||
		!result.Status.Valid() ||
		result.Status != model.StatusPendingPickup {
		return apperr.ErrProviderInvalid
	}
	result.ShipmentNo = providerNo.String()
	result.TrackingNo = strings.TrimSpace(result.TrackingNo)
	result.Carrier = strings.TrimSpace(result.Carrier)
	return nil
}

func mapOrderError(err error) error {
	switch {
	case errors.Is(err, shippingport.ErrSubOrderNotFound):
		return apperr.ErrShipmentNotFound
	case errors.Is(err, shippingport.ErrSubOrderNotShippable):
		return apperr.ErrOrderNotShippable
	default:
		return err
	}
}
