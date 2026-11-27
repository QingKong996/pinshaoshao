import { apiClient } from "@/shared/api/client"
import { authorized } from "@/shared/api/authorized"
import type {
  MerchantOrderDetail,
  MerchantOrderListQuery,
  MerchantOrderListResult,
  MerchantOrderReceiver,
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

export async function getMerchantOrder(
  token: string,
  subOrderID: number,
) {
  const { data } = await apiClient.get<MerchantOrderDetail>(
    `/merchant/orders/${subOrderID}`,
    authorized(token),
  )
  return data
}

export async function getMerchantOrderReceiver(
  token: string,
  subOrderID: number,
) {
  const { data } = await apiClient.get<MerchantOrderReceiver>(
    `/merchant/orders/${subOrderID}/receiver`,
    authorized(token),
  )
  return data
}
