export const mockShipmentQueryKeys = {
  all: ["mock-shipments"] as const,
  shipment: (shipmentNo: string) =>
    [...mockShipmentQueryKeys.all, "shipment", shipmentNo] as const,
  notifications: (shipmentNo: string) =>
    [...mockShipmentQueryKeys.all, "notifications", shipmentNo] as const,
}
