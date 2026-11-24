import type {
  MerchantOrderListQuery,
  MerchantOrderStatus,
} from "../types"

export const ALL = "all"
export type StatusFilter = typeof ALL | MerchantOrderStatus

export const MERCHANT_ORDER_STATUSES: MerchantOrderStatus[] = [
  "pending_payment",
  "awaiting_shipment",
  "shipped",
  "completed",
  "cancelled",
]

export function buildMerchantOrderListQuery({
  page,
  pageSize,
  statusFilter,
}: {
  page: number
  pageSize: number
  statusFilter: StatusFilter
}): MerchantOrderListQuery {
  const query: MerchantOrderListQuery = {
    page,
    page_size: pageSize,
  }
  if (statusFilter !== ALL) {
    query.status = statusFilter
  }
  return query
}
