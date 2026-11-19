import { apiClient } from "@/shared/api/client"
import { authorized } from "@/shared/api/authorized"
import type { InventoryItem } from "../types"

export async function listInventory(token: string) {
  const { data } = await apiClient.get<InventoryItem[]>(
    "/merchant/inventory",
    authorized(token),
  )
  return data
}

export async function updateInventoryTotalStock(
  token: string,
  skuID: number,
  totalStock: number,
  version: number,
) {
  const { data } = await apiClient.patch<InventoryItem>(
    `/merchant/inventory/${skuID}`,
    { total_stock: totalStock, version },
    authorized(token),
  )
  return data
}
