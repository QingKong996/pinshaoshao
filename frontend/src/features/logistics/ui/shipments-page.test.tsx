import { screen, waitFor, within } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { HttpResponse, http } from "msw"
import { describe, expect, it, vi } from "vitest"
import { renderWithProviders } from "@/test/render"
import { server } from "@/test/server"
import { shipment, shipmentSummary } from "@/test/commerce-fixtures"
import MerchantShipmentsPage from "./shipments-page"

describe("MerchantShipmentsPage", () => {
  it("renders shipments and creates a new shipment", async () => {
    const user = userEvent.setup()
    const listRequests = vi.fn()
    const createBody = vi.fn<(body: unknown) => void>()
    server.use(
      http.get("*/api/v1/merchant/shipments", ({ request }) => {
        listRequests(new URL(request.url).searchParams)
        return HttpResponse.json({
          items: [shipmentSummary],
          page: 1,
          page_size: 10,
          total: 1,
          total_pages: 1,
        })
      }),
      http.post("*/api/v1/merchant/shipments", async ({ request }) => {
        createBody(await request.json())
        return HttpResponse.json(shipment, { status: 201 })
      }),
    )

    renderWithProviders(<MerchantShipmentsPage token="merchant-token" />)

    expect(await screen.findByText("#4f2614df")).toBeVisible()
    expect(screen.getByText("Receiver")).toBeVisible()
    expect(screen.getByText("SF Express")).toBeVisible()
    expect(screen.getByText("In transit")).toBeVisible()
    expect(screen.getByText("Shanghai")).toBeVisible()
    expect(listRequests).toHaveBeenCalled()

    await user.click(screen.getByRole("button", { name: "New shipment" }))
    const dialog = await screen.findByRole("dialog")
    await user.type(within(dialog).getByLabelText("Sub-order ID"), "4")
    await user.type(within(dialog).getByLabelText("Carrier"), "SF Express")
    await user.click(
      within(dialog).getByRole("button", { name: "Create shipment" }),
    )

    await waitFor(() => {
      expect(createBody).toHaveBeenCalledWith({
        sub_order_id: 4,
        carrier: "SF Express",
      })
    })
    await waitFor(() => {
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument()
    })
  })

  it("shows validation errors before submitting", async () => {
    const user = userEvent.setup()
    server.use(
      http.get("*/api/v1/merchant/shipments", () =>
        HttpResponse.json({
          items: [],
          page: 1,
          page_size: 10,
          total: 0,
          total_pages: 0,
        }),
      ),
    )

    renderWithProviders(<MerchantShipmentsPage token="merchant-token" />)

    expect(await screen.findByText("No shipments yet")).toBeVisible()
    await user.click(screen.getByRole("button", { name: "New shipment" }))
    await user.click(
      screen.getByRole("button", { name: "Create shipment" }),
    )

    expect(
      await screen.findByText("Enter a sub-order ID."),
    ).toBeVisible()
  })

  it("shows a load error", async () => {
    server.use(
      http.get("*/api/v1/merchant/shipments", () =>
        HttpResponse.json({ error: "unavailable" }, { status: 500 }),
      ),
    )

    renderWithProviders(<MerchantShipmentsPage token="merchant-token" />)

    expect(
      await screen.findByText("Could not load shipments"),
    ).toBeVisible()
  })
})
