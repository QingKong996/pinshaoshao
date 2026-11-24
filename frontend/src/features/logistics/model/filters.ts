import type { ShipmentListQuery, ShipmentStatus } from "../types"

export const ALL = "all"
export type ShipmentStatusFilter = typeof ALL | ShipmentStatus

export const SHIPMENT_STATUSES: ShipmentStatus[] = [
  "pending_pickup",
  "picked_up",
  "in_transit",
  "at_transit_hub",
  "out_for_delivery",
  "delivered",
]

export function buildShipmentListQuery({
  page,
  pageSize,
  statusFilter,
}: {
  page: number
  pageSize: number
  statusFilter: ShipmentStatusFilter
}): ShipmentListQuery {
  const query: ShipmentListQuery = {
    page,
    page_size: pageSize,
  }
  if (statusFilter !== ALL) {
    query.status = statusFilter
  }
  return query
}
