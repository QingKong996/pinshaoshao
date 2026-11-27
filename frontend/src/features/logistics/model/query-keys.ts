import type { ShipmentListQuery } from "../types"

export const shipmentQueryKeys = {
  all: () => ["logistics"] as const,
  orderShipments: (token: string, orderID: number) =>
    [...shipmentQueryKeys.all(), "order", token, orderID] as const,
  buyerShipment: (token: string, shipmentNo: string) =>
    [...shipmentQueryKeys.all(), "buyer", token, shipmentNo] as const,
  merchantShipments: () =>
    [...shipmentQueryKeys.all(), "merchant"] as const,
  merchantShipmentList: (query: ShipmentListQuery) =>
    [...shipmentQueryKeys.merchantShipments(), "list", query] as const,
  merchantShipment: (shipmentNo: string) =>
    [...shipmentQueryKeys.merchantShipments(), shipmentNo] as const,
}
