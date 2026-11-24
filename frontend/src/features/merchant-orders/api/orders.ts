import { apiClient } from "@/shared/api/client"
import { authorized } from "@/shared/api/authorized"
import type {
  MerchantOrderListQuery,
  MerchantOrderListResult,
} from "../types"

export async function listMerchantOrders(
  token: string,
  query: MerchantOrderListQuery,
) {
  const { data } = await apiClient.get<MerchantOrderListResult>(
    "/merchant/orders",
    { ...authorized(token), params: query },
  )
  return data
}
