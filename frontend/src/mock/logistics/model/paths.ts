export const MOCK_LOGISTICS_ROUTE_PATHS = {
  signIn: "/mocklogistics/signin",
  list: "/mocklogistics",
  detail: "/mocklogistics/shipments/:shipmentNo",
} as const

export function mockLogisticsDetailPath(shipmentNo: string) {
  return `/mocklogistics/shipments/${encodeURIComponent(shipmentNo)}`
}
