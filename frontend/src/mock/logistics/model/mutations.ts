import { useMutation, useQueryClient } from "@tanstack/react-query"
import {
  advanceMockShipment,
  completeMockShipment,
  createMockShipment,
  outboundMockShipment,
  retryMockShipmentNotification,
} from "../api/shipments"
import type {
  MockShipment,
  MockShipmentNotificationList,
} from "../dto/shipment"
import { mockShipmentQueryKeys } from "./query-keys"

function useShipmentAction(
  shipmentNo: string,
  action: (shipmentNo: string) => Promise<MockShipment>,
) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => action(shipmentNo),
    onSuccess: (shipment) => {
      queryClient.setQueryData(
        mockShipmentQueryKeys.shipment(shipmentNo),
        shipment,
      )
      void queryClient.invalidateQueries({
        queryKey: mockShipmentQueryKeys.notifications(shipmentNo),
      })
    },
  })
}

export function useCreateMockShipment() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: createMockShipment,
    onSuccess: (shipment) => {
      queryClient.setQueryData(
        mockShipmentQueryKeys.shipment(shipment.shipment_no),
        shipment,
      )
    },
  })
}

export function useOutboundMockShipment(shipmentNo: string) {
  return useShipmentAction(shipmentNo, outboundMockShipment)
}

export function useAdvanceMockShipment(shipmentNo: string) {
  return useShipmentAction(shipmentNo, advanceMockShipment)
}

export function useCompleteMockShipment(shipmentNo: string) {
  return useShipmentAction(shipmentNo, completeMockShipment)
}

export function useRetryMockShipmentNotification(shipmentNo: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (notificationID: number) =>
      retryMockShipmentNotification(shipmentNo, notificationID),
    onSuccess: (notification) => {
      queryClient.setQueryData<MockShipmentNotificationList>(
        mockShipmentQueryKeys.notifications(shipmentNo),
        (current) => {
          if (!current) {
            return current
          }
          return {
            items: current.items.map((item) =>
              item.id === notification.id ? notification : item,
            ),
          }
        },
      )
      void queryClient.invalidateQueries({
        queryKey: mockShipmentQueryKeys.notifications(shipmentNo),
      })
    },
  })
}
