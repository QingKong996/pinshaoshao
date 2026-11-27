import { keepPreviousData, useQuery } from "@tanstack/react-query"
import {
  getBuyerShipment,
  getMerchantShipment,
  listMerchantShipments,
  listOrderShipments,
} from "../api/shipments"
import type { ShipmentListQuery } from "../types"
import { shipmentQueryKeys } from "./query-keys"

const TRACKING_REFRESH_INTERVAL = 30_000

export function useOrderShipments(
  token: string,
  orderID: number,
  enabled = true,
) {
  return useQuery({
    queryKey: shipmentQueryKeys.orderShipments(token, orderID),
    queryFn: () => listOrderShipments(token, orderID),
    enabled: enabled && Number.isInteger(orderID) && orderID > 0,
    refetchInterval: TRACKING_REFRESH_INTERVAL,
  })
}

export function useBuyerShipment(
  token: string,
  shipmentNo: string,
  enabled = true,
) {
  return useQuery({
    queryKey: shipmentQueryKeys.buyerShipment(token, shipmentNo),
    queryFn: () => getBuyerShipment(token, shipmentNo),
    enabled: enabled && shipmentNo !== "",
    refetchInterval: TRACKING_REFRESH_INTERVAL,
  })
}

export function useMerchantShipments(
  token: string,
  query: ShipmentListQuery,
) {
  return useQuery({
    queryKey: shipmentQueryKeys.merchantShipmentList(query),
    queryFn: () => listMerchantShipments(token, query),
    placeholderData: keepPreviousData,
    refetchInterval: TRACKING_REFRESH_INTERVAL,
  })
}

export function useMerchantShipment(
  token: string,
  shipmentNo: string,
  enabled = true,
) {
  return useQuery({
    queryKey: shipmentQueryKeys.merchantShipment(shipmentNo),
    queryFn: () => getMerchantShipment(token, shipmentNo),
    enabled: enabled && shipmentNo !== "",
    refetchInterval: TRACKING_REFRESH_INTERVAL,
  })
}
