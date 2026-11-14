import { apiClient } from "@/shared/api/client"
import { authorized } from "@/shared/api/authorized"
import type { ProductCategory, ProductCategoryInput } from "../types"

export async function listProductCategories(token: string) {
  const { data } = await apiClient.get<ProductCategory[]>(
    "/merchant/categories",
    authorized(token),
  )
  return data
}

export async function createProductCategory(
  token: string,
  input: ProductCategoryInput,
) {
  const { data } = await apiClient.post<ProductCategory>(
    "/merchant/categories",
    input,
    authorized(token),
  )
  return data
}

export async function updateProductCategory(
  token: string,
  categoryID: number,
  input: ProductCategoryInput,
) {
  const { data } = await apiClient.patch<ProductCategory>(
    `/merchant/categories/${categoryID}`,
    input,
    authorized(token),
  )
  return data
}

export async function deleteProductCategory(
  token: string,
  categoryID: number,
) {
  await apiClient.delete(
    `/merchant/categories/${categoryID}`,
    authorized(token),
  )
}
