import { authorized } from "@/shared/api/authorized"
import { apiClient } from "@/shared/api/client"
import type {
  CreateShipmentInput,
  OrderShipmentListResult,
  Shipment,
  ShipmentListQuery,
  ShipmentListResult,
} from "../types"

const SHIPMENTS_PATH = "/shipments"
const MERCHANT_SHIPMENTS_PATH = "/merchant/shipments"

function shipmentPath(basePath: string, shipmentNo: string) {
  return `${basePath}/${encodeURIComponent(shipmentNo)}`
}

export async function listOrderShipments(
  token: string,
  orderID: number,
) {
  const { data } = await apiClient.get<OrderShipmentListResult>(
    SHIPMENTS_PATH,
    {
      ...authorized(token),
      params: { order_id: orderID },
    },
  )
  return data
}

export async function getBuyerShipment(
  token: string,
  shipmentNo: string,
) {
  const { data } = await apiClient.get<Shipment>(
    shipmentPath(SHIPMENTS_PATH, shipmentNo),
    authorized(token),
  )
  return data
}

export async function createShipment(
  token: string,
  input: CreateShipmentInput,
) {
  const { data } = await apiClient.post<Shipment>(
    MERCHANT_SHIPMENTS_PATH,
    input,
    authorized(token),
  )
  return data
}

export async function listMerchantShipments(
  token: string,
  query: ShipmentListQuery,
) {
  const { data } = await apiClient.get<ShipmentListResult>(
    MERCHANT_SHIPMENTS_PATH,
    { ...authorized(token), params: query },
  )
  return data
}

export async function getMerchantShipment(
  token: string,
  shipmentNo: string,
) {
  const { data } = await apiClient.get<Shipment>(
    shipmentPath(MERCHANT_SHIPMENTS_PATH, shipmentNo),
    authorized(token),
  )
  return data
}
