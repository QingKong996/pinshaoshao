export type ShipmentStatus =
  | "pending_outbound"
  | "pending_pickup"
  | "picked_up"
  | "in_transit"
  | "at_transit_hub"
  | "out_for_delivery"
  | "delivered"

export type NotificationStatus =
  | "pending"
  | "delivering"
  | "succeeded"
  | "exhausted"

export type ShipmentParty = {
  name: string
  phone: string
  address: string
}

export type ShipmentEvent = {
  sequence: number
  status: ShipmentStatus
  event_type: string
  description: string
  location: string
  occurred_at: string
}

export type MockShipment = {
  shipment_no: string
  tracking_no: string
  merchant_order_no: string
  status: ShipmentStatus
  carrier: string
  item_summary: string
  sender: ShipmentParty
  receiver: ShipmentParty
  origin_city: string
  destination_city: string
  notify_url: string
  current_location: string
  events: ShipmentEvent[]
  dispatched_at: string | null
  picked_up_at: string | null
  delivered_at: string | null
  created_at: string
  updated_at: string
}

export type MockShipmentNotification = {
  id: number
  event_id: string
  event_type: string
  target_url: string
  status: NotificationStatus
  attempt_count: number
  max_attempts: number
  next_attempt_at: string
  last_attempt_at: string | null
  delivered_at: string | null
  last_error: string | null
  created_at: string
  updated_at: string
}

export type MockShipmentNotificationList = {
  items: MockShipmentNotification[]
}

export type CreateMockShipmentInput = {
  merchant_order_no: string
  notify_url: string
  item_summary?: string
  carrier?: string
  sender_name?: string
  sender_phone?: string
  sender_address?: string
  receiver_name?: string
  receiver_phone?: string
  receiver_address?: string
  origin_city?: string
  destination_city?: string
}
