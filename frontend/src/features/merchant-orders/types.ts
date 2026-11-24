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
  receiver_name: string
  receiver_phone: string
  shipping_address: string
  paid_at: string | null
  created_at: string
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
