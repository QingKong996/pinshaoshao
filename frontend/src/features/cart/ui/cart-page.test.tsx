import { screen, waitFor, within } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { HttpResponse, http } from "msw"
import { beforeEach, describe, expect, it, vi } from "vitest"
import type { CartItem } from "@/features/cart/types"
import { CartPage } from "@/features/cart/ui/cart-page"
import { renderWithProviders } from "@/test/render"
import { server } from "@/test/server"

const items: CartItem[] = [
  {
    id: 1,
    product_id: 10,
    sku_id: 5,
    product_name: "Cotton T-shirt",
    sku_specs: { Color: "Blue", Size: "M" },
    unit_price_cents: 1999,
    quantity: 2,
    line_total_cents: 3998,
    available_stock: 8,
    availability: "available",
    created_at: "2026-11-19T08:00:00Z",
    updated_at: "2026-11-19T08:00:00Z",
  },
  {
    id: 2,
    product_id: 11,
    sku_id: 6,
    product_name: "Canvas Bag",
    sku_specs: {},
    unit_price_cents: 990,
    quantity: 1,
    line_total_cents: 990,
    available_stock: 3,
    availability: "product_off_shelf",
    created_at: "2026-11-19T09:00:00Z",
    updated_at: "2026-11-19T09:00:00Z",
  },
]

function mockCart(itemsResponse: CartItem[] = items) {
  server.use(
    http.get("*/api/v1/cart", () => HttpResponse.json(itemsResponse)),
  )
}

function renderCartPage() {
  return renderWithProviders(<CartPage token="token" />)
}

describe("CartPage", () => {
  beforeEach(() => {
    mockCart()
  })

  it("shows snapshot items and excludes unavailable items from selection", async () => {
    const user = userEvent.setup()
    renderCartPage()

    expect(await screen.findByText("Cotton T-shirt")).toBeVisible()
    expect(screen.getByText("Color: Blue / Size: M")).toBeVisible()
    expect(screen.getByText("Product is off shelf")).toBeVisible()

    const unavailable = screen.getByRole("checkbox", {
      name: "Select Canvas Bag for checkout",
    })
    expect(unavailable).toBeDisabled()

    const summary = screen.getByText("Selection").closest("div")
    const subtotal = screen.getByText("Subtotal").closest("div")
    expect(summary).not.toBeNull()
    expect(subtotal).not.toBeNull()
    expect(within(summary!).getByText("1")).toBeVisible()
    expect(within(subtotal!).getByText(/39\.98/)).toBeVisible()

    await user.click(
      screen.getByRole("checkbox", {
        name: "Select Cotton T-shirt for checkout",
      }),
    )

    expect(within(summary!).getByText("0")).toBeVisible()
    expect(within(subtotal!).getByText(/0\.00/)).toBeVisible()
  })

  it("updates quantity and removes an item", async () => {
    const user = userEvent.setup()
    const updateRequest = vi.fn()
    const deleteRequest = vi.fn()

    server.use(
      http.patch("*/api/v1/cart/items/1", async ({ request }) => {
        updateRequest(await request.json())
        return HttpResponse.json({
          ...items[0],
          quantity: 3,
          line_total_cents: 5997,
        })
      }),
      http.delete("*/api/v1/cart/items/1", () => {
        deleteRequest()
        return new HttpResponse(null, { status: 204 })
      }),
    )

    renderCartPage()

    const row = (await screen.findByText("Cotton T-shirt")).closest("li")
    expect(row).not.toBeNull()
    await user.click(
      within(row!).getByRole("button", { name: "Increase quantity" }),
    )
    await waitFor(() => {
      expect(updateRequest).toHaveBeenCalledWith({ quantity: 3 })
    })

    await user.click(
      within(row!).getByRole("button", { name: "Remove Cotton T-shirt" }),
    )
    expect(
      await screen.findByRole("heading", { name: "Remove this item?" }),
    ).toBeVisible()
    await user.click(screen.getByRole("button", { name: "Remove" }))

    await waitFor(() => {
      expect(deleteRequest).toHaveBeenCalledOnce()
    })
  })
})
