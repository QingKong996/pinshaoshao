export type CartItemAvailability =
  | "available"
  | "product_deleted"
  | "product_off_shelf"
  | "sku_deleted"
  | "sku_disabled"
  | "out_of_stock"
  | "insufficient_stock"

export type CartItem = {
  id: number
  product_id: number
  sku_id: number
  product_name: string
  sku_specs: Record<string, string>
  unit_price_cents: number
  quantity: number
  line_total_cents: number
  available_stock: number | null
  availability: CartItemAvailability
  created_at: string
  updated_at: string
}

export type AddCartItemInput = {
  sku_id: number
  quantity: number
}
