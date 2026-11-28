import { act, screen, waitFor } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { createHash, webcrypto } from "node:crypto"
import { HttpResponse, http } from "msw"
import type { InitialEntry } from "react-router"
import { beforeEach, describe, expect, it, vi } from "vitest"
import {
  UserAuthContext,
  type UserAuthContextValue,
} from "@/features/auth"
import { cartQueryKeys } from "@/features/cart"
import { feedQueryKeys } from "@/features/feed"
import { directCheckoutAttemptKey } from "@/features/order"
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
    Object.defineProperty(globalThis.crypto, "subtle", {
      configurable: true,
      value: webcrypto.subtle,
    })
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

  it("allows a pending direct purchase to retry after its stock is reserved", async () => {
    const user = userEvent.setup()
    const checkoutKey = "4f2614df-f674-4b5e-a9e4-3fc5e7f62013"
    const attemptStorageKey = directCheckoutAttemptKey("alice", 5, 2)
    const payload = {
      sku_id: 5,
      quantity: 2,
      receiver_name: "Alice",
      receiver_phone: "1234567890",
      shipping_address: "Shanghai",
    }
    const signature = createHash("sha256")
      .update(JSON.stringify(payload))
      .digest("hex")
    sessionStorage.setItem(
      attemptStorageKey,
      JSON.stringify({ signature, key: checkoutKey }),
    )
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
              in_stock: false,
            },
          ],
        }),
      ),
      http.post("*/api/v1/orders/direct", async ({ request }) => {
        expect(await request.json()).toMatchObject({
          checkout_key: checkoutKey,
        })
        return HttpResponse.json(
          {
            id: 6,
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
          { status: 200 },
        )
      }),
    )

    renderCheckoutRoute("/checkout?product_id=7&sku_id=5&quantity=2")

    expect(
      await screen.findByText("An unfinished purchase attempt was found"),
    ).toBeVisible()
    expect(screen.queryByRole("button", { name: /start a new purchase/i }))
      .not.toBeInTheDocument()
    await user.type(screen.getByLabelText("Recipient"), payload.receiver_name)
    await user.type(screen.getByLabelText("Phone"), payload.receiver_phone)
    await user.type(screen.getByLabelText("Shipping address"), payload.shipping_address)
    await user.click(
      screen.getByRole("button", { name: "Retry previous attempt" }),
    )

    await waitFor(() => {
      expect(sessionStorage.getItem(attemptStorageKey)).toBeNull()
    })
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

  it("requires confirmation when selected cart quantities change", async () => {
    const user = userEvent.setup()
    server.use(
      http.get("*/api/v1/cart", () =>
        HttpResponse.json([{ ...cartItem, quantity: 3, line_total_cents: 5997 }]),
      ),
    )

    renderCheckoutRoute({
      pathname: "/checkout",
      search: "?items=1",
      state: { cartSelectionSnapshot: JSON.stringify([[1, 5, 2, 1999]]) },
    })

    expect(
      await screen.findByText("Your selected cart items changed"),
    ).toBeVisible()
    expect(screen.getByText("3 item(s)")).toBeVisible()
    expect(screen.getByRole("button", { name: "Place order" })).toBeDisabled()

    await user.click(
      screen.getByRole("button", { name: "Confirm updated selection" }),
    )

    expect(screen.queryByRole("alert")).not.toBeInTheDocument()
    expect(screen.getByRole("button", { name: "Place order" })).toBeEnabled()
  })

  it("requires another confirmation if an available item disappears after review", async () => {
    const user = userEvent.setup()
    const queryClient = createTestQueryClient()
    const secondItem = { ...cartItem, id: 2, sku_id: 6, product_name: "Canvas Bag" }
    server.use(
      http.get("*/api/v1/cart", () => HttpResponse.json([cartItem, secondItem])),
    )
    renderCheckoutRoute("/checkout?items=1,2,3", queryClient)

    await user.click(
      await screen.findByRole("button", { name: "Confirm updated selection" }),
    )
    expect(screen.getByRole("button", { name: "Place order" })).toBeEnabled()

    act(() => {
      queryClient.setQueryData(cartQueryKeys.cart("token"), [cartItem])
    })

    expect(await screen.findByText("Your selected cart items changed")).toBeVisible()
    expect(screen.queryByText("Canvas Bag")).not.toBeInTheDocument()
    expect(screen.getByRole("button", { name: "Place order" })).toBeDisabled()
  })

  it("refreshes the cart and requests confirmation after a checkout conflict", async () => {
    const user = userEvent.setup()
    let items = [cartItem, { ...cartItem, id: 2, sku_id: 6 }]
    server.use(
      http.get("*/api/v1/cart", () => HttpResponse.json(items)),
      http.post("*/api/v1/orders", () => {
        items = [cartItem]
        return HttpResponse.json(
          { error: "selected cart items changed" },
          { status: 409 },
        )
      }),
    )
    renderCheckoutRoute("/checkout?items=1,2")

    await user.type(await screen.findByLabelText("Recipient"), "Receiver")
    await user.type(screen.getByLabelText("Phone"), "1234567890")
    await user.type(screen.getByLabelText("Shipping address"), "Shanghai")
    await user.click(screen.getByRole("button", { name: "Place order" }))

    expect(await screen.findByText("Your selected cart items changed")).toBeVisible()
    expect(screen.getByRole("button", { name: "Place order" })).toBeDisabled()
    expect(screen.getByLabelText("Recipient")).toHaveValue("Receiver")
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
  entry: InitialEntry,
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
        expireSession: vi.fn(() => false),
      }}
    >
      <CheckoutRoute />
    </UserAuthContext.Provider>,
    { queryClient, initialEntries: [entry] },
  )
}
