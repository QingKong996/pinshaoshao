import { apiClient } from "@/api/client"

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

function authorized(token: string) {
  return {
    headers: { Authorization: `Bearer ${token}` },
  }
}

export async function getDashboard(token: string) {
  const { data } = await apiClient.get<Dashboard>(
    "/merchant/dashboard",
    authorized(token),
  )
  return data
}

export async function getShopProfile(token: string) {
  const { data } = await apiClient.get<ShopProfile>(
    "/merchant/shop",
    authorized(token),
  )
  return data
}

export async function listProducts(token: string) {
  const { data } = await apiClient.get<Product[]>(
    "/merchant/products",
    authorized(token),
  )
  return data
}

export async function getProduct(token: string, productID: number) {
  const { data } = await apiClient.get<Product>(
    `/merchant/products/${productID}`,
    authorized(token),
  )
  return data
}

export async function createProduct(token: string, input: ProductInput) {
  const { data } = await apiClient.post<Product>(
    "/merchant/products",
    input,
    authorized(token),
  )
  return data
}

export async function updateProduct(
  token: string,
  productID: number,
  input: ProductInput,
) {
  const { data } = await apiClient.patch<Product>(
    `/merchant/products/${productID}`,
    input,
    authorized(token),
  )
  return data
}

export async function deleteProduct(token: string, productID: number) {
  await apiClient.delete(
    `/merchant/products/${productID}`,
    authorized(token),
  )
}

export async function listProductCategories(token: string) {
  const { data } = await apiClient.get<ProductCategory[]>(
    "/merchant/categories",
    authorized(token),
  )
  return data
}

export async function listProductSKUs(token: string, productID: number) {
  const { data } = await apiClient.get<ProductSKU[]>(
    `/merchant/products/${productID}/skus`,
    authorized(token),
  )
  return data
}

export async function createProductSKU(
  token: string,
  productID: number,
  input: ProductSKUInput,
) {
  const { data } = await apiClient.post<ProductSKU>(
    `/merchant/products/${productID}/skus`,
    input,
    authorized(token),
  )
  return data
}

export async function updateProductSKU(
  token: string,
  productID: number,
  skuID: number,
  input: ProductSKUInput,
) {
  const { data } = await apiClient.patch<ProductSKU>(
    `/merchant/products/${productID}/skus/${skuID}`,
    input,
    authorized(token),
  )
  return data
}

export async function deleteProductSKU(
  token: string,
  productID: number,
  skuID: number,
) {
  await apiClient.delete(
    `/merchant/products/${productID}/skus/${skuID}`,
    authorized(token),
  )
}
