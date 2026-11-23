package dto

import "time"

type CreateOrderRequest struct {
	CheckoutKey     string                 `json:"checkout_key" binding:"required" example:"4f2614df-f674-4b5e-a9e4-3fc5e7f62013"`
	Items           []CreateOrderItemInput `json:"items" binding:"required,min=1,max=100,dive"`
	ReceiverName    string                 `json:"receiver_name" binding:"required,max=50" example:"张三"`
	ReceiverPhone   string                 `json:"receiver_phone" binding:"required,max=32" example:"13800138000"`
	ShippingAddress string                 `json:"shipping_address" binding:"required,max=1000" example:"上海市浦东新区世纪大道 100 号"`
}

type CreateOrderItemInput struct {
	CartItemID int `json:"cart_item_id" binding:"required,min=1" example:"1"`
	Quantity   int `json:"quantity" binding:"required,min=1,max=2147483647" example:"2"`
}

type CreateDirectOrderRequest struct {
	CheckoutKey     string `json:"checkout_key" binding:"required" example:"4f2614df-f674-4b5e-a9e4-3fc5e7f62013"`
	SKUID           int    `json:"sku_id" binding:"required,min=1" example:"5"`
	Quantity        int    `json:"quantity" binding:"required,min=1,max=2147483647" example:"2"`
	ReceiverName    string `json:"receiver_name" binding:"required,max=50" example:"张三"`
	ReceiverPhone   string `json:"receiver_phone" binding:"required,max=32" example:"13800138000"`
	ShippingAddress string `json:"shipping_address" binding:"required,max=1000" example:"上海市浦东新区世纪大道 100 号"`
}

type ListOrdersQuery struct {
	Page     *int `form:"page" binding:"omitempty,min=1,max=1000000" example:"1"`
	PageSize *int `form:"page_size" binding:"omitempty,min=1,max=100" example:"10"`
}

type OrderResponse struct {
	ID               int                `json:"id" example:"1"`
	Status           string             `json:"status" example:"pending_payment"`
	TotalAmountCents int64              `json:"total_amount_cents" example:"3998"`
	ReceiverName     string             `json:"receiver_name" example:"张三"`
	ReceiverPhone    string             `json:"receiver_phone" example:"13800138000"`
	ShippingAddress  string             `json:"shipping_address" example:"上海市浦东新区世纪大道 100 号"`
	ExpiresAt        time.Time          `json:"expires_at" example:"2026-11-20T12:30:00Z"`
	PaidAt           *time.Time         `json:"paid_at"`
	CreatedAt        time.Time          `json:"created_at" example:"2026-11-20T12:00:00Z"`
	UpdatedAt        time.Time          `json:"updated_at" example:"2026-11-20T12:00:00Z"`
	SubOrders        []SubOrderResponse `json:"sub_orders"`
}

type SubOrderResponse struct {
	ID               int                 `json:"id" example:"1"`
	ShopID           int                 `json:"shop_id" example:"3"`
	ShopName         string              `json:"shop_name" example:"示例店铺"`
	Status           string              `json:"status" example:"pending_payment"`
	TotalAmountCents int64               `json:"total_amount_cents" example:"3998"`
	Items            []OrderItemResponse `json:"items"`
}

type OrderItemResponse struct {
	ID              int               `json:"id" example:"1"`
	ProductID       int               `json:"product_id" example:"10"`
	SKUID           int               `json:"sku_id" example:"5"`
	ProductName     string            `json:"product_name" example:"纯棉T恤"`
	SKUSpecs        map[string]string `json:"sku_specs"`
	UnitPriceCents  int64             `json:"unit_price_cents" example:"1999"`
	Quantity        int               `json:"quantity" example:"2"`
	LineAmountCents int64             `json:"line_amount_cents" example:"3998"`
}

type OrderSummaryResponse struct {
	ID               int       `json:"id" example:"1"`
	Status           string    `json:"status" example:"pending_payment"`
	TotalAmountCents int64     `json:"total_amount_cents" example:"3998"`
	ItemCount        int       `json:"item_count" example:"2"`
	ShopCount        int       `json:"shop_count" example:"1"`
	ExpiresAt        time.Time `json:"expires_at" example:"2026-11-20T12:30:00Z"`
	CreatedAt        time.Time `json:"created_at" example:"2026-11-20T12:00:00Z"`
}

type OrderListResponse struct {
	Items      []OrderSummaryResponse `json:"items"`
	Page       int                    `json:"page" example:"1"`
	PageSize   int                    `json:"page_size" example:"10"`
	Total      int64                  `json:"total" example:"25"`
	TotalPages int                    `json:"total_pages" example:"3"`
}
