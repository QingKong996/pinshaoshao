export const MOCK_LOGISTICS_ROUTE_PATHS = {
  list: "/mocklogistics",
  detail: "/mocklogistics/shipments/:shipmentNo",
} as const

export function mockLogisticsDetailPath(shipmentNo: string) {
  return `/mocklogistics/shipments/${encodeURIComponent(shipmentNo)}`
}
