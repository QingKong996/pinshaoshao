// Package mocks provides generated gomock test doubles for shop interfaces.
package mocks

//go:generate go tool mockgen -source=../service/shop.go -destination=mock_shop_repository.go -package=mocks
//go:generate go tool mockgen -source=../service/dashboard.go -destination=mock_dashboard_repository.go -package=mocks
//go:generate go tool mockgen -source=../controller/shop.go -destination=mock_shop_service.go -package=mocks
//go:generate go tool mockgen -source=../controller/dashboard.go -destination=mock_dashboard_service.go -package=mocks
