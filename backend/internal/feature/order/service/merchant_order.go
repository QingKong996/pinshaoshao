package service

import (
	"context"
	"fmt"

	"pinshaoshao/internal/feature/order/apperr"
	"pinshaoshao/internal/feature/order/model"
)

func (s *OrderService) ListMerchant(ctx context.Context, shopID int, filter model.MerchantOrderListFilter) (*model.MerchantOrderListResult, error) {
	normalized, err := normalizeMerchantOrderListFilter(filter)
	if err != nil {
		return nil, fmt.Errorf("service.ListMerchantOrders: %w", err)
	}

	orders, err := s.orderRepository.ListMerchant(ctx, shopID, normalized)
	if err != nil {
		return nil, fmt.Errorf("service.ListMerchantOrders: %w", err)
	}
	return orders, nil
}

func (s *OrderService) GetMerchant(ctx context.Context, shopID int, subOrderID int) (*model.MerchantOrderDetail, error) {
	if subOrderID <= 0 {
		return nil, fmt.Errorf("service.GetMerchantOrder: %w", apperr.ErrInvalidOrderID)
	}

	order, err := s.orderRepository.GetMerchant(ctx, shopID, subOrderID)
	if err != nil {
		return nil, fmt.Errorf("service.GetMerchantOrder: %w", err)
	}
	return order, nil
}

func normalizeMerchantOrderListFilter(filter model.MerchantOrderListFilter) (model.MerchantOrderListFilter, error) {
	normalized, err := normalizeOrderListFilter(model.OrderListFilter{
		Page:     filter.Page,
		PageSize: filter.PageSize,
	})
	if err != nil {
		return model.MerchantOrderListFilter{}, err
	}
	if filter.Status != "" && !validSubOrderStatus(filter.Status) {
		return model.MerchantOrderListFilter{}, apperr.ErrInvalidOrderList
	}
	return model.MerchantOrderListFilter{
		Status:   filter.Status,
		Page:     normalized.Page,
		PageSize: normalized.PageSize,
	}, nil
}

func validSubOrderStatus(status model.SubOrderStatus) bool {
	switch status {
	case model.SubOrderStatusPendingPayment,
		model.SubOrderStatusAwaitingShipment,
		model.SubOrderStatusShipped,
		model.SubOrderStatusCompleted,
		model.SubOrderStatusCancelled:
		return true
	default:
		return false
	}
}
