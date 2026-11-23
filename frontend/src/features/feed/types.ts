export type FeedItem = {
  id: number
  shop_id: number
  shop_name: string
  name: string
  description: string
  category_name: string | null
  min_price_cents: number
  max_price_cents: number
  available_stock: number
}

export type FeedListResponse = {
  items: FeedItem[]
  page: number
  page_size: number
  total: number
  total_pages: number
}

export type ProductSKU = {
  id: number
  specs: Record<string, string>
  price_cents: number
  is_enabled: boolean
  available_stock: number
}

export type ProductDetail = {
  id: number
  shop_id: number
  shop_name: string
  name: string
  description: string
  category_name: string | null
  skus: ProductSKU[]
}
