import { screen } from "@testing-library/react"
import { beforeEach, describe, expect, it, vi } from "vitest"
import { renderWithProviders } from "@/test/render"
import { OrderShipmentPanel } from "./order-shipment-panel"

const mocks = vi.hoisted(() => ({
  useOrderShipments: vi.fn(),
}))

vi.mock("@/features/logistics/model/queries", () => ({
  useOrderShipments: mocks.useOrderShipments,
}))

describe("OrderShipmentPanel", () => {
  beforeEach(() => {
    mocks.useOrderShipments.mockReturnValue({
      data: {
        items: [
          {
            shipment_no: "4f2614df-f674-4b5e-a9e4-3fc5e7f62013",
            sub_order_id: 4,
            order_id: 12,
            status: "in_transit",
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
          },
        ],
      },
      error: null,
      isPending: false,
      refetch: vi.fn(),
    })
  })

  it("shows tracking for a shipped sub-order", () => {
    renderWithProviders(
      <OrderShipmentPanel
        token="token"
        orderID={12}
        subOrderID={4}
        subOrderStatus="shipped"
      />,
       { initialEntries: ["/orders/12"] },
    )

    expect(screen.getByText("SF Express")).toBeVisible()
    expect(screen.getByText("SF123456789")).toBeVisible()
    expect(screen.getByText("In transit")).toBeVisible()
    expect(
      screen.getByRole("link", { name: "View tracking" }),
    ).toHaveAttribute(
      "href",
      "/shipments/4f2614df-f674-4b5e-a9e4-3fc5e7f62013",
    )
  })

  it("does not query tracking before the merchant ships", () => {
    renderWithProviders(
      <OrderShipmentPanel
        token="token"
        orderID={12}
        subOrderID={4}
        subOrderStatus="awaiting_shipment"
      />,
    )

    expect(
      screen.getByText("This sub-order is waiting for merchant dispatch."),
    ).toBeVisible()
    expect(mocks.useOrderShipments).toHaveBeenCalledWith(
      "token",
      12,
      false,
    )
  })
})
