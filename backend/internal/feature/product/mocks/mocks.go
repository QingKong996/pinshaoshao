// Package mocks provides generated gomock test doubles for product interfaces.
package mocks

//go:generate go tool mockgen -source=../service/product.go -destination=mock_product_repository.go -package=mocks
//go:generate go tool mockgen -source=../service/product_category.go -destination=mock_product_category_repository.go -package=mocks
//go:generate go tool mockgen -source=../service/product_sku.go -destination=mock_product_sku.go -package=mocks
//go:generate go tool mockgen -source=../controller/product.go -destination=mock_product_service.go -package=mocks
//go:generate go tool mockgen -source=../controller/product_category.go -destination=mock_product_category_service.go -package=mocks
//go:generate go tool mockgen -source=../controller/product_sku.go -destination=mock_product_sku_service.go -package=mocks
