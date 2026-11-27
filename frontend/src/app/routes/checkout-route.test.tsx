import { screen, waitFor } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { HttpResponse, http } from "msw"
import { beforeEach, describe, expect, it, vi } from "vitest"
import {
  UserAuthContext,
  type UserAuthContextValue,
} from "@/features/auth"
import { feedQueryKeys } from "@/features/feed"
import { createTestQueryClient, renderWithProviders } from "@/test/render"
import { server } from "@/test/server"
import { CheckoutRoute } from "./checkout-route"

const product = {
  id: 7,
  shop_id: 2,
  shop_name: "Example shop",
  name: "Coffee beans",
  description: "Roasted for filter coffee",
  category_name: "Drinks",
}

const cartItem = {
  id: 1,
  product_id: 10,
  sku_id: 5,
  product_name: "Cotton T-shirt",
  sku_specs: { Color: "Blue" },
  unit_price_cents: 1999,
  quantity: 2,
  line_total_cents: 3998,
  availability: "available",
  created_at: "2026-11-19T08:00:00Z",
  updated_at: "2026-11-19T08:00:00Z",
}

describe("CheckoutRoute", () => {
  beforeEach(() => {
    Object.defineProperty(globalThis.crypto, "randomUUID", {
      configurable: true,
      value: vi.fn(() => "4f2614df-f674-4b5e-a9e4-3fc5e7f62013"),
    })
    server.use(
      http.get("*/api/v1/addresses", () => HttpResponse.json([])),
    )
  })

  it("loads a direct-purchase SKU without requesting the cart", async () => {
    let cartRequests = 0
    server.use(
      http.get("*/api/v1/products/7", () =>
        HttpResponse.json({
          ...product,
          skus: [
            {
              id: 5,
              specs: { Color: "Black" },
              price_cents: 1999,
              is_enabled: true,
              in_stock: true,
            },
          ],
        }),
      ),
      http.get("*/api/v1/cart", () => {
        cartRequests++
        return HttpResponse.json([])
      }),
    )

    renderCheckoutRoute("/checkout?product_id=7&sku_id=5&quantity=2")

    expect(
      await screen.findByRole("heading", { name: "Confirm order" }),
    ).toBeInTheDocument()
    expect(screen.getByText("Coffee beans")).toBeInTheDocument()
    expect(screen.getByText("Color: Black")).toBeInTheDocument()
    expect(screen.getByText("2 item(s)")).toBeInTheDocument()
    expect(
      screen.getByRole("link", { name: "Back to product" }),
    ).toHaveAttribute("href", "/products/7")
    expect(cartRequests).toBe(0)
  })

  it("rejects an invalid direct-purchase link", async () => {
    renderCheckoutRoute("/checkout?product_id=7&sku_id=nope")

    expect(
      screen.getByText("Invalid direct purchase link"),
    ).toBeInTheDocument()
    expect(
      screen.getByRole("link", { name: "Back to product" }),
    ).toHaveAttribute("href", "/products/7")
  })

  it("requires a session before loading checkout", () => {
    renderCheckoutRoute(
      "/checkout?product_id=7&sku_id=5&quantity=2",
      undefined,
      null,
    )

    expect(
      screen.getByRole("heading", { name: "Sign in to view orders" }),
    ).toBeInTheDocument()
  })

  it("loads selected cart items for checkout", async () => {
    server.use(
      http.get("*/api/v1/cart", () => HttpResponse.json([cartItem])),
    )

    renderCheckoutRoute("/checkout")

    expect(
      await screen.findByRole("heading", { name: "Confirm order" }),
    ).toBeInTheDocument()
    expect(screen.getByText("Cotton T-shirt")).toBeInTheDocument()
    expect(
      screen.getByRole("link", { name: "Back to cart" }),
    ).toHaveAttribute("href", "/cart")
  })

  it("prefills checkout with the default saved address", async () => {
    server.use(
      http.get("*/api/v1/cart", () => HttpResponse.json([cartItem])),
      http.get("*/api/v1/addresses", () =>
        HttpResponse.json([
          {
            id: 9,
            receiver_name: "Alice",
            receiver_phone: "2222222222",
            shipping_address: "Beijing",
            is_default: true,
            created_at: "2026-11-27T12:00:00Z",
            updated_at: "2026-11-27T12:00:00Z",
          },
        ]),
      ),
    )

    renderCheckoutRoute("/checkout")

    expect(
      await screen.findByRole("heading", { name: "Confirm order" }),
    ).toBeInTheDocument()
    expect(screen.getByLabelText("Recipient")).toHaveValue(
      "Alice",
    )
    expect(screen.getByLabelText("Phone")).toHaveValue(
      "2222222222",
    )
    expect(screen.getByLabelText("Shipping address")).toHaveValue(
      "Beijing",
    )
    expect(
      screen.getByLabelText("Use a saved address"),
    ).toBeInTheDocument()
  })

  it("reports a cart load failure", async () => {
    server.use(
      http.get("*/api/v1/cart", () =>
        HttpResponse.json({ error: "unavailable" }, { status: 500 }),
      ),
    )

    renderCheckoutRoute("/checkout")

    expect(
      await screen.findByText("Could not load your cart"),
    ).toBeInTheDocument()
  })

  it.each([
    {
      name: "a disabled SKU",
      sku: {
        id: 5,
        specs: { Color: "Black" },
        price_cents: 1999,
        is_enabled: false,
        in_stock: true,
      },
    },
    {
      name: "an out-of-stock SKU",
      sku: {
        id: 5,
        specs: { Color: "Black" },
        price_cents: 1999,
        is_enabled: true,
        in_stock: false,
      },
    },
  ])("rejects $name", async ({ sku }) => {
    server.use(
      http.get("*/api/v1/products/7", () =>
        HttpResponse.json({
          ...product,
          skus: [sku],
        }),
      ),
    )

    renderCheckoutRoute("/checkout?product_id=7&sku_id=5&quantity=2")

    expect(
      await screen.findByText("This option is unavailable"),
    ).toBeInTheDocument()
    expect(
      screen.getByRole("link", { name: "Back to product" }),
    ).toHaveAttribute("href", "/products/7")
  })

  it("invalidates feed data after a direct purchase", async () => {
    const user = userEvent.setup()
    const queryClient = createTestQueryClient({ gcTime: 60_000 })
    queryClient.setQueryData(feedQueryKeys.feed("token"), {
      pages: [],
      pageParams: [],
    })
    const requestBody = vi.fn<(body: unknown) => void>()
    server.use(
      http.get("*/api/v1/products/7", () =>
        HttpResponse.json({
          ...product,
          skus: [
            {
              id: 5,
              specs: { Color: "Black" },
              price_cents: 1999,
              is_enabled: true,
              in_stock: true,
            },
          ],
        }),
      ),
      http.post("*/api/v1/orders/direct", async ({ request }) => {
        requestBody(await request.json())
        return HttpResponse.json(
          {
            id: 2,
            status: "pending_payment",
            total_amount_cents: 3998,
            receiver_name: "Receiver",
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
      http.get("*/api/v1/cart", () => HttpResponse.json([])),
    )

    renderCheckoutRoute(
      "/checkout?product_id=7&sku_id=5&quantity=2",
      queryClient,
    )

    expect(
      await screen.findByRole("heading", { name: "Confirm order" }),
    ).toBeInTheDocument()
    await user.type(screen.getByLabelText("Recipient"), "Receiver")
    await user.type(screen.getByLabelText("Phone"), "1234567890")
    await user.type(screen.getByLabelText("Shipping address"), "Shanghai")
    await user.click(screen.getByRole("button", { name: "Place order" }))

    await waitFor(() => {
      expect(requestBody).toHaveBeenCalledWith({
        checkout_key: "4f2614df-f674-4b5e-a9e4-3fc5e7f62013",
        sku_id: 5,
        quantity: 2,
        receiver_name: "Receiver",
        receiver_phone: "1234567890",
        shipping_address: "Shanghai",
      })
    })
    await waitFor(() => {
      expect(
        queryClient.getQueryState(feedQueryKeys.feed("token"))
          ?.isInvalidated,
      ).toBe(true)
    })
  })
})

function renderCheckoutRoute(
  path: string,
  queryClient?: ReturnType<typeof createTestQueryClient>,
  session: UserAuthContextValue["session"] = {
    token: "token",
    username: "alice",
  },
) {
  return renderWithProviders(
    <UserAuthContext.Provider
      value={{
        session,
        signIn: vi.fn(async () => undefined),
        signUp: vi.fn(async () => undefined),
        signOut: vi.fn(async () => undefined),
      }}
    >
      <CheckoutRoute />
    </UserAuthContext.Provider>,
    { queryClient, initialEntries: [path] },
  )
}
