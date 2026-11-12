export type ProductStatus = "on_shelf" | "off_shelf"

export type Product = {
  id: number
  name: string
  description: string
  status: ProductStatus
  category_id: number | null
  category_name: string | null
  sku_count: number
  total_stock: number
  min_price_cents: number | null
  max_price_cents: number | null
  created_at: string
  updated_at: string
}

export type ProductCategory = {
  id: number
  name: string
  product_count: number
  created_at: string
  updated_at: string
}

export type ProductSKU = {
  id: number
  product_id: number
  specs: Record<string, string>
  price_cents: number
  stock: number
  is_enabled: boolean
  created_at: string
  updated_at: string
}

export type ProductInput = {
  name: string
  description: string
  category_id: number | null
  status: ProductStatus
}

export type ProductSKUInput = {
  specs: Record<string, string>
  price_cents: number
  stock: number
  is_enabled: boolean
}

export type CatalogChanged = () => Promise<unknown>
