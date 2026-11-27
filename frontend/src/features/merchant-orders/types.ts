export type MerchantOrderStatus =
  | "pending_payment"
  | "awaiting_shipment"
  | "shipped"
  | "completed"
  | "cancelled"

export type MerchantOrder = {
  id: number
  order_id: number
  status: MerchantOrderStatus
  total_amount_cents: number
  item_count: number
  paid_at: string | null
  created_at: string
}

export type MerchantOrderItem = {
  id: number
  product_id: number
  sku_id: number
  product_name: string
  sku_specs: Record<string, string>
  unit_price_cents: number
  quantity: number
  line_amount_cents: number
}

export type MerchantOrderDetail = MerchantOrder & {
  updated_at: string
  items: MerchantOrderItem[]
}

export type MerchantOrderReceiver = {
  receiver_name: string
  receiver_phone: string
  shipping_address: string
}

export type MerchantOrderListQuery = {
  page: number
  page_size: number
  status?: MerchantOrderStatus
}

export type MerchantOrderListResult = {
  items: MerchantOrder[]
  page: number
  page_size: number
  total: number
  total_pages: number
}
