import type {
  CreateMockShipmentInput,
  MockShipment,
  MockShipmentNotification,
  MockShipmentNotificationList,
} from "../dto/shipment"
import { mockLogisticsApiClient } from "./client"

const SHIPMENTS_PATH = "/shipments"

function shipmentPath(shipmentNo: string) {
  return `${SHIPMENTS_PATH}/${encodeURIComponent(shipmentNo)}`
}

export async function createMockShipment(input: CreateMockShipmentInput) {
  const { data } = await mockLogisticsApiClient.post<MockShipment>(
    SHIPMENTS_PATH,
    input,
  )
  return data
}

export async function getMockShipment(shipmentNo: string) {
  const { data } = await mockLogisticsApiClient.get<MockShipment>(
    shipmentPath(shipmentNo),
  )
  return data
}

export async function outboundMockShipment(shipmentNo: string) {
  const { data } = await mockLogisticsApiClient.post<MockShipment>(
    `${shipmentPath(shipmentNo)}/outbound`,
    undefined,
  )
  return data
}

export async function advanceMockShipment(shipmentNo: string) {
  const { data } = await mockLogisticsApiClient.post<MockShipment>(
    `${shipmentPath(shipmentNo)}/advance`,
    undefined,
  )
  return data
}

export async function completeMockShipment(shipmentNo: string) {
  const { data } = await mockLogisticsApiClient.post<MockShipment>(
    `${shipmentPath(shipmentNo)}/complete`,
    undefined,
  )
  return data
}

export async function listMockShipmentNotifications(shipmentNo: string) {
  const { data } = await mockLogisticsApiClient.get<MockShipmentNotificationList>(
    `${shipmentPath(shipmentNo)}/notifications`,
  )
  return data
}

export async function retryMockShipmentNotification(
  shipmentNo: string,
  notificationID: number,
) {
  const { data } = await mockLogisticsApiClient.post<MockShipmentNotification>(
    `${shipmentPath(shipmentNo)}/notifications/${notificationID}/retry`,
    undefined,
  )
  return data
}
