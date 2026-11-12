import { apiClient } from "@/shared/api/client"
import { authorized } from "@/shared/api/authorized"
import type { Dashboard } from "../types"

export async function getDashboard(token: string) {
  const { data } = await apiClient.get<Dashboard>(
    "/merchant/dashboard",
    authorized(token),
  )
  return data
}
