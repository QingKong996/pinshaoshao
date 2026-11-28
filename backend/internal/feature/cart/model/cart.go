package model

import "time"

type CartItemAvailability string

const (
	CartItemAvailable         CartItemAvailability = "available"
	CartItemProductDeleted    CartItemAvailability = "product_deleted"
	CartItemShopBanned        CartItemAvailability = "shop_banned"
	CartItemProductOffShelf   CartItemAvailability = "product_off_shelf"
	CartItemSKUDeleted        CartItemAvailability = "sku_deleted"
	CartItemSKUDisabled       CartItemAvailability = "sku_disabled"
	CartItemOutOfStock        CartItemAvailability = "out_of_stock"
	CartItemInsufficientStock CartItemAvailability = "insufficient_stock"
)

type CartItem struct {
	ID             int
	UserID         int
	ProductID      int
	SKUID          int
	ProductName    string
	Specs          map[string]string
	UnitPriceCents int64
	Quantity       int
	CreatedAt      time.Time
	UpdatedAt      time.Time

	ProductStatus  *string
	ShopStatus     *string
	SKUEnabled     *bool
	AvailableStock *int
}

func (i CartItem) Availability() CartItemAvailability {
	switch {
	case i.ProductStatus == nil:
		return CartItemProductDeleted
	case i.ShopStatus == nil || *i.ShopStatus != "active":
		return CartItemShopBanned
	case *i.ProductStatus != "on_shelf":
		return CartItemProductOffShelf
	case i.SKUEnabled == nil:
		return CartItemSKUDeleted
	case !*i.SKUEnabled:
		return CartItemSKUDisabled
	case i.AvailableStock == nil || *i.AvailableStock <= 0:
		return CartItemOutOfStock
	case *i.AvailableStock < i.Quantity:
		return CartItemInsufficientStock
	default:
		return CartItemAvailable
	}
}

func (i CartItem) LineTotalCents() int64 {
	return i.UnitPriceCents * int64(i.Quantity)
}
