import { screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { HttpResponse, http } from "msw"
import { describe, expect, it, vi } from "vitest"
import { renderWithProviders } from "@/test/render"
import { server } from "@/test/server"
import { ProductDetailPage } from "./product-detail-page"

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
      specs: { Color: "Black", Size: "M" },
      price_cents: 1999,
      is_enabled: true,
      available_stock: 3,
    },
    {
      id: 6,
      specs: { Color: "White", Size: "L" },
      price_cents: 2499,
      is_enabled: true,
      available_stock: 2,
    },
    {
      id: 8,
      specs: { Color: "Red", Size: "S" },
      price_cents: 2999,
      is_enabled: false,
      available_stock: 0,
    },
    {
      id: 9,
      specs: { Color: "Green", Size: "XL" },
      price_cents: 1499,
      is_enabled: true,
      available_stock: 0,
    },
  ],
}

describe("ProductDetailPage", () => {
  it("shows every SKU and supports cart and direct checkout actions", async () => {
    const user = userEvent.setup()
    const addToCart = vi.fn(async () => undefined)
    server.use(
      http.get("*/api/v1/products/7", () => HttpResponse.json(product)),
    )

    renderWithProviders(
      <ProductDetailPage
        token="token"
        productID={7}
        onAddToCart={addToCart}
      />,
      { initialEntries: ["/products/7"] },
    )

    expect(
      await screen.findByRole("heading", { name: "Coffee beans" }),
    ).toBeInTheDocument()
    expect(
      screen.getByRole("radio", { name: /Color: Black \/ Size: M/ }),
    ).toBeChecked()
    expect(
      screen.getByRole("radio", { name: /Color: White \/ Size: L/ }),
    ).toBeEnabled()
    expect(
      screen.getByRole("radio", { name: /Color: Red \/ Size: S/ }),
    ).toBeDisabled()
    expect(
      screen.getByRole("radio", { name: /Color: Green \/ Size: XL/ }),
    ).toBeDisabled()
    expect(screen.getAllByText("Unavailable")).toHaveLength(2)
    expect(screen.getAllByText(/19\.99/)).toHaveLength(2)
    expect(screen.getByText(/24\.99/)).toBeInTheDocument()
    expect(screen.getByText(/29\.99/)).toBeInTheDocument()
    expect(screen.getByText(/14\.99/)).toBeInTheDocument()

    await user.click(
      screen.getByRole("radio", { name: /Color: White \/ Size: L/ }),
    )
    await user.click(
      screen.getByRole("button", { name: "Increase quantity" }),
    )
    expect(
      screen.getByRole("button", { name: "Increase quantity" }),
    ).toBeDisabled()
    expect(screen.getByRole("link", { name: "Buy now" })).toHaveAttribute(
      "href",
      "/checkout?product_id=7&sku_id=6&quantity=2",
    )

    await user.click(screen.getByRole("button", { name: "Add to cart" }))

    expect(await screen.findByText("Added to your cart.")).toBeInTheDocument()
    expect(addToCart).toHaveBeenCalledWith(6, 2)
  })

  it("shows a product-not-found message", async () => {
    server.use(
      http.get(
        "*/api/v1/products/7",
        () => new HttpResponse(null, { status: 404 }),
      ),
    )

    renderWithProviders(
      <ProductDetailPage
        token="token"
        productID={7}
        onAddToCart={vi.fn(async () => undefined)}
      />,
      { initialEntries: ["/products/7"] },
    )

    expect(
      await screen.findByText("This product is not available."),
    ).toBeInTheDocument()
  })

  it("offers a way back when the product link is invalid", () => {
    renderWithProviders(
      <ProductDetailPage
        token="token"
        productID={0}
        onAddToCart={vi.fn(async () => undefined)}
      />,
      { initialEntries: ["/products/invalid"] },
    )

    expect(
      screen.getByText("The product link is invalid."),
    ).toBeInTheDocument()
    expect(
      screen.getByRole("link", { name: "Back to feed" }),
    ).toHaveAttribute("href", "/")
  })
})
