import { apiClient } from "@/shared/api/client"
import { authorized } from "@/shared/api/authorized"
import type { ProductSKU, ProductSKUInput } from "../types"

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
