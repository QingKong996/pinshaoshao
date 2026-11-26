import { screen, waitFor } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { HttpResponse, http } from "msw"
import { describe, expect, it, vi } from "vitest"
import { renderWithProviders } from "@/test/render"
import { server } from "@/test/server"
import { order, orderID } from "@/test/commerce-fixtures"
import { OrderDetailPage } from "./order-detail-page"

describe("OrderDetailPage", () => {
  it("renders an order and cancels it", async () => {
    const user = userEvent.setup()
    let cancelled = false
    const renderSidebar = vi.fn(() => <p>Payment sidebar</p>)
    const renderSubOrderFooter = vi.fn(() => <p>Shipment footer</p>)
    server.use(
      http.get(`*/api/v1/orders/${orderID}`, () =>
        HttpResponse.json(
          cancelled
            ? {
                ...order,
                status: "cancelled",
                sub_orders: order.sub_orders.map((subOrder) => ({
                  ...subOrder,
                  status: "cancelled",
                })),
              }
            : order,
        ),
      ),
      http.post(`*/api/v1/orders/${orderID}/cancel`, () => {
        cancelled = true
        return HttpResponse.json({
          ...order,
          status: "cancelled",
          sub_orders: order.sub_orders.map((subOrder) => ({
            ...subOrder,
            status: "cancelled",
          })),
        })
      }),
    )

    renderWithProviders(
      <OrderDetailPage
        token="buyer-token"
        orderID={orderID}
        renderSidebar={renderSidebar}
        renderSubOrderFooter={renderSubOrderFooter}
      />,
      { initialEntries: [`/orders/${orderID}`] },
    )

    expect(
      await screen.findByRole("heading", { name: `Order #${orderID}` }),
    ).toBeVisible()
    expect(screen.getByText("Payment sidebar")).toBeVisible()
    expect(screen.getByText("Shipment footer")).toBeVisible()
    expect(screen.getByText("Cotton T-shirt")).toBeVisible()
    expect(screen.getByText("Size: M")).toBeVisible()
    expect(renderSidebar).toHaveBeenCalledWith(order)
    expect(renderSubOrderFooter).toHaveBeenCalledWith(
      order,
      order.sub_orders[0],
    )

    await user.click(screen.getByRole("button", { name: "Cancel order" }))
    expect(
      await screen.findByText("Cancel this order?"),
    ).toBeVisible()
    await user.click(screen.getByRole("button", { name: "Cancel order" }))

    await waitFor(() => {
      expect(screen.getAllByText("Cancelled").length).toBeGreaterThan(0)
    })
  })

  it("shows not-found for an invalid order ID", () => {
    renderWithProviders(
      <OrderDetailPage token="buyer-token" orderID={0} />,
    )

    expect(screen.getByText("Order not found")).toBeVisible()
  })

  it("shows a load error", async () => {
    server.use(
      http.get(`*/api/v1/orders/${orderID}`, () =>
        HttpResponse.json({ error: "unavailable" }, { status: 500 }),
      ),
    )

    renderWithProviders(
      <OrderDetailPage token="buyer-token" orderID={orderID} />,
    )

    expect(
      await screen.findByText("Could not load the order"),
    ).toBeVisible()
  })
})
