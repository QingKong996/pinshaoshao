import type { ReactNode } from "react"
import { QueryClientProvider } from "@tanstack/react-query"
import { renderHook, waitFor } from "@testing-library/react"
import { HttpResponse, http } from "msw"
import { describe, expect, it } from "vitest"
import type {
  OrderShipmentListResult,
  Shipment,
} from "@/features/logistics/types"
import { createTestQueryClient } from "@/test/render"
import { server } from "@/test/server"
import { useCreateShipment } from "./mutations"
import { useOrderShipments } from "./queries"
import { shipmentQueryKeys } from "./query-keys"

const shipmentSummary = {
  shipment_no: "4f2614df-f674-4b5e-a9e4-3fc5e7f62013",
  sub_order_id: 4,
  order_id: 12,
  status: "in_transit" as const,
  carrier: "SF Express",
  tracking_no: "SF123456789",
  item_summary: "Keyboard x1",
  receiver_name: "Receiver",
  receiver_phone: "1234567890",
  current_location: "Shanghai",
  shipped_at: "2026-11-24T08:00:00Z",
  delivered_at: null,
  created_at: "2026-11-24T08:00:00Z",
  updated_at: "2026-11-24T09:00:00Z",
}

function wrapper({ children }: { children: ReactNode }) {
  return (
    <QueryClientProvider client={createTestQueryClient()}>
      {children}
    </QueryClientProvider>
  )
}

describe("logistics queries", () => {
  it("separates buyer logistics cache entries by session token", () => {
    expect(
      shipmentQueryKeys.buyerShipment("buyer-a", "shipment-1"),
    ).not.toEqual(shipmentQueryKeys.buyerShipment("buyer-b", "shipment-1"))
    expect(
      shipmentQueryKeys.orderShipments("buyer-a", 12),
    ).not.toEqual(shipmentQueryKeys.orderShipments("buyer-b", 12))
  })

  it("loads buyer shipments for an order", async () => {
    let orderID: string | null = null
    let authorization: string | null = null
    const response: OrderShipmentListResult = {
      items: [shipmentSummary],
    }
    server.use(
      http.get("*/api/v1/shipments", ({ request }) => {
        const url = new URL(request.url)
        orderID = url.searchParams.get("order_id")
        authorization = request.headers.get("Authorization")
        return HttpResponse.json(response)
      }),
    )

    const query = renderHook(
      () => useOrderShipments("buyer-token", 12),
      { wrapper },
    )

    await waitFor(() => {
      expect(query.result.current.isSuccess).toBe(true)
    })
    expect(orderID).toBe("12")
    expect(authorization).toBe("Bearer buyer-token")
    expect(query.result.current.data).toEqual(response)
  })

  it("creates a merchant shipment", async () => {
    let requestBody: unknown = null
    const created: Shipment = {
      ...shipmentSummary,
      receiver: {
        name: "Receiver",
        phone: "1234567890",
        address: "Shanghai",
      },
      origin_city: "Shanghai",
      destination_city: "Hangzhou",
      events: [],
    }
    server.use(
      http.post("*/api/v1/merchant/shipments", async ({ request }) => {
        requestBody = await request.json()
        return HttpResponse.json(created, { status: 201 })
      }),
    )

    const mutation = renderHook(
      () => useCreateShipment("merchant-token"),
      { wrapper },
    )

    const result = await mutation.result.current.mutateAsync({
      sub_order_id: 4,
      carrier: "SF Express",
    })

    expect(result).toEqual(created)
    expect(requestBody).toEqual({
      sub_order_id: 4,
      carrier: "SF Express",
    })
  })
})
