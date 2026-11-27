package model

import "time"

type Address struct {
	ID              int
	UserID          int
	ReceiverName    string
	ReceiverPhone   string
	ShippingAddress string
	IsDefault       bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
