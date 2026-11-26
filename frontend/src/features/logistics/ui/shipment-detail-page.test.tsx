import { screen } from "@testing-library/react"
import { HttpResponse, http } from "msw"
import { describe, expect, it } from "vitest"
import { renderWithProviders } from "@/test/render"
import { server } from "@/test/server"
import {
  shipment,
  shipmentNo,
} from "@/test/commerce-fixtures"
import { ShipmentDetailPage } from "./shipment-detail-page"
import { ShipmentTimeline } from "./shipment-timeline"

describe("ShipmentDetailPage", () => {
  it("renders buyer tracking with history", async () => {
    server.use(
      http.get(`*/api/v1/shipments/${shipmentNo}`, () =>
        HttpResponse.json({
          ...shipment,
          status: "delivered",
          delivered_at: "2026-11-25T09:00:00Z",
        }),
      ),
    )

    renderWithProviders(
      <ShipmentDetailPage
        token="buyer-token"
        shipmentNo={shipmentNo}
        scope="buyer"
        backTo="/orders/12"
        backLabel="Back to orders"
      />,
      { initialEntries: [`/shipments/${shipmentNo}`] },
    )

    expect(
      await screen.findByRole("heading", { name: "Shipment tracking" }),
    ).toBeVisible()
    expect(screen.getByText("SF Express")).toBeVisible()
    expect(screen.getByText("SF123456789")).toBeVisible()
    expect(screen.getAllByText("Delivered").length).toBeGreaterThan(0)
    expect(screen.getByText("Departed Shanghai")).toBeVisible()
    expect(screen.getAllByText("Picked up").length).toBeGreaterThan(0)
    expect(
      screen.getByRole("link", { name: "Back to orders" }),
    ).toHaveAttribute("href", "/orders/12")
  })

  it("renders merchant tracking through the merchant endpoint", async () => {
    server.use(
      http.get(`*/api/v1/merchant/shipments/${shipmentNo}`, () =>
        HttpResponse.json(shipment),
      ),
    )

    renderWithProviders(
      <ShipmentDetailPage
        token="merchant-token"
        shipmentNo={shipmentNo}
        scope="merchant"
        backTo="/merchant/logistics"
        backLabel="Back to shipments"
      />,
      { initialEntries: [`/merchant/logistics/${shipmentNo}`] },
    )

    expect(
      (await screen.findAllByText("In transit")).length,
    ).toBeGreaterThan(0)
    expect(
      screen.getByRole("link", { name: "Back to shipments" }),
    ).toHaveAttribute("href", "/merchant/logistics")
  })
})

describe("ShipmentTimeline", () => {
  it("renders an empty state without events", () => {
    renderWithProviders(<ShipmentTimeline events={[]} />)

    expect(
      screen.getByText(
        "Tracking updates will appear here as the carrier scans the parcel.",
      ),
    ).toBeVisible()
  })
})
