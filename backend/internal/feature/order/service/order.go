package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"pinshaoshao/internal/feature/order/apperr"
	"pinshaoshao/internal/feature/order/model"
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
	List(ctx context.Context, userID int, filter model.OrderListFilter) (*model.OrderListResult, error)
	Get(ctx context.Context, userID int, orderID int) (*model.Order, error)
	Cancel(ctx context.Context, userID int, orderID int) (*model.Order, error)
}

type CreateOrderInput struct {
	CheckoutKey     string
	Items           []model.CreateOrderItem
	ReceiverName    string
	ReceiverPhone   string
	ShippingAddress string
}

type OrderService struct {
	orderRepository OrderRepository
}

// NewOrderService builds the order service from its repository.
func NewOrderService(orderRepository OrderRepository) *OrderService {
	return &OrderService{orderRepository: orderRepository}
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

	order, err := s.orderRepository.Cancel(ctx, userID, orderID)
	if err != nil {
		return nil, fmt.Errorf("service.CancelOrder: %w", err)
	}
	return order, nil
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

	receiverName := strings.TrimSpace(input.ReceiverName)
	receiverPhone := strings.TrimSpace(input.ReceiverPhone)
	shippingAddress := strings.TrimSpace(input.ShippingAddress)
	if receiverName == "" || utf8.RuneCountInString(receiverName) > maxReceiverNameLen {
		return CreateOrderInput{}, "", apperr.ErrInvalidCheckout
	}
	if receiverPhone == "" || utf8.RuneCountInString(receiverPhone) > maxReceiverPhoneLen {
		return CreateOrderInput{}, "", apperr.ErrInvalidCheckout
	}
	if shippingAddress == "" || utf8.RuneCountInString(shippingAddress) > maxAddressLen {
		return CreateOrderInput{}, "", apperr.ErrInvalidCheckout
	}

	normalized := CreateOrderInput{
		CheckoutKey:     checkoutKey.String(),
		Items:           items,
		ReceiverName:    receiverName,
		ReceiverPhone:   receiverPhone,
		ShippingAddress: shippingAddress,
	}
	return normalized, requestHash(normalized), nil
}

func requestHash(input CreateOrderInput) string {
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
	encoded, _ := json.Marshal(payload)
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}
