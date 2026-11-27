package dto

import "time"

type SaveAddressRequest struct {
	ReceiverName    string `json:"receiver_name" binding:"required,max=50" example:"张三"`
	ReceiverPhone   string `json:"receiver_phone" binding:"required,max=32" example:"13800138000"`
	ShippingAddress string `json:"shipping_address" binding:"required,max=1000" example:"上海市浦东新区世纪大道 100 号"`
	IsDefault       bool   `json:"is_default"`
}

type AddressResponse struct {
	ID              int       `json:"id" example:"1"`
	ReceiverName    string    `json:"receiver_name" example:"张三"`
	ReceiverPhone   string    `json:"receiver_phone" example:"13800138000"`
	ShippingAddress string    `json:"shipping_address" example:"上海市浦东新区世纪大道 100 号"`
	IsDefault       bool      `json:"is_default" example:"true"`
	CreatedAt       time.Time `json:"created_at" example:"2026-11-27T12:00:00Z"`
	UpdatedAt       time.Time `json:"updated_at" example:"2026-11-27T12:00:00Z"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"invalid address"`
}
