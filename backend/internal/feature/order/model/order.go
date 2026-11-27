package model

import "time"

const PendingPaymentTTL = 30 * time.Minute

type OrderStatus string

const (
	OrderStatusPendingPayment   OrderStatus = "pending_payment"
	OrderStatusPaid             OrderStatus = "paid"
	OrderStatusPartiallyShipped OrderStatus = "partially_shipped"
	OrderStatusShipped          OrderStatus = "shipped"
	OrderStatusCompleted        OrderStatus = "completed"
	OrderStatusCancelled        OrderStatus = "cancelled"
)

type SubOrderStatus string

const (
	SubOrderStatusPendingPayment   SubOrderStatus = "pending_payment"
	SubOrderStatusAwaitingShipment SubOrderStatus = "awaiting_shipment"
	SubOrderStatusShipped          SubOrderStatus = "shipped"
	SubOrderStatusCompleted        SubOrderStatus = "completed"
	SubOrderStatusCancelled        SubOrderStatus = "cancelled"
)

type Order struct {
	ID               int
	UserID           int
	CheckoutKey      string
	Status           OrderStatus
	TotalAmountCents int64
	ReceiverName     string
	ReceiverPhone    string
	ShippingAddress  string
	PaidAt           *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	SubOrders        []SubOrder
}

func (o Order) ExpiresAt() time.Time {
	return o.CreatedAt.Add(PendingPaymentTTL)
}

type SubOrder struct {
	ID               int
	OrderID          int
	ShopID           int
	ShopName         string
	Status           SubOrderStatus
	TotalAmountCents int64
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Items            []OrderItem
}

type OrderItem struct {
	ID              int
	SubOrderID      int
	ProductID       int
	SKUID           int
	ProductName     string
	Specs           map[string]string
	UnitPriceCents  int64
	Quantity        int
	LineAmountCents int64
}

type OrderSummary struct {
	ID               int
	Status           OrderStatus
	TotalAmountCents int64
	ItemCount        int
	ShopCount        int
	CreatedAt        time.Time
}

func (s OrderSummary) ExpiresAt() time.Time {
	return s.CreatedAt.Add(PendingPaymentTTL)
}

type OrderListFilter struct {
	Page     int
	PageSize int
}

type OrderListResult struct {
	Orders   []OrderSummary
	Page     int
	PageSize int
	Total    int64
}

type MerchantOrderListFilter struct {
	Status   SubOrderStatus
	Page     int
	PageSize int
}

type MerchantOrderSummary struct {
	ID               int
	OrderID          int
	Status           SubOrderStatus
	TotalAmountCents int64
	ItemCount        int
	PaidAt           *time.Time
	CreatedAt        time.Time
}

type MerchantOrderDetail struct {
	ID               int
	OrderID          int
	Status           SubOrderStatus
	TotalAmountCents int64
	ItemCount        int
	PaidAt           *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Items            []OrderItem
}

type MerchantOrderReceiver struct {
	ReceiverName    string
	ReceiverPhone   string
	ShippingAddress string
}

type MerchantOrderListResult struct {
	Orders   []MerchantOrderSummary
	Page     int
	PageSize int
	Total    int64
}

type CreateOrderItem struct {
	CartItemID int
	Quantity   int
}

type CreateOrderInput struct {
	CheckoutKey     string
	RequestHash     string
	Items           []CreateOrderItem
	ReceiverName    string
	ReceiverPhone   string
	ShippingAddress string
}

type CreateDirectOrderInput struct {
	CheckoutKey     string
	RequestHash     string
	SKUID           int
	Quantity        int
	ReceiverName    string
	ReceiverPhone   string
	ShippingAddress string
}

type ReservationItem struct {
	SKUID    int
	Quantity int
}

type CancelTarget struct {
	OrderID int
	UserID  int
	Status  OrderStatus
	Items   []ReservationItem
}
