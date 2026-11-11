import { requestJson } from "@/lib/api"

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

function authorized(token: string): RequestInit {
  return {
    headers: { Authorization: `Bearer ${token}` },
  }
}

function authorizedJson(
  token: string,
  method: "POST" | "PATCH" | "DELETE",
  body?: unknown,
): RequestInit {
  const init: RequestInit = { ...authorized(token), method }
  if (body !== undefined) {
    init.body = JSON.stringify(body)
  }
  return init
}

export function getDashboard(token: string) {
  return requestJson<Dashboard>("/merchant/dashboard", authorized(token))
}

export function getShopProfile(token: string) {
  return requestJson<ShopProfile>("/merchant/shop", authorized(token))
}

export function listProducts(token: string) {
  return requestJson<Product[]>("/merchant/products", authorized(token))
}

export function getProduct(token: string, productID: number) {
  return requestJson<Product>(
    `/merchant/products/${productID}`,
    authorized(token),
  )
}

export function createProduct(token: string, input: ProductInput) {
  return requestJson<Product>(
    "/merchant/products",
    authorizedJson(token, "POST", input),
  )
}

export function updateProduct(
  token: string,
  productID: number,
  input: ProductInput,
) {
  return requestJson<Product>(
    `/merchant/products/${productID}`,
    authorizedJson(token, "PATCH", input),
  )
}

export function deleteProduct(token: string, productID: number) {
  return requestJson<void>(
    `/merchant/products/${productID}`,
    authorizedJson(token, "DELETE"),
  )
}

export function listProductCategories(token: string) {
  return requestJson<ProductCategory[]>("/merchant/categories", authorized(token))
}

export function listProductSKUs(token: string, productID: number) {
  return requestJson<ProductSKU[]>(
    `/merchant/products/${productID}/skus`,
    authorized(token),
  )
}

export function createProductSKU(
  token: string,
  productID: number,
  input: ProductSKUInput,
) {
  return requestJson<ProductSKU>(
    `/merchant/products/${productID}/skus`,
    authorizedJson(token, "POST", input),
  )
}

export function updateProductSKU(
  token: string,
  productID: number,
  skuID: number,
  input: ProductSKUInput,
) {
  return requestJson<ProductSKU>(
    `/merchant/products/${productID}/skus/${skuID}`,
    authorizedJson(token, "PATCH", input),
  )
}

export function deleteProductSKU(
  token: string,
  productID: number,
  skuID: number,
) {
  return requestJson<void>(
    `/merchant/products/${productID}/skus/${skuID}`,
    authorizedJson(token, "DELETE"),
  )
}
