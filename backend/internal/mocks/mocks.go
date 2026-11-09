// Package mocks provides generated gomock test doubles for service interfaces.
package mocks

//go:generate go tool mockgen -source=../service/user.go -destination=mock_user_repository.go -package=mocks
//go:generate go tool mockgen -source=../service/shop.go -destination=mock_shop_repository.go -package=mocks
//go:generate go tool mockgen -source=../service/product_sku.go -destination=mock_product_sku.go -package=mocks
//go:generate go tool mockgen -source=../service/product_category.go -destination=mock_product_category_repository.go -package=mocks
//go:generate go tool mockgen -source=../controller/user.go -destination=mock_user_service.go -package=mocks
//go:generate go tool mockgen -source=../controller/shop.go -destination=mock_shop_service.go -package=mocks
//go:generate go tool mockgen -source=../controller/product_sku.go -destination=mock_product_sku_service.go -package=mocks
//go:generate go tool mockgen -source=../controller/product_category.go -destination=mock_product_category_service.go -package=mocks
