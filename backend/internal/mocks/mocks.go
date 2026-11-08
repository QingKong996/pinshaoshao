// Package mocks provides generated gomock test doubles for service interfaces.
package mocks

//go:generate go tool mockgen -source=../service/user.go -destination=mock_user_repository.go -package=mocks
//go:generate go tool mockgen -source=../service/shop.go -destination=mock_shop_repository.go -package=mocks
//go:generate go tool mockgen -source=../service/product_sku.go -destination=mock_product_sku.go -package=mocks
