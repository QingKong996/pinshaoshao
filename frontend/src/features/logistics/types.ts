export type ShipmentStatus =
  | "pending_pickup"
  | "picked_up"
  | "in_transit"
  | "at_transit_hub"
  | "out_for_delivery"
  | "delivered"

export type ShipmentEvent = {
  sequence: number
  status: ShipmentStatus
  event_type: string
  description: string
  location: string
  occurred_at: string
}

export type ShipmentParty = {
  name: string
  phone: string
  address: string
}

export type ShipmentSummary = {
  shipment_no: string
  sub_order_id: number
  order_id: number
  status: ShipmentStatus
  carrier: string
  tracking_no: string
  item_summary: string
  receiver_name: string
  receiver_phone: string
  current_location: string
  shipped_at: string
  delivered_at: string | null
  created_at: string
  updated_at: string
}

export type Shipment = {
  shipment_no: string
  sub_order_id: number
  order_id: number
  status: ShipmentStatus
  carrier: string
  tracking_no: string
  item_summary: string
  receiver: ShipmentParty
  origin_city: string
  destination_city: string
  current_location: string
  events: ShipmentEvent[]
  shipped_at: string
  delivered_at: string | null
  created_at: string
  updated_at: string
}

export type ShipmentListQuery = {
  page: number
  page_size: number
  status?: ShipmentStatus
}

export type ShipmentListResult = {
  items: ShipmentSummary[]
  page: number
  page_size: number
  total: number
  total_pages: number
}

export type OrderShipmentListResult = {
  items: ShipmentSummary[]
}

export type CreateShipmentInput = {
  sub_order_id: number
  carrier?: string
}
