import { fireEvent, screen, within } from "@testing-library/react"
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
      in_stock: true,
    },
    {
      id: 6,
      specs: { Color: "Black", Size: "L" },
      price_cents: 2499,
      is_enabled: true,
      in_stock: false,
    },
    {
      id: 8,
      specs: { Color: "White", Size: "L" },
      price_cents: 2999,
      is_enabled: true,
      in_stock: true,
    },
    {
      id: 9,
      specs: { Color: "White", Size: "XL" },
      price_cents: 3499,
      is_enabled: false,
      in_stock: true,
    },
  ],
}

describe("ProductDetailPage", () => {
  it("selects specifications in stages and supports cart and direct checkout", async () => {
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
      screen.getByRole("radio", { name: "Color: Black" }),
    ).toBeChecked()
    expect(
      screen.getByRole("radio", { name: "Color: White" }),
    ).toBeEnabled()
    expect(
      screen.getByRole("radio", { name: "Size: M" }),
    ).toBeChecked()
    expect(
      screen.getByRole("radio", { name: "Size: L" }),
    ).toBeEnabled()
    expect(
      screen.getByRole("radio", { name: "Size: XL" }),
    ).toBeDisabled()
    expect(screen.getByText(/19\.99/)).toBeInTheDocument()
    expect(screen.getByText("Out of stock")).toBeInTheDocument()
    expect(screen.queryByText("In stock")).not.toBeInTheDocument()
    expect(screen.queryByText("3 available")).not.toBeInTheDocument()

    await user.click(screen.getByRole("radio", { name: "Color: White" }))

    expect(
      screen.getByRole("radio", { name: "Color: White" }),
    ).toBeChecked()
    expect(screen.getByText("Select all options")).toBeInTheDocument()
    expect(
      screen.getByRole("radio", { name: "Size: M" }),
    ).toBeDisabled()
    expect(
      screen.getByRole("radio", { name: "Size: L" }),
    ).toBeEnabled()
    expect(
      screen.getByRole("radio", { name: "Size: XL" }),
    ).toBeDisabled()
    expect(
      screen.getByRole("button", { name: "Add to cart" }),
    ).toBeDisabled()

    await user.click(screen.getByRole("radio", { name: "Size: L" }))
    const quantityInput = screen.getByRole("spinbutton", {
      name: "Quantity",
    })
    expect(quantityInput).not.toHaveAttribute("max")
    fireEvent.change(quantityInput, { target: { value: "100" } })
    expect(screen.queryByText("2 available")).not.toBeInTheDocument()
    expect(screen.getByRole("link", { name: "Buy now" })).toHaveAttribute(
      "href",
      "/checkout?product_id=7&sku_id=8&quantity=100",
    )

    await user.click(screen.getByRole("button", { name: "Add to cart" }))

    expect(await screen.findByText("Added to your cart.")).toBeInTheDocument()
    expect(addToCart).toHaveBeenCalledWith(8, 100)
  })

  it("shows an out-of-stock selection and disables purchase actions", async () => {
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
    await user.click(screen.getByRole("radio", { name: "Size: L" }))

    const summary = screen.getByText(/24\.99/).parentElement
    expect(summary).not.toBeNull()
    expect(within(summary!).getByText("Out of stock")).toBeInTheDocument()
    expect(
      screen.getByRole("button", { name: "Add to cart" }),
    ).toBeDisabled()
    expect(
      screen.getByRole("button", { name: "Buy now" }),
    ).toBeDisabled()
    expect(
      screen.queryByRole("link", { name: "Buy now" }),
    ).not.toBeInTheDocument()
    expect(
      screen.getByRole("spinbutton", { name: "Quantity" }),
    ).toBeDisabled()
  })

  it("selects an out-of-stock SKU by default when none are purchasable", async () => {
    const outOfStockProduct = {
      ...product,
      skus: [
        {
          id: 5,
          specs: { Color: "Black", Size: "M" },
          price_cents: 1999,
          is_enabled: true,
          in_stock: false,
        },
      ],
    }
    server.use(
      http.get("*/api/v1/products/7", () =>
        HttpResponse.json(outOfStockProduct),
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
      await screen.findByRole("heading", { name: "Coffee beans" }),
    ).toBeInTheDocument()
    expect(
      screen.getByRole("radio", { name: "Size: M" }),
    ).toBeChecked()
    const summary = screen.getByText(/19\.99/).parentElement
    expect(summary).not.toBeNull()
    expect(within(summary!).getByText("Out of stock")).toBeInTheDocument()
    expect(
      screen.getByRole("button", { name: "Add to cart" }),
    ).toBeDisabled()
    expect(
      screen.getByRole("button", { name: "Buy now" }),
    ).toBeDisabled()
  })

  it("shows status per SKU when specifications cannot be staged", async () => {
    const user = userEvent.setup()
    const flatProduct = {
      ...product,
      skus: [
        {
          id: 5,
          specs: { Color: "Black" },
          price_cents: 1999,
          is_enabled: true,
          in_stock: true,
        },
        {
          id: 6,
          specs: { Size: "L" },
          price_cents: 2499,
          is_enabled: true,
          in_stock: false,
        },
        {
          id: 8,
          specs: {},
          price_cents: 2999,
          is_enabled: false,
          in_stock: true,
        },
      ],
    }
    server.use(
      http.get("*/api/v1/products/7", () =>
        HttpResponse.json(flatProduct),
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
      await screen.findByRole("heading", { name: "Coffee beans" }),
    ).toBeInTheDocument()

    const availableRow = screen
      .getByRole("radio", { name: "Color: Black" })
      .closest("label")
    const outOfStockRow = screen
      .getByRole("radio", { name: "Size: L" })
      .closest("label")
    const unavailableRow = screen
      .getByRole("radio", { name: "No specifications" })
      .closest("label")

    expect(availableRow).not.toBeNull()
    expect(outOfStockRow).not.toBeNull()
    expect(unavailableRow).not.toBeNull()
    expect(
      within(availableRow!).queryByText("Out of stock"),
    ).not.toBeInTheDocument()
    expect(
      within(availableRow!).queryByText("Unavailable"),
    ).not.toBeInTheDocument()
    expect(
      within(outOfStockRow!).getByText("Out of stock"),
    ).toBeInTheDocument()
    expect(
      within(unavailableRow!).getByText("Unavailable"),
    ).toBeInTheDocument()
    expect(
      screen.getByRole("radio", { name: "Size: L" }),
    ).toBeEnabled()
    expect(
      screen.getByRole("radio", { name: "No specifications" }),
    ).toBeDisabled()

    await user.click(screen.getByRole("radio", { name: "Size: L" }))

    expect(
      screen.getByRole("radio", { name: "Size: L" }),
    ).toBeChecked()
    expect(
      screen.getByRole("button", { name: "Add to cart" }),
    ).toBeDisabled()
  })

  it("disables purchase when the product has no SKUs", async () => {
    server.use(
      http.get("*/api/v1/products/7", () =>
        HttpResponse.json({ ...product, skus: [] }),
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
      await screen.findByText(
        "No options are available for this product.",
      ),
    ).toBeInTheDocument()
    expect(
      screen.getByRole("button", { name: "Add to cart" }),
    ).toBeDisabled()
    expect(
      screen.getByRole("button", { name: "Buy now" }),
    ).toBeDisabled()
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
