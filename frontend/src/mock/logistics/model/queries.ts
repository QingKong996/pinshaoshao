import { useQuery } from "@tanstack/react-query"
import {
  getMockShipment,
  listMockShipmentNotifications,
} from "../api/shipments"
import { mockShipmentQueryKeys } from "./query-keys"

const SHIPMENT_POLL_INTERVAL_MS = 2_000
const NOTIFICATION_POLL_INTERVAL_MS = 2_000

export function useMockShipment(shipmentNo: string) {
  return useQuery({
    queryKey: mockShipmentQueryKeys.shipment(shipmentNo),
    queryFn: () => getMockShipment(shipmentNo),
    enabled: shipmentNo.length > 0,
    refetchInterval: (query) => {
      const status = query.state.data?.status
      if (!status || status === "pending_outbound" || status === "delivered") {
        return false
      }
      return SHIPMENT_POLL_INTERVAL_MS
    },
  })
}

export function useMockShipmentNotifications(
  shipmentNo: string,
  isActive: boolean,
) {
  return useQuery({
    queryKey: mockShipmentQueryKeys.notifications(shipmentNo),
    queryFn: () => listMockShipmentNotifications(shipmentNo),
    enabled: shipmentNo.length > 0,
    refetchInterval: (query) => {
      if (isActive) {
        return NOTIFICATION_POLL_INTERVAL_MS
      }
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
