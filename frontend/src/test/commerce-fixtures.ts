import type {
  Order,
  OrderListResponse,
} from "@/features/order/types"
import type {
  Shipment,
  ShipmentSummary,
} from "@/features/logistics/types"

export const shipmentNo = "4f2614df-f674-4b5e-a9e4-3fc5e7f62013"
export const orderID = 12

export const shipmentSummary: ShipmentSummary = {
  shipment_no: shipmentNo,
  sub_order_id: 4,
  order_id: orderID,
  status: "in_transit",
  carrier: "SF Express",
  tracking_no: "SF123456789",
  item_summary: "Keyboard x1",
  receiver_name: "Receiver",
  receiver_phone: "1234567890",
  current_location: "Shanghai",
  shipped_at: "2026-11-24T08:00:00Z",
  delivered_at: null,
  created_at: "2026-11-24T08:00:00Z",
  updated_at: "2026-11-24T09:00:00Z",
}

export const shipment: Shipment = {
  ...shipmentSummary,
  receiver: {
    name: "Receiver",
    phone: "1234567890",
    address: "Shanghai",
  },
  origin_city: "Shanghai",
  destination_city: "Hangzhou",
  events: [
    {
      sequence: 2,
      status: "in_transit",
      event_type: "shipment.in_transit",
      description: "Departed Shanghai",
      location: "Shanghai",
      occurred_at: "2026-11-24T09:00:00Z",
    },
    {
      sequence: 1,
      status: "picked_up",
      event_type: "shipment.picked_up",
      description: "Picked up",
      location: "Shanghai",
      occurred_at: "2026-11-24T08:00:00Z",
    },
  ],
}

export const order: Order = {
  id: orderID,
  status: "pending_payment",
  total_amount_cents: 3998,
  receiver_name: "Receiver",
  receiver_phone: "1234567890",
  shipping_address: "Shanghai",
  expires_at: "2026-11-24T12:30:00Z",
  paid_at: null,
  created_at: "2026-11-24T12:00:00Z",
  updated_at: "2026-11-24T12:00:00Z",
  sub_orders: [
    {
      id: 4,
      shop_id: 3,
      shop_name: "E2E Shop",
      status: "pending_payment",
      total_amount_cents: 3998,
      items: [
        {
          id: 21,
          product_id: 10,
          sku_id: 5,
          product_name: "Cotton T-shirt",
          sku_specs: { Size: "M" },
          unit_price_cents: 1999,
          quantity: 2,
          line_amount_cents: 3998,
        },
      ],
    },
  ],
}

export const orderList: OrderListResponse = {
  items: [
    {
      id: orderID,
      status: "pending_payment",
      total_amount_cents: 3998,
      item_count: 2,
      shop_count: 1,
      expires_at: "2026-11-24T12:30:00Z",
      created_at: "2026-11-24T12:00:00Z",
    },
  ],
  page: 1,
  page_size: 10,
  total: 1,
  total_pages: 1,
}
