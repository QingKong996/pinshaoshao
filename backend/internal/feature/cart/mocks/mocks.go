// Package mocks provides generated gomock test doubles for cart interfaces.
package mocks

//go:generate go tool mockgen -source=../service/cart.go -destination=mock_cart_repository.go -package=mocks
//go:generate go tool mockgen -source=../controller/cart.go -destination=mock_cart_service.go -package=mocks
