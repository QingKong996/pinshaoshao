import { useQuery } from "@tanstack/react-query"
import {
  getMockShipment,
  listMockShipmentNotifications,
} from "../api/shipments"
import { mockShipmentQueryKeys } from "./query-keys"

const NOTIFICATION_POLL_INTERVAL_MS = 2_000

export function useMockShipment(shipmentNo: string) {
  return useQuery({
    queryKey: mockShipmentQueryKeys.shipment(shipmentNo),
    queryFn: () => getMockShipment(shipmentNo),
    enabled: shipmentNo.length > 0,
  })
}

export function useMockShipmentNotifications(shipmentNo: string) {
  return useQuery({
    queryKey: mockShipmentQueryKeys.notifications(shipmentNo),
    queryFn: () => listMockShipmentNotifications(shipmentNo),
    enabled: shipmentNo.length > 0,
    refetchInterval: (query) => {
      const items = query.state.data?.items
      if (!items) {
        return false
      }
      return items.some(
        (item) => item.status === "pending" || item.status === "delivering",
      )
        ? NOTIFICATION_POLL_INTERVAL_MS
        : false
    },
  })
}
