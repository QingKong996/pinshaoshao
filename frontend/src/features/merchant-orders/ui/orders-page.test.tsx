import { screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { beforeEach, describe, expect, it, vi } from "vitest"
import MerchantOrdersPage from "@/features/merchant-orders/ui/orders-page"
import type { MerchantOrder } from "@/features/merchant-orders/types"
import { renderWithProviders } from "@/test/render"

const order: MerchantOrder = {
  id: 4,
  order_id: 12,
  status: "awaiting_shipment",
  total_amount_cents: 3998,
  item_count: 2,
  paid_at: "2026-11-20T12:05:00Z",
  created_at: "2026-11-20T12:00:00Z",
}

const mocks = vi.hoisted(() => ({
  useMerchantOrders: vi.fn(),
}))

vi.mock("@/features/merchant-orders/model/queries", () => ({
  useMerchantOrders: mocks.useMerchantOrders,
}))

describe("MerchantOrdersPage", () => {
  beforeEach(() => {
    mocks.useMerchantOrders.mockReturnValue({
      data: {
        items: [order],
        page: 1,
        page_size: 10,
        total: 1,
        total_pages: 1,
      },
      error: null,
      isPending: false,
      refetch: vi.fn(),
    })
  })

  it("renders the merchant order list", () => {
    renderWithProviders(<MerchantOrdersPage token="token" />)

    expect(
      screen.getByRole("heading", { name: "Orders" }),
    ).toBeVisible()
    expect(screen.getByText("#12")).toBeVisible()
    expect(screen.getByText("Sub-order #4")).toBeVisible()
    expect(screen.getByText("Awaiting shipment")).toBeVisible()
    expect(screen.queryByText("Recipient")).not.toBeInTheDocument()
    expect(screen.queryByText("Shipping address")).not.toBeInTheDocument()
    expect(mocks.useMerchantOrders).toHaveBeenCalledWith("token", {
      page: 1,
      page_size: 10,
    })
  })

  it("requests the next page", async () => {
    const user = userEvent.setup()
    mocks.useMerchantOrders.mockReturnValue({
      data: {
        items: [order],
        page: 1,
        page_size: 10,
        total: 11,
        total_pages: 2,
      },
      error: null,
      isPending: false,
      refetch: vi.fn(),
    })

    renderWithProviders(<MerchantOrdersPage token="token" />)
    await user.click(screen.getByRole("button", { name: "Next page" }))

    expect(mocks.useMerchantOrders).toHaveBeenLastCalledWith("token", {
      page: 2,
      page_size: 10,
    })
  })

  it("shows an API error and retries the query", async () => {
    const user = userEvent.setup()
    const refetch = vi.fn()
    mocks.useMerchantOrders.mockReturnValue({
      data: undefined,
      error: new Error("failed"),
      isPending: false,
      refetch,
    })

    renderWithProviders(<MerchantOrdersPage token="token" />)

    expect(screen.getByText("Could not load data")).toBeVisible()
    expect(
      screen.getByText("Something went wrong. Please try again later."),
    ).toBeVisible()
    await user.click(screen.getByRole("button", { name: "Retry" }))
    expect(refetch).toHaveBeenCalledOnce()
  })
})
