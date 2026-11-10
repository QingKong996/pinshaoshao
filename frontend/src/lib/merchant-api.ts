import { requestJson } from "@/lib/api"

export type MerchantProductStatus = "on_shelf" | "off_shelf"

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

export type ShopProfileUpdate = {
  name: string
  description: string
}

export type Category = {
  id: number
  name: string
  product_count: number
  created_at: string
  updated_at: string
}

export type CategoryInput = {
  name: string
}

export type Product = {
  id: number
  name: string
  description: string
  status: MerchantProductStatus
  category_id: number | null
  category_name: string | null
  sku_count: number
  total_stock: number
  min_price_cents: number | null
  max_price_cents: number | null
  created_at: string
  updated_at: string
}

export type ProductInput = {
  name: string
  description: string
  category_id: number | null
  status: MerchantProductStatus
}

export type SKU = {
  id: number
  product_id: number
  specs: Record<string, string>
  price_cents: number
  stock: number
  is_enabled: boolean
  created_at: string
  updated_at: string
}

export type SKUInput = {
  specs: Record<string, string>
  price_cents: number
  stock: number
  is_enabled: boolean
}

export type SKUUpdate = Partial<SKUInput>

export type InventoryItem = {
  sku_id: number
  product_id: number
  product_name: string
  category_id: number | null
  category_name: string | null
  specs: Record<string, string>
  price_cents: number
  stock: number
  is_enabled: boolean
  updated_at: string
}

export type Dashboard = {
  product_count: number
  on_shelf_product_count: number
  sku_count: number
  category_count: number
  total_stock: number
  inventory_value_cents: number
  out_of_stock_sku_count: number
}

function authorized(token: string): RequestInit {
  return {
    headers: { Authorization: `Bearer ${token}` },
  }
}

export function getShopProfile(token: string) {
  return requestJson<ShopProfile>("/merchant/shop", authorized(token))
}

export function updateShopProfile(token: string, body: ShopProfileUpdate) {
  return requestJson<ShopProfile>("/merchant/shop", {
    ...authorized(token),
    method: "PATCH",
    body: JSON.stringify(body),
  })
}

export function listCategories(token: string) {
  return requestJson<Category[]>("/merchant/categories", authorized(token))
}

export function createCategory(token: string, body: CategoryInput) {
  return requestJson<Category>("/merchant/categories", {
    ...authorized(token),
    method: "POST",
    body: JSON.stringify(body),
  })
}

export function updateCategory(
  token: string,
  categoryID: number,
  body: CategoryInput,
) {
  return requestJson<Category>(`/merchant/categories/${categoryID}`, {
    ...authorized(token),
    method: "PATCH",
    body: JSON.stringify(body),
  })
}

export function deleteCategory(token: string, categoryID: number) {
  return requestJson<void>(`/merchant/categories/${categoryID}`, {
    ...authorized(token),
    method: "DELETE",
  })
}

export function listProducts(token: string) {
  return requestJson<Product[]>("/merchant/products", authorized(token))
}

export function createProduct(token: string, body: ProductInput) {
  return requestJson<Product>("/merchant/products", {
    ...authorized(token),
    method: "POST",
    body: JSON.stringify(body),
  })
}

export function getProduct(token: string, productID: number) {
  return requestJson<Product>(`/merchant/products/${productID}`, authorized(token))
}

export function updateProduct(
  token: string,
  productID: number,
  body: ProductInput,
) {
  return requestJson<Product>(`/merchant/products/${productID}`, {
    ...authorized(token),
    method: "PATCH",
    body: JSON.stringify(body),
  })
}

export function deleteProduct(token: string, productID: number) {
  return requestJson<void>(`/merchant/products/${productID}`, {
    ...authorized(token),
    method: "DELETE",
  })
}

export function listProductSKUs(token: string, productID: number) {
  return requestJson<SKU[]>(
    `/merchant/products/${productID}/skus`,
    authorized(token),
  )
}

export function createProductSKU(
  token: string,
  productID: number,
  body: SKUInput,
) {
  return requestJson<SKU>(`/merchant/products/${productID}/skus`, {
    ...authorized(token),
    method: "POST",
    body: JSON.stringify(body),
  })
}

export function updateProductSKU(
  token: string,
  productID: number,
  skuID: number,
  body: SKUUpdate,
) {
  return requestJson<SKU>(
    `/merchant/products/${productID}/skus/${skuID}`,
    {
      ...authorized(token),
      method: "PATCH",
      body: JSON.stringify(body),
    },
  )
}

export function deleteProductSKU(
  token: string,
  productID: number,
  skuID: number,
) {
  return requestJson<void>(`/merchant/products/${productID}/skus/${skuID}`, {
    ...authorized(token),
    method: "DELETE",
  })
}

export function getInventory(token: string) {
  return requestJson<InventoryItem[]>("/merchant/inventory", authorized(token))
}

export function getDashboard(token: string) {
  return requestJson<Dashboard>("/merchant/dashboard", authorized(token))
}
