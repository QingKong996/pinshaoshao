import { apiClient } from "@/shared/api/client"
import type {
  CreateMockShipmentInput,
  MockShipment,
  MockShipmentNotification,
  MockShipmentNotificationList,
} from "../dto/shipment"

const MOCK_LOGISTICS_API_BASE_URL = "/mocklogistics/api/v1"
const SHIPMENTS_PATH = "/shipments"

function shipmentPath(shipmentNo: string) {
  return `${SHIPMENTS_PATH}/${encodeURIComponent(shipmentNo)}`
}

export async function createMockShipment(input: CreateMockShipmentInput) {
  const { data } = await apiClient.post<MockShipment>(SHIPMENTS_PATH, input, {
    baseURL: MOCK_LOGISTICS_API_BASE_URL,
  })
  return data
}

export async function getMockShipment(shipmentNo: string) {
  const { data } = await apiClient.get<MockShipment>(
    shipmentPath(shipmentNo),
    { baseURL: MOCK_LOGISTICS_API_BASE_URL },
  )
  return data
}

export async function outboundMockShipment(shipmentNo: string) {
  const { data } = await apiClient.post<MockShipment>(
    `${shipmentPath(shipmentNo)}/outbound`,
    undefined,
    { baseURL: MOCK_LOGISTICS_API_BASE_URL },
  )
  return data
}

export async function advanceMockShipment(shipmentNo: string) {
  const { data } = await apiClient.post<MockShipment>(
    `${shipmentPath(shipmentNo)}/advance`,
    undefined,
    { baseURL: MOCK_LOGISTICS_API_BASE_URL },
  )
  return data
}

export async function completeMockShipment(shipmentNo: string) {
  const { data } = await apiClient.post<MockShipment>(
    `${shipmentPath(shipmentNo)}/complete`,
    undefined,
    { baseURL: MOCK_LOGISTICS_API_BASE_URL },
  )
  return data
}

export async function listMockShipmentNotifications(shipmentNo: string) {
  const { data } = await apiClient.get<MockShipmentNotificationList>(
    `${shipmentPath(shipmentNo)}/notifications`,
    { baseURL: MOCK_LOGISTICS_API_BASE_URL },
  )
  return data
}

export async function retryMockShipmentNotification(
  shipmentNo: string,
  notificationID: number,
) {
  const { data } = await apiClient.post<MockShipmentNotification>(
    `${shipmentPath(shipmentNo)}/notifications/${notificationID}/retry`,
    undefined,
    { baseURL: MOCK_LOGISTICS_API_BASE_URL },
  )
  return data
}
