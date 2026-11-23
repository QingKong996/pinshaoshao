import { screen, waitFor } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { HttpResponse, http } from "msw"
import { Route, Routes } from "react-router"
import { describe, expect, it, vi } from "vitest"
import { UserAuthContext } from "@/features/auth"
import { cartQueryKeys } from "@/features/cart"
import { createTestQueryClient, renderWithProviders } from "@/test/render"
import { server } from "@/test/server"
import { ProductDetailRoute } from "./product-detail-route"

const product = {
  id: 7,
  shop_id: 2,
  shop_name: "Example shop",
  name: "Coffee beans",
  description: "Roasted for filter coffee",
  category_name: "Drinks",
  skus: [
    {
      id: 5,
      specs: { Color: "Black" },
      price_cents: 1999,
      is_enabled: true,
      available_stock: 4,
    },
  ],
}

describe("ProductDetailRoute", () => {
  it("prompts signed-out buyers to sign in", () => {
    renderWithProviders(
      <UserAuthContext.Provider
        value={{
          session: null,
          signIn: vi.fn(async () => undefined),
          signUp: vi.fn(async () => undefined),
          signOut: vi.fn(async () => undefined),
        }}
      >
        <Routes>
          <Route
            path="/products/:productID"
            element={<ProductDetailRoute />}
          />
        </Routes>
      </UserAuthContext.Provider>,
      { initialEntries: ["/products/7"] },
    )

    expect(
      screen.getByText("Sign in to view product details"),
    ).toBeInTheDocument()
    expect(screen.getByRole("link", { name: "Sign in" })).toHaveAttribute(
      "href",
      "/signin",
    )
  })

  it("adds the selected SKU through the cart feature and invalidates its cache", async () => {
    const user = userEvent.setup()
    const requestBody = vi.fn<(body: unknown) => void>()
    const queryClient = createTestQueryClient({ gcTime: 60_000 })
    queryClient.setQueryData(cartQueryKeys.cart("token"), [])
    server.use(
      http.get("*/api/v1/products/7", () => HttpResponse.json(product)),
      http.post("*/api/v1/cart/items", async ({ request }) => {
        requestBody(await request.json())
        return HttpResponse.json({ id: 1 }, { status: 201 })
      }),
    )

    renderWithProviders(
      <UserAuthContext.Provider
        value={{
          session: { token: "token", username: "alice" },
          signIn: vi.fn(async () => undefined),
          signUp: vi.fn(async () => undefined),
          signOut: vi.fn(async () => undefined),
        }}
      >
        <Routes>
          <Route
            path="/products/:productID"
            element={<ProductDetailRoute />}
          />
        </Routes>
      </UserAuthContext.Provider>,
      {
        queryClient,
        initialEntries: ["/products/7"],
      },
    )

    await user.click(
      await screen.findByRole("button", { name: "Add to cart" }),
    )

    await waitFor(() => {
      expect(requestBody).toHaveBeenCalledWith({
        sku_id: 5,
        quantity: 1,
      })
    })
    await waitFor(() => {
      expect(
        queryClient.getQueryState(cartQueryKeys.cart("token"))
          ?.isInvalidated,
      ).toBe(true)
    })
  })
})
