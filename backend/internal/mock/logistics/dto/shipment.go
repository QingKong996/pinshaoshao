package dto

import "time"

type CreateShipmentRequest struct {
	MerchantOrderNo string `json:"merchant_order_no" binding:"required,max=64" example:"order-20261124-001"`
	ItemSummary     string `json:"item_summary" binding:"omitempty,max=200" example:"Wireless keyboard"`
	Carrier         string `json:"carrier" binding:"omitempty,max=64" example:"顺丰速运"`
	SenderName      string `json:"sender_name" binding:"omitempty,max=100"`
	SenderPhone     string `json:"sender_phone" binding:"omitempty,max=32"`
	SenderAddress   string `json:"sender_address" binding:"omitempty,max=500"`
	ReceiverName    string `json:"receiver_name" binding:"omitempty,max=100"`
	ReceiverPhone   string `json:"receiver_phone" binding:"omitempty,max=32"`
	ReceiverAddress string `json:"receiver_address" binding:"omitempty,max=500"`
	OriginCity      string `json:"origin_city" binding:"omitempty,max=64" example:"上海"`
	DestinationCity string `json:"destination_city" binding:"omitempty,max=64" example:"杭州"`
	NotifyURL       string `json:"notify_url" binding:"required,url,max=2048" example:"http://localhost:8080/mocklogistics/webhook"`
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
	TrackingNo      string                  `json:"tracking_no"`
	MerchantOrderNo string                  `json:"merchant_order_no"`
	Status          string                  `json:"status"`
	Carrier         string                  `json:"carrier"`
	ItemSummary     string                  `json:"item_summary"`
	Sender          PartyResponse           `json:"sender"`
	Receiver        PartyResponse           `json:"receiver"`
	OriginCity      string                  `json:"origin_city"`
	DestinationCity string                  `json:"destination_city"`
	NotifyURL       string                  `json:"notify_url"`
	CurrentLocation string                  `json:"current_location"`
	Events          []ShipmentEventResponse `json:"events"`
	DispatchedAt    *time.Time              `json:"dispatched_at"`
	PickedUpAt      *time.Time              `json:"picked_up_at"`
	DeliveredAt     *time.Time              `json:"delivered_at"`
	CreatedAt       time.Time               `json:"created_at"`
	UpdatedAt       time.Time               `json:"updated_at"`
}

type NotificationResponse struct {
	ID            int64      `json:"id"`
	EventID       string     `json:"event_id"`
	EventType     string     `json:"event_type"`
	TargetURL     string     `json:"target_url"`
	Status        string     `json:"status"`
	AttemptCount  int        `json:"attempt_count"`
	MaxAttempts   int        `json:"max_attempts"`
	NextAttemptAt time.Time  `json:"next_attempt_at"`
	LastAttemptAt *time.Time `json:"last_attempt_at"`
	DeliveredAt   *time.Time `json:"delivered_at"`
	LastError     *string    `json:"last_error"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type NotificationListResponse struct {
	Items []NotificationResponse `json:"items"`
}
