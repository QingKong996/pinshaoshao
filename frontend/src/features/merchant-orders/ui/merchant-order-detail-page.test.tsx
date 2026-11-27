import { screen } from "@testing-library/react"
import { beforeEach, describe, expect, it, vi } from "vitest"
import MerchantOrderDetailPage from "@/features/merchant-orders/ui/merchant-order-detail-page"
import type { MerchantOrderDetail } from "@/features/merchant-orders/types"
import { renderWithProviders } from "@/test/render"

const order: MerchantOrderDetail = {
  id: 4,
  order_id: 12,
  status: "awaiting_shipment",
  total_amount_cents: 3998,
  item_count: 2,
  paid_at: "2026-11-20T12:05:00Z",
  created_at: "2026-11-20T12:00:00Z",
  updated_at: "2026-11-20T12:05:00Z",
  items: [
    {
      id: 8,
      product_id: 10,
      sku_id: 5,
      product_name: "T-shirt",
      sku_specs: { color: "black" },
      unit_price_cents: 1999,
      quantity: 2,
      line_amount_cents: 3998,
    },
  ],
}

const receiver = {
  receiver_name: "Receiver",
  receiver_phone: "1234567890",
  shipping_address: "Shanghai",
}

const mocks = vi.hoisted(() => ({
  useMerchantOrder: vi.fn(),
  useMerchantOrderReceiver: vi.fn(),
}))

vi.mock("@/features/merchant-orders/model/queries", () => ({
  useMerchantOrder: mocks.useMerchantOrder,
  useMerchantOrderReceiver: mocks.useMerchantOrderReceiver,
}))

describe("MerchantOrderDetailPage", () => {
  beforeEach(() => {
    mocks.useMerchantOrder.mockReturnValue({
      data: order,
      error: null,
      isPending: false,
      refetch: vi.fn(),
    })
    mocks.useMerchantOrderReceiver.mockReturnValue({
      data: receiver,
      error: null,
      isPending: false,
      refetch: vi.fn(),
    })
  })

  it("renders the sub-order, items, and separately loaded delivery details", () => {
    renderWithProviders(
      <MerchantOrderDetailPage
        token="token"
        subOrderID={4}
        backHref="/merchant/orders"
      />,
      { initialEntries: ["/merchant/orders/4"] },
    )

    expect(
      screen.getByRole("heading", { name: /Sub-order #4/ }),
    ).toBeVisible()
    expect(screen.getByText(/Order #12 · Placed/)).toBeVisible()
    expect(screen.getByText("T-shirt")).toBeVisible()
    expect(screen.getByText("color: black")).toBeVisible()
    expect(screen.getByText("Awaiting shipment")).toBeVisible()
    expect(screen.getByText("Receiver")).toBeVisible()
    expect(screen.getByText("Shanghai")).toBeVisible()
    expect(
      screen.getByRole("link", { name: "Back to orders" }),
    ).toHaveAttribute("href", "/merchant/orders")
    expect(mocks.useMerchantOrder).toHaveBeenCalledWith("token", 4)
    expect(mocks.useMerchantOrderReceiver).toHaveBeenCalledWith(
      "token",
      4,
      true,
    )
  })

  it("does not load delivery details for an unpaid sub-order", () => {
    mocks.useMerchantOrder.mockReturnValue({
      data: { ...order, status: "pending_payment" },
      error: null,
      isPending: false,
      refetch: vi.fn(),
    })

    renderWithProviders(
      <MerchantOrderDetailPage
        token="token"
        subOrderID={4}
        backHref="/merchant/orders"
      />,
    )

    expect(screen.queryByText("Delivery details")).not.toBeInTheDocument()
    expect(mocks.useMerchantOrderReceiver).toHaveBeenCalledWith(
      "token",
      4,
      false,
    )
  })

  it("shows an API error and retries the query", async () => {
    mocks.useMerchantOrder.mockReturnValue({
      data: undefined,
      error: new Error("failed"),
      isPending: false,
      refetch: vi.fn(),
    })

    renderWithProviders(
      <MerchantOrderDetailPage
        token="token"
        subOrderID={4}
        backHref="/merchant/orders"
      />,
    )

    expect(screen.getByText("Could not load data")).toBeVisible()
    expect(
      screen.getByText("Something went wrong. Please try again later."),
    ).toBeVisible()
  })

  it("shows not found for an invalid sub-order id", () => {
    renderWithProviders(
      <MerchantOrderDetailPage
        token="token"
        subOrderID={0}
        backHref="/merchant/orders"
      />,
    )

    expect(screen.getByText("Could not load data")).toBeVisible()
    expect(
      screen.getByText("The resource does not exist or was deleted."),
    ).toBeVisible()
  })
})
