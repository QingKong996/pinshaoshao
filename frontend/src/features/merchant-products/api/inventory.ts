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
