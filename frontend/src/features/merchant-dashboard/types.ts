export type Dashboard = {
  product_count: number
  on_shelf_product_count: number
  sku_count: number
  category_count: number
  total_stock: number
  inventory_value_cents: number
  out_of_stock_sku_count: number
}

export type ShopProfile = {
  id: number
  username: string
  name: string
  description: string
  email: string
  status: "active" | "banned"
  created_at: string
  updated_at: string
}
