package model

import "time"

type ShopStatus string

const (
	ShopStatusActive ShopStatus = "active"
	ShopStatusBanned ShopStatus = "banned"
)

type Shop struct {
	ID           int
	Username     string
	Name         string
	Description  string
	PasswordHash string
	Email        string
	Status       ShopStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
