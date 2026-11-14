import { apiClient } from "@/shared/api/client"
import { authorized } from "@/shared/api/authorized"
import type {
  Product,
  ProductInput,
  ProductListQuery,
  ProductListResult,
} from "../types"

export async function listProducts(token: string, query: ProductListQuery) {
  const { data } = await apiClient.get<ProductListResult>(
    "/merchant/products",
    { ...authorized(token), params: query },
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
