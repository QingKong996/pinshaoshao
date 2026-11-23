// Package mocks provides generated gomock test doubles for order interfaces.
package mocks

//go:generate go tool mockgen -source=../service/order.go -destination=mock_order_repository.go -package=mocks
//go:generate go tool mockgen -source=../controller/order.go -destination=mock_order_service.go -package=mocks
