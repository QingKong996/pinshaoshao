// Package mocks provides generated gomock test doubles for inventory interfaces.
package mocks

//go:generate go tool mockgen -source=../service/inventory.go -destination=mock_inventory_repository.go -package=mocks
//go:generate go tool mockgen -source=../controller/inventory.go -destination=mock_inventory_service.go -package=mocks
