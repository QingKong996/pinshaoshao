package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"pinshaoshao/internal/database"
	"pinshaoshao/internal/feature/inventory"
	inventorymodel "pinshaoshao/internal/feature/inventory/model"
	"pinshaoshao/internal/feature/order/apperr"
	"pinshaoshao/internal/feature/order/model"
	"pinshaoshao/internal/feature/order/paymentport"
	"pinshaoshao/internal/outbox"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	maxCheckoutItems     = 100
	maxReceiverNameLen   = 50
	maxReceiverPhoneLen  = 32
	maxAddressLen        = 1000
	defaultOrderPage     = 1
	defaultOrderPageSize = 10
	maxOrderPage         = 1_000_000
	maxOrderPageSize     = 100
)

type OrderRepository interface {
	Create(ctx context.Context, userID int, input model.CreateOrderInput) (*model.Order, error)
	CreateDirect(ctx context.Context, userID int, input model.CreateDirectOrderInput) (*model.Order, error)
	List(ctx context.Context, userID int, filter model.OrderListFilter) (*model.OrderListResult, error)
	ListMerchant(ctx context.Context, shopID int, filter model.MerchantOrderListFilter) (*model.MerchantOrderListResult, error)
	GetMerchant(ctx context.Context, shopID int, subOrderID int) (*model.MerchantOrderDetail, error)
	Get(ctx context.Context, userID int, orderID int) (*model.Order, error)
	LockForCancel(ctx context.Context, db database.DBTX, userID int, orderID int) (*model.CancelTarget, error)
	LockDueForCancel(ctx context.Context, db database.DBTX, orderID int) (*model.CancelTarget, error)
	MarkCancelled(ctx context.Context, db database.DBTX, orderID int) error
	Load(ctx context.Context, db database.DBTX, userID int, orderID int) (*model.Order, error)
	ListDueOrderIDs(ctx context.Context, limit int) ([]int, error)
}

type PaymentCloser interface {
	CloseForOrder(ctx context.Context, db database.DBTX, orderID int) (*paymentport.CloseResult, error)
}

type OutboxWriter interface {
	Enqueue(ctx context.Context, db database.DBTX, event outbox.Event) error
}

type Transactor interface {
	WithinTx(ctx context.Context, fn func(context.Context, database.DBTX) error) error
}

type CreateOrderInput struct {
	CheckoutKey     string
	Items           []model.CreateOrderItem
	ReceiverName    string
	ReceiverPhone   string
	ShippingAddress string
}

type CreateDirectOrderInput struct {
	CheckoutKey     string
	SKUID           int
	Quantity        int
	ReceiverName    string
	ReceiverPhone   string
	ShippingAddress string
}

type OrderService struct {
	orderRepository OrderRepository
	reservation     inventory.Reservation
	paymentCloser   PaymentCloser
	outboxWriter    OutboxWriter
	transactor      Transactor
}

// NewOrderService builds the order service from its repository.
func NewOrderService(
	orderRepository OrderRepository,
	reservation inventory.Reservation,
	paymentCloser PaymentCloser,
	outboxWriter OutboxWriter,
	transactor Transactor,
) *OrderService {
	return &OrderService{
		orderRepository: orderRepository,
		reservation:     reservation,
		paymentCloser:   paymentCloser,
		outboxWriter:    outboxWriter,
		transactor:      transactor,
	}
}

func (s *OrderService) SetPaymentCloser(paymentCloser PaymentCloser) {
	s.paymentCloser = paymentCloser
}

func (s *OrderService) Create(ctx context.Context, userID int, input CreateOrderInput) (*model.Order, error) {
	normalized, requestHash, err := normalizeCreateOrderInput(input)
	if err != nil {
		return nil, fmt.Errorf("service.CreateOrder: %w", err)
	}

	order, err := s.orderRepository.Create(ctx, userID, model.CreateOrderInput{
		CheckoutKey:     normalized.CheckoutKey,
		RequestHash:     requestHash,
		Items:           normalized.Items,
		ReceiverName:    normalized.ReceiverName,
		ReceiverPhone:   normalized.ReceiverPhone,
		ShippingAddress: normalized.ShippingAddress,
	})
	if err != nil {
		return nil, fmt.Errorf("service.CreateOrder: %w", err)
	}
	return order, nil
}

func (s *OrderService) CreateDirect(ctx context.Context, userID int, input CreateDirectOrderInput) (*model.Order, error) {
	normalized, requestHash, err := normalizeCreateDirectOrderInput(input)
	if err != nil {
		return nil, fmt.Errorf("service.CreateDirectOrder: %w", err)
	}

	order, err := s.orderRepository.CreateDirect(ctx, userID, model.CreateDirectOrderInput{
		CheckoutKey:     normalized.CheckoutKey,
		RequestHash:     requestHash,
		SKUID:           normalized.SKUID,
		Quantity:        normalized.Quantity,
		ReceiverName:    normalized.ReceiverName,
		ReceiverPhone:   normalized.ReceiverPhone,
		ShippingAddress: normalized.ShippingAddress,
	})
	if err != nil {
		return nil, fmt.Errorf("service.CreateDirectOrder: %w", err)
	}
	return order, nil
}

func (s *OrderService) List(ctx context.Context, userID int, filter model.OrderListFilter) (*model.OrderListResult, error) {
	normalized, err := normalizeOrderListFilter(filter)
	if err != nil {
		return nil, fmt.Errorf("service.ListOrders: %w", err)
	}

	orders, err := s.orderRepository.List(ctx, userID, normalized)
	if err != nil {
		return nil, fmt.Errorf("service.ListOrders: %w", err)
	}
	return orders, nil
}

func (s *OrderService) Get(ctx context.Context, userID int, orderID int) (*model.Order, error) {
	if orderID <= 0 {
		return nil, fmt.Errorf("service.GetOrder: %w", apperr.ErrInvalidOrderID)
	}

	order, err := s.orderRepository.Get(ctx, userID, orderID)
	if err != nil {
		return nil, fmt.Errorf("service.GetOrder: %w", err)
	}
	return order, nil
}

func (s *OrderService) Cancel(ctx context.Context, userID int, orderID int) (*model.Order, error) {
	if orderID <= 0 {
		return nil, fmt.Errorf("service.CancelOrder: %w", apperr.ErrInvalidOrderID)
	}

	var result *model.Order
	err := s.transactor.WithinTx(ctx, func(ctx context.Context, db database.DBTX) error {
		target, err := s.orderRepository.LockForCancel(ctx, db, userID, orderID)
		if err != nil {
			return err
		}
		if target.Status == model.OrderStatusCancelled {
			if err := s.closePayment(ctx, db, target); err != nil {
				return err
			}
			result, err = s.orderRepository.Load(ctx, db, userID, orderID)
			return err
		}
		if target.Status != model.OrderStatusPendingPayment {
			return apperr.ErrOrderCannotCancel
		}
		if err := s.cancelTarget(ctx, db, target); err != nil {
			return err
		}
		result, err = s.orderRepository.Load(ctx, db, userID, orderID)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("service.CancelOrder: %w", err)
	}
	return result, nil
}

func (s *OrderService) CancelForPayment(ctx context.Context, userID int, orderID int) error {
	_, err := s.Cancel(ctx, userID, orderID)
	return err
}

func (s *OrderService) Expire(ctx context.Context, orderID int) error {
	if orderID <= 0 {
		return apperr.ErrInvalidOrderID
	}
	return s.transactor.WithinTx(ctx, func(ctx context.Context, db database.DBTX) error {
		target, err := s.orderRepository.LockDueForCancel(ctx, db, orderID)
		if err != nil {
			return err
		}
		if target == nil {
			return nil
		}
		return s.cancelTarget(ctx, db, target)
	})
}

func (s *OrderService) DueOrderIDs(ctx context.Context, limit int) ([]int, error) {
	return s.orderRepository.ListDueOrderIDs(ctx, limit)
}

func (s *OrderService) cancelTarget(ctx context.Context, db database.DBTX, target *model.CancelTarget) error {
	if len(target.Items) > 0 {
		items := make([]inventorymodel.ReservationItem, 0, len(target.Items))
		for _, item := range target.Items {
			items = append(items, inventorymodel.ReservationItem{
				SKUID:    item.SKUID,
				Quantity: item.Quantity,
			})
		}
		if err := s.reservation.Release(ctx, db, items); err != nil {
			return err
		}
	}
	if err := s.closePayment(ctx, db, target); err != nil {
		return err
	}
	return s.orderRepository.MarkCancelled(ctx, db, target.OrderID)
}

func (s *OrderService) closePayment(ctx context.Context, db database.DBTX, target *model.CancelTarget) error {
	if s.paymentCloser == nil {
		return nil
	}
	result, err := s.paymentCloser.CloseForOrder(ctx, db, target.OrderID)
	if err != nil {
		return err
	}
	if result == nil {
		return nil
	}
	if result.AlreadySettled {
		if target.Status == model.OrderStatusCancelled {
			return nil
		}
		return apperr.ErrOrderCannotCancel
	}
	if !result.Closed || result.ProviderTransactionID == nil {
		return nil
	}
	if s.outboxWriter == nil {
		return apperr.ErrInternalServer
	}
	if result.PaymentID <= 0 {
		return apperr.ErrOrderStateInvalid
	}
	payload, err := json.Marshal(struct {
		PaymentNo             string `json:"payment_no"`
		ProviderTransactionID string `json:"provider_transaction_id"`
	}{
		PaymentNo:             result.PaymentNo,
		ProviderTransactionID: *result.ProviderTransactionID,
	})
	if err != nil {
		return fmt.Errorf("encode payment close event: %w", err)
	}
	if err := s.outboxWriter.Enqueue(ctx, db, outbox.Event{
		Key:           "payment.close:" + result.PaymentNo,
		Topic:         outbox.TopicPaymentCloseProvider,
		AggregateType: "payment",
		AggregateID:   result.PaymentID,
		Payload:       payload,
	}); err != nil {
		return err
	}
	return nil
}

func normalizeOrderListFilter(filter model.OrderListFilter) (model.OrderListFilter, error) {
	if filter.Page == 0 {
		filter.Page = defaultOrderPage
	}
	if filter.PageSize == 0 {
		filter.PageSize = defaultOrderPageSize
	}
	if filter.Page < 1 ||
		filter.Page > maxOrderPage ||
		filter.PageSize < 1 ||
		filter.PageSize > maxOrderPageSize {
		return model.OrderListFilter{}, apperr.ErrInvalidOrderList
	}
	return filter, nil
}

func normalizeCreateOrderInput(input CreateOrderInput) (CreateOrderInput, string, error) {
	checkoutKey, err := uuid.Parse(input.CheckoutKey)
	if err != nil || checkoutKey == uuid.Nil {
		return CreateOrderInput{}, "", apperr.ErrInvalidCheckout
	}
	if len(input.Items) == 0 || len(input.Items) > maxCheckoutItems {
		return CreateOrderInput{}, "", apperr.ErrInvalidCheckout
	}

	seen := make(map[int]struct{}, len(input.Items))
	items := make([]model.CreateOrderItem, 0, len(input.Items))
	for _, item := range input.Items {
		if item.CartItemID <= 0 || item.Quantity <= 0 {
			return CreateOrderInput{}, "", apperr.ErrInvalidCheckout
		}
		if _, exists := seen[item.CartItemID]; exists {
			return CreateOrderInput{}, "", apperr.ErrInvalidCheckout
		}
		seen[item.CartItemID] = struct{}{}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CartItemID < items[j].CartItemID
	})

	delivery, err := normalizeDelivery(
		input.ReceiverName,
		input.ReceiverPhone,
		input.ShippingAddress,
	)
	if err != nil {
		return CreateOrderInput{}, "", err
	}

	normalized := CreateOrderInput{
		CheckoutKey:     checkoutKey.String(),
		Items:           items,
		ReceiverName:    delivery.receiverName,
		ReceiverPhone:   delivery.receiverPhone,
		ShippingAddress: delivery.shippingAddress,
	}
	hash, err := requestHash(normalized)
	if err != nil {
		return CreateOrderInput{}, "", err
	}
	return normalized, hash, nil
}

func normalizeCreateDirectOrderInput(input CreateDirectOrderInput) (CreateDirectOrderInput, string, error) {
	checkoutKey, err := uuid.Parse(input.CheckoutKey)
	if err != nil || checkoutKey == uuid.Nil {
		return CreateDirectOrderInput{}, "", apperr.ErrInvalidCheckout
	}
	if input.SKUID <= 0 || input.Quantity <= 0 || input.Quantity > math.MaxInt32 {
		return CreateDirectOrderInput{}, "", apperr.ErrInvalidCheckout
	}

	delivery, err := normalizeDelivery(
		input.ReceiverName,
		input.ReceiverPhone,
		input.ShippingAddress,
	)
	if err != nil {
		return CreateDirectOrderInput{}, "", err
	}

	normalized := CreateDirectOrderInput{
		CheckoutKey:     checkoutKey.String(),
		SKUID:           input.SKUID,
		Quantity:        input.Quantity,
		ReceiverName:    delivery.receiverName,
		ReceiverPhone:   delivery.receiverPhone,
		ShippingAddress: delivery.shippingAddress,
	}
	hash, err := directRequestHash(normalized)
	if err != nil {
		return CreateDirectOrderInput{}, "", err
	}
	return normalized, hash, nil
}

type orderDelivery struct {
	receiverName    string
	receiverPhone   string
	shippingAddress string
}

func normalizeDelivery(receiverName, receiverPhone, shippingAddress string) (orderDelivery, error) {
	receiverName = strings.TrimSpace(receiverName)
	receiverPhone = strings.TrimSpace(receiverPhone)
	shippingAddress = strings.TrimSpace(shippingAddress)
	if receiverName == "" || utf8.RuneCountInString(receiverName) > maxReceiverNameLen {
		return orderDelivery{}, apperr.ErrInvalidCheckout
	}
	if receiverPhone == "" || utf8.RuneCountInString(receiverPhone) > maxReceiverPhoneLen {
		return orderDelivery{}, apperr.ErrInvalidCheckout
	}
	if shippingAddress == "" || utf8.RuneCountInString(shippingAddress) > maxAddressLen {
		return orderDelivery{}, apperr.ErrInvalidCheckout
	}
	return orderDelivery{
		receiverName:    receiverName,
		receiverPhone:   receiverPhone,
		shippingAddress: shippingAddress,
	}, nil
}

func requestHash(input CreateOrderInput) (string, error) {
	payload := struct {
		Items           []model.CreateOrderItem
		ReceiverName    string
		ReceiverPhone   string
		ShippingAddress string
	}{
		Items:           input.Items,
		ReceiverName:    input.ReceiverName,
		ReceiverPhone:   input.ReceiverPhone,
		ShippingAddress: input.ShippingAddress,
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("hash checkout request: %w", err)
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}

func directRequestHash(input CreateDirectOrderInput) (string, error) {
	payload := struct {
		SKUID           int
		Quantity        int
		ReceiverName    string
		ReceiverPhone   string
		ShippingAddress string
	}{
		SKUID:           input.SKUID,
		Quantity:        input.Quantity,
		ReceiverName:    input.ReceiverName,
		ReceiverPhone:   input.ReceiverPhone,
		ShippingAddress: input.ShippingAddress,
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("hash direct checkout request: %w", err)
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}
