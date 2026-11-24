package logisticsdto

import "time"

type CreateShipmentRequest struct {
	SubOrderID int    `json:"sub_order_id" binding:"required,min=1" example:"12"`
	Carrier    string `json:"carrier" binding:"omitempty,max=64" example:"顺丰速运"`
}

type ListShipmentsQuery struct {
	Page     *int   `form:"page" binding:"omitempty,min=1,max=1000000" example:"1"`
	PageSize *int   `form:"page_size" binding:"omitempty,min=1,max=100" example:"10"`
	Status   string `form:"status" binding:"omitempty,oneof=pending_pickup picked_up in_transit at_transit_hub out_for_delivery delivered" example:"in_transit"`
}

type ListBuyerShipmentsQuery struct {
	OrderID int `form:"order_id" binding:"required,min=1" example:"1"`
}

type PartyResponse struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

type ShipmentEventResponse struct {
	Sequence    int       `json:"sequence"`
	Status      string    `json:"status"`
	EventType   string    `json:"event_type"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	OccurredAt  time.Time `json:"occurred_at"`
}

type ShipmentResponse struct {
	ShipmentNo      string                  `json:"shipment_no"`
	SubOrderID      int                     `json:"sub_order_id"`
	OrderID         int                     `json:"order_id"`
	Status          string                  `json:"status"`
	Carrier         string                  `json:"carrier"`
	TrackingNo      string                  `json:"tracking_no"`
	ItemSummary     string                  `json:"item_summary"`
	Receiver        PartyResponse           `json:"receiver"`
	OriginCity      string                  `json:"origin_city"`
	DestinationCity string                  `json:"destination_city"`
	CurrentLocation string                  `json:"current_location"`
	Events          []ShipmentEventResponse `json:"events"`
	ShippedAt       time.Time               `json:"shipped_at"`
	DeliveredAt     *time.Time              `json:"delivered_at"`
	CreatedAt       time.Time               `json:"created_at"`
	UpdatedAt       time.Time               `json:"updated_at"`
}

type ShipmentSummaryResponse struct {
	ShipmentNo      string     `json:"shipment_no"`
	SubOrderID      int        `json:"sub_order_id"`
	OrderID         int        `json:"order_id"`
	Status          string     `json:"status"`
	Carrier         string     `json:"carrier"`
	TrackingNo      string     `json:"tracking_no"`
	ItemSummary     string     `json:"item_summary"`
	ReceiverName    string     `json:"receiver_name"`
	ReceiverPhone   string     `json:"receiver_phone"`
	CurrentLocation string     `json:"current_location"`
	ShippedAt       time.Time  `json:"shipped_at"`
	DeliveredAt     *time.Time `json:"delivered_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type ShipmentListResponse struct {
	Items      []ShipmentSummaryResponse `json:"items"`
	Page       int                       `json:"page"`
	PageSize   int                       `json:"page_size"`
	Total      int64                     `json:"total"`
	TotalPages int                       `json:"total_pages"`
}

type BuyerShipmentListResponse struct {
	Items []ShipmentSummaryResponse `json:"items"`
}

type ShipmentCallbackRequest struct {
	ID         string               `json:"id"`
	Type       string               `json:"type"`
	OccurredAt time.Time            `json:"occurred_at"`
	Data       ShipmentCallbackData `json:"data"`
}

type ShipmentCallbackData struct {
	ShipmentNo      string    `json:"shipment_no"`
	TrackingNo      string    `json:"tracking_no"`
	MerchantOrderNo string    `json:"merchant_order_no"`
	Status          string    `json:"status"`
	Carrier         string    `json:"carrier"`
	Sequence        int       `json:"sequence"`
	Description     string    `json:"description"`
	Location        string    `json:"location"`
	OccurredAt      time.Time `json:"occurred_at"`
}
