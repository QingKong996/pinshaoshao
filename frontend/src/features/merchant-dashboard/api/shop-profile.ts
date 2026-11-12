import { apiClient } from "@/shared/api/client"
import { authorized } from "@/shared/api/authorized"
import type { ShopProfile } from "../types"

export async function getShopProfile(token: string) {
  const { data } = await apiClient.get<ShopProfile>(
    "/merchant/shop",
    authorized(token),
  )
  return data
}
