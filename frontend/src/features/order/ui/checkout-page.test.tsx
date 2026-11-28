import { screen, waitFor } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { webcrypto } from "node:crypto"
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
    Object.defineProperty(globalThis.crypto, "subtle", {
      configurable: true,
      value: webcrypto.subtle,
    })
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

  it("reuses a cart checkout key after an uncertain failure and page reload", async () => {
    const user = userEvent.setup()
    const keys: string[] = []
    let attempts = 0
    server.use(
      http.post("*/api/v1/orders", async ({ request }) => {
        const input = (await request.json()) as CreateOrderInput
        keys.push(input.checkout_key)
        attempts++
        if (attempts === 1) {
          return HttpResponse.error()
        }
        return HttpResponse.json(
          {
            id: 8,
            status: "pending_payment",
            total_amount_cents: 3998,
            receiver_name: "Alice",
            receiver_phone: "1234567890",
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
    const props = {
      token: "token",
      items,
      cartAttemptStorageKey: 'cart-checkout-attempt:"alice"',
    }

    const firstRender = renderWithProviders(<CheckoutPage {...props} />, {
      initialEntries: ["/checkout"],
    })
    await user.type(screen.getByLabelText("Recipient"), "Alice")
    await user.type(screen.getByLabelText("Phone"), "1234567890")
    await user.type(screen.getByLabelText("Shipping address"), "Shanghai")
    await user.click(screen.getByRole("button", { name: "Place order" }))
    await waitFor(() => expect(attempts).toBe(1))
    expect(sessionStorage.getItem(props.cartAttemptStorageKey)).not.toBeNull()

    firstRender.unmount()
    renderWithProviders(<CheckoutPage {...props} />, {
      initialEntries: ["/checkout"],
    })
    await user.type(screen.getByLabelText("Recipient"), "Alice")
    await user.type(screen.getByLabelText("Phone"), "1234567890")
    await user.type(screen.getByLabelText("Shipping address"), "Shanghai")
    await user.click(screen.getByRole("button", { name: "Place order" }))

    await waitFor(() => {
      expect(sessionStorage.getItem(props.cartAttemptStorageKey)).toBeNull()
    })
    expect(keys).toEqual([
      "4f2614df-f674-4b5e-a9e4-3fc5e7f62013",
      "4f2614df-f674-4b5e-a9e4-3fc5e7f62013",
    ])
    expect(crypto.randomUUID).toHaveBeenCalledOnce()
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
        mode={{
          type: "direct",
          skuID: 5,
          quantity: 2,
          attemptStorageKey: "direct-checkout-attempt:alice:5:2",
          canStartNew: true,
        }}
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

  it("reuses a direct checkout key after an uncertain failure and page reload", async () => {
    const user = userEvent.setup()
    const keys: string[] = []
    let attempts = 0
    server.use(
      http.post("*/api/v1/orders/direct", async ({ request }) => {
        const input = (await request.json()) as CreateDirectOrderInput
        keys.push(input.checkout_key)
        attempts++
        if (attempts === 1) {
          return HttpResponse.error()
        }
        return HttpResponse.json(
          {
            id: 4,
            status: "pending_payment",
            total_amount_cents: 3998,
            receiver_name: "Alice",
            receiver_phone: "1234567890",
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
    const props = {
      token: "token",
      items,
      mode: {
        type: "direct" as const,
        skuID: 5,
        quantity: 2,
        attemptStorageKey: "direct-checkout-attempt:alice:5:2",
        canStartNew: true,
      },
    }

    const firstRender = renderWithProviders(
      <CheckoutPage {...props} />,
      { initialEntries: ["/checkout?sku_id=5"] },
    )
    await user.type(screen.getByLabelText("Recipient"), "Alice")
    await user.type(screen.getByLabelText("Phone"), "1234567890")
    await user.type(screen.getByLabelText("Shipping address"), "Shanghai")
    await user.click(screen.getByRole("button", { name: "Place order" }))
    await waitFor(() => expect(attempts).toBe(1))
    expect(sessionStorage.getItem(props.mode.attemptStorageKey)).not.toBeNull()

    firstRender.unmount()
    renderWithProviders(
      <CheckoutPage {...props} />,
      { initialEntries: ["/checkout?sku_id=5"] },
    )
    await user.type(screen.getByLabelText("Recipient"), "Alice")
    await user.type(screen.getByLabelText("Phone"), "1234567890")
    await user.type(screen.getByLabelText("Shipping address"), "Shanghai")
    await user.click(screen.getByRole("button", { name: "Retry previous attempt" }))

    await waitFor(() => {
      expect(sessionStorage.getItem(props.mode.attemptStorageKey)).toBeNull()
    })
    expect(keys).toEqual([
      "4f2614df-f674-4b5e-a9e4-3fc5e7f62013",
      "4f2614df-f674-4b5e-a9e4-3fc5e7f62013",
    ])
    expect(crypto.randomUUID).toHaveBeenCalledOnce()
  })

  it("requires explicit confirmation before starting with changed delivery details", async () => {
    const user = userEvent.setup()
    const keys: string[] = []
    let attempts = 0
    server.use(
      http.post("*/api/v1/orders/direct", async ({ request }) => {
        const input = (await request.json()) as CreateDirectOrderInput
        keys.push(input.checkout_key)
        attempts++
        if (attempts === 1) {
          return HttpResponse.json(
            { error: "temporarily unavailable" },
            { status: 500 },
          )
        }
        return HttpResponse.json(
          {
            id: 5,
            status: "pending_payment",
            total_amount_cents: 3998,
            receiver_name: "Alice",
            receiver_phone: "1234567890",
            shipping_address: "Beijing",
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
        mode={{
          type: "direct",
          skuID: 5,
          quantity: 2,
          attemptStorageKey: "direct-checkout-attempt:alice:5:2",
          canStartNew: true,
        }}
      />,
      { initialEntries: ["/checkout?sku_id=5"] },
    )
    await user.type(screen.getByLabelText("Recipient"), "Alice")
    await user.type(screen.getByLabelText("Phone"), "1234567890")
    await user.type(screen.getByLabelText("Shipping address"), "Shanghai")
    await user.click(screen.getByRole("button", { name: "Place order" }))
    await waitFor(() => expect(attempts).toBe(1))

    await user.clear(screen.getByLabelText("Shipping address"))
    await user.type(screen.getByLabelText("Shipping address"), "Beijing")
    await user.click(screen.getByRole("button", { name: "Retry previous attempt" }))
    expect(await screen.findByText(/delivery details differ/i)).toBeVisible()
    expect(attempts).toBe(1)

    vi.mocked(crypto.randomUUID).mockReturnValueOnce(
      "11111111-1111-4111-8111-111111111111",
    )
    await user.click(
      screen.getByRole("button", {
        name: "Start a new purchase with these details",
      }),
    )
    await waitFor(() => expect(attempts).toBe(2))
    expect(keys[1]).not.toBe(keys[0])
  })

  it("prefills the default saved address and submits it as a snapshot", async () => {
    const user = userEvent.setup()
    const requestBody = vi.fn<(body: CreateOrderInput) => void>()
    server.use(
      http.post("*/api/v1/orders", async ({ request }) => {
        requestBody((await request.json()) as CreateOrderInput)
        return HttpResponse.json(
          {
            id: 3,
            status: "pending_payment",
            total_amount_cents: 3998,
            receiver_name: "Robert",
            receiver_phone: "1234567890",
            shipping_address: "Beijing",
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
        savedAddresses={[
          {
            id: 1,
            receiver_name: "Alice",
            receiver_phone: "111",
            shipping_address: "Shanghai",
            is_default: false,
          },
          {
            id: 2,
            receiver_name: "Robert",
            receiver_phone: "1234567890",
            shipping_address: "Beijing",
            is_default: true,
          },
        ]}
      />,
      { initialEntries: ["/checkout"] },
    )

    await waitFor(() => {
      expect(screen.getByLabelText("Recipient")).toHaveValue("Robert")
    })
    expect(screen.getByLabelText("Phone")).toHaveValue("1234567890")
    expect(screen.getByLabelText("Shipping address")).toHaveValue(
      "Beijing",
    )
    expect(
      screen.getByLabelText("Use a saved address"),
    ).toBeInTheDocument()

    await user.click(screen.getByRole("button", { name: "Place order" }))

    await waitFor(() => {
      expect(requestBody).toHaveBeenCalledOnce()
    })
    expect(requestBody.mock.calls[0][0]).toMatchObject({
      receiver_name: "Robert",
      receiver_phone: "1234567890",
      shipping_address: "Beijing",
    })
  })
})
