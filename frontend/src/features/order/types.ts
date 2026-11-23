export type OrderStatus =
  | "pending_payment"
  | "paid"
  | "partially_shipped"
  | "shipped"
  | "completed"
  | "cancelled"

export type SubOrderStatus =
  | "pending_payment"
  | "awaiting_shipment"
  | "shipped"
  | "completed"
  | "cancelled"

export type OrderItem = {
  id: number
  product_id: number
  sku_id: number
  product_name: string
  sku_specs: Record<string, string>
  unit_price_cents: number
  quantity: number
  line_amount_cents: number
}

export type SubOrder = {
  id: number
  shop_id: number
  shop_name: string
  status: SubOrderStatus
  total_amount_cents: number
  items: OrderItem[]
}

export type Order = {
  id: number
  status: OrderStatus
  total_amount_cents: number
  receiver_name: string
  receiver_phone: string
  shipping_address: string
  expires_at: string
  paid_at: string | null
  created_at: string
  updated_at: string
  sub_orders: SubOrder[]
}

export type OrderSummary = {
  id: number
  status: OrderStatus
  total_amount_cents: number
  item_count: number
  shop_count: number
  expires_at: string
  created_at: string
}

export type OrderListResponse = {
  items: OrderSummary[]
  page: number
  page_size: number
  total: number
  total_pages: number
}

export type CreateOrderItemInput = {
  cart_item_id: number
  quantity: number
}

export type CreateOrderInput = {
  checkout_key: string
  items: CreateOrderItemInput[]
  receiver_name: string
  receiver_phone: string
  shipping_address: string
}

export type CreateDirectOrderInput = {
  checkout_key: string
  sku_id: number
  quantity: number
  receiver_name: string
  receiver_phone: string
  shipping_address: string
}

export type CheckoutLine = {
  id: number
  product_name: string
  sku_specs: Record<string, string>
  unit_price_cents: number
  quantity: number
  line_total_cents: number
}
