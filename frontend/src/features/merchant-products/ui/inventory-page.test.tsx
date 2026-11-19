import { screen, waitFor } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { beforeEach, describe, expect, it, vi } from "vitest"
import type { InventoryItem } from "@/features/merchant-products/types"
import MerchantConsoleInventoryPage from "@/features/merchant-products/ui/inventory-page"
import { renderWithProviders } from "@/test/render"

const items: InventoryItem[] = [
  {
    sku_id: 1,
    product_id: 10,
    product_name: "Cotton T-shirt",
    category_id: 3,
    category_name: "Clothing",
    specs: { Color: "Blue", Size: "M" },
    price_cents: 1990,
    total_stock: 8,
    frozen_stock: 2,
    available_stock: 6,
    version: 3,
    is_enabled: true,
    updated_at: "2026-11-15T08:00:00Z",
  },
  {
    sku_id: 2,
    product_id: 11,
    product_name: "Canvas Bag",
    category_id: null,
    category_name: null,
    specs: {},
    price_cents: 990,
    total_stock: 0,
    frozen_stock: 0,
    available_stock: 0,
    version: 0,
    is_enabled: false,
    updated_at: "2026-11-15T09:00:00Z",
  },
]

const mocks = vi.hoisted(() => ({
  useInventory: vi.fn(),
  useUpdateInventoryStock: vi.fn(),
}))

vi.mock("@/features/merchant-products/model/queries", () => ({
  useInventory: mocks.useInventory,
}))

vi.mock("@/features/merchant-products/model/mutations", () => ({
  useUpdateInventoryStock: mocks.useUpdateInventoryStock,
}))

const updateMutation = {
  error: null,
  isPending: false,
  mutateAsync: vi.fn().mockResolvedValue(undefined),
  reset: vi.fn(),
}

function renderPage() {
  return renderWithProviders(
    <MerchantConsoleInventoryPage
      token="token"
      productHref={(productID) => `/merchant/products/${productID}`}
      onCatalogChanged={() => Promise.resolve()}
    />,
  )
}

describe("MerchantConsoleInventoryPage", () => {
  beforeEach(() => {
    mocks.useInventory.mockReturnValue({
      data: items,
      error: null,
      isPending: false,
      refetch: vi.fn(),
    })
    mocks.useUpdateInventoryStock.mockReturnValue(updateMutation)
  })

  it("searches products and spec values locally", async () => {
    const user = userEvent.setup()
    renderPage()

    expect(
      screen.getByRole("link", { name: "Cotton T-shirt" }),
    ).toBeVisible()
    expect(screen.getByRole("link", { name: "Canvas Bag" })).toBeVisible()

    await user.type(
      screen.getByLabelText("Search products or specs"),
      "blue",
    )

    await waitFor(() => {
      expect(
        screen.getByRole("link", { name: "Cotton T-shirt" }),
      ).toBeVisible()
      expect(
        screen.queryByRole("link", { name: "Canvas Bag" }),
      ).not.toBeInTheDocument()
    })
  })

  it("submits a stock adjustment", async () => {
    const user = userEvent.setup()
    renderPage()

    await user.click(
      screen.getAllByRole("button", { name: "Adjust stock" })[0],
    )
    expect(
      await screen.findByRole("heading", { name: "Adjust total stock" }),
    ).toBeVisible()
    expect(screen.getByText("Total: 8 · Frozen: 2")).toBeVisible()

    const stockInput = screen.getByLabelText("Total stock")
    await user.clear(stockInput)
    await user.type(stockInput, "12")
    await user.click(screen.getByRole("button", { name: "Save" }))

    expect(updateMutation.mutateAsync).toHaveBeenCalledWith({
      skuID: 1,
      totalStock: 12,
      version: 3,
    })
  })
})
