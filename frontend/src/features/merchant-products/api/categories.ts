import { apiClient } from "@/shared/api/client"
import { authorized } from "@/shared/api/authorized"
import type { ProductCategory } from "../types"

export async function listProductCategories(token: string) {
  const { data } = await apiClient.get<ProductCategory[]>(
    "/merchant/categories",
    authorized(token),
  )
  return data
}
