// Package mocks provides generated gomock test doubles for user interfaces.
package mocks

//go:generate go tool mockgen -source=../service/user.go -destination=mock_user_repository.go -package=mocks
//go:generate go tool mockgen -source=../controller/user.go -destination=mock_user_service.go -package=mocks
