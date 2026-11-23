import { screen, waitFor } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { HttpResponse, http } from "msw"
import { beforeEach, describe, expect, it, vi } from "vitest"
import type {
  CreateDirectOrderInput,
  CreateOrderInput,
} from "@/features/order/types"
import { CheckoutPage } from "@/features/order/ui/checkout-page"
import { renderWithProviders } from "@/test/render"
import { server } from "@/test/server"

const items = [
  {
    id: 1,
    product_name: "Cotton T-shirt",
    sku_specs: { Color: "Blue" },
    unit_price_cents: 1999,
    quantity: 2,
    line_total_cents: 3998,
  },
]

describe("CheckoutPage", () => {
  beforeEach(() => {
    Object.defineProperty(globalThis.crypto, "randomUUID", {
      configurable: true,
      value: vi.fn(() => "4f2614df-f674-4b5e-a9e4-3fc5e7f62013"),
    })
  })

  it("submits selected cart items with delivery details", async () => {
    const user = userEvent.setup()
    const requestBody = vi.fn<(body: CreateOrderInput) => void>()
    server.use(
      http.post("*/api/v1/orders", async ({ request }) => {
        requestBody((await request.json()) as CreateOrderInput)
        return HttpResponse.json(
          {
            id: 1,
            status: "pending_payment",
            total_amount_cents: 3998,
            receiver_name: "Alice",
            receiver_phone: "13800138000",
            shipping_address: "Shanghai",
            expires_at: "2026-11-20T12:30:00Z",
            paid_at: null,
            created_at: "2026-11-20T12:00:00Z",
            updated_at: "2026-11-20T12:00:00Z",
            sub_orders: [],
          },
          { status: 201 },
        )
      }),
    )

    renderWithProviders(
      <CheckoutPage token="token" items={items} />,
      { initialEntries: ["/checkout"] },
    )

    await user.clear(screen.getByLabelText("Recipient"))
    await user.type(screen.getByLabelText("Recipient"), "Alice")
    await user.type(screen.getByLabelText("Phone"), "13800138000")
    await user.type(screen.getByLabelText("Shipping address"), "Shanghai")
    await user.click(screen.getByRole("button", { name: "Place order" }))

    await waitFor(() => {
      expect(requestBody).toHaveBeenCalledOnce()
    })
    expect(requestBody.mock.calls[0][0]).toMatchObject({
      checkout_key: "4f2614df-f674-4b5e-a9e4-3fc5e7f62013",
      items: [{ cart_item_id: 1, quantity: 2 }],
      receiver_name: "Alice",
      receiver_phone: "13800138000",
      shipping_address: "Shanghai",
    })
  })

  it("submits a direct purchase without cart item IDs", async () => {
    const user = userEvent.setup()
    const requestBody = vi.fn<(body: CreateDirectOrderInput) => void>()
    server.use(
      http.post("*/api/v1/orders/direct", async ({ request }) => {
        requestBody((await request.json()) as CreateDirectOrderInput)
        return HttpResponse.json(
          {
            id: 2,
            status: "pending_payment",
            total_amount_cents: 3998,
            receiver_name: "Alice",
            receiver_phone: "13800138000",
            shipping_address: "Shanghai",
            expires_at: "2026-11-20T12:30:00Z",
            paid_at: null,
            created_at: "2026-11-20T12:00:00Z",
            updated_at: "2026-11-20T12:00:00Z",
            sub_orders: [],
          },
          { status: 201 },
        )
      }),
    )

    renderWithProviders(
      <CheckoutPage
        token="token"
        items={items}
        mode={{ type: "direct", skuID: 5 }}
        backHref="/products/10"
      />,
      { initialEntries: ["/checkout?sku_id=5"] },
    )

    expect(
      screen.getByRole("link", { name: "Back to product" }),
    ).toHaveAttribute("href", "/products/10")
    await user.type(screen.getByLabelText("Recipient"), "Alice")
    await user.type(screen.getByLabelText("Phone"), "13800138000")
    await user.type(screen.getByLabelText("Shipping address"), "Shanghai")
    await user.click(screen.getByRole("button", { name: "Place order" }))

    await waitFor(() => {
      expect(requestBody).toHaveBeenCalledOnce()
    })
    expect(requestBody.mock.calls[0][0]).toEqual({
      checkout_key: "4f2614df-f674-4b5e-a9e4-3fc5e7f62013",
      sku_id: 5,
      quantity: 2,
      receiver_name: "Alice",
      receiver_phone: "13800138000",
      shipping_address: "Shanghai",
    })
  })
})
