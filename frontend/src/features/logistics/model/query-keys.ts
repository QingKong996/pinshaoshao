import type { ShipmentListQuery } from "../types"

export const shipmentQueryKeys = {
  all: () => ["logistics"] as const,
  orderShipments: (orderID: number) =>
    [...shipmentQueryKeys.all(), "order", orderID] as const,
  buyerShipment: (shipmentNo: string) =>
    [...shipmentQueryKeys.all(), "buyer", shipmentNo] as const,
  merchantShipments: () =>
    [...shipmentQueryKeys.all(), "merchant"] as const,
  merchantShipmentList: (query: ShipmentListQuery) =>
    [...shipmentQueryKeys.merchantShipments(), "list", query] as const,
  merchantShipment: (shipmentNo: string) =>
    [...shipmentQueryKeys.merchantShipments(), shipmentNo] as const,
}
