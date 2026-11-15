import { screen, waitFor, within } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { beforeEach, describe, expect, it, vi } from "vitest"
import MerchantConsoleProductsPage from "@/features/merchant-products/ui/products-page"
import type { Product } from "@/features/merchant-products/types"
import { renderWithProviders } from "@/test/render"

const product: Product = {
  id: 7,
  name: "Cotton T-shirt",
  description: "Soft cotton",
  status: "off_shelf",
  category_id: 3,
  category_name: "Clothing",
  sku_count: 2,
  total_stock: 15,
  min_price_cents: 1990,
  max_price_cents: 2590,
  created_at: "2026-11-15T08:00:00Z",
  updated_at: "2026-11-15T09:00:00Z",
}

const mocks = vi.hoisted(() => ({
  useProducts: vi.fn(),
  useProductCategories: vi.fn(),
  useCreateProduct: vi.fn(),
  useUpdateProduct: vi.fn(),
  useDeleteProduct: vi.fn(),
}))

vi.mock("@/features/merchant-products/model/queries", () => ({
  useProducts: mocks.useProducts,
  useProductCategories: mocks.useProductCategories,
}))

vi.mock("@/features/merchant-products/model/mutations", () => ({
  useCreateProduct: mocks.useCreateProduct,
  useUpdateProduct: mocks.useUpdateProduct,
  useDeleteProduct: mocks.useDeleteProduct,
}))

const deleteMutation = {
  error: null,
  isPending: false,
  mutate: vi.fn(),
}
const updateMutation = {
  error: null,
  isPending: false,
  mutate: vi.fn(),
  mutateAsync: vi.fn(),
  reset: vi.fn(),
}

function renderPage() {
  return renderWithProviders(
    <MerchantConsoleProductsPage
      token="token"
      productHref={(productID) => `/merchant/products/${productID}`}
      onCatalogChanged={() => Promise.resolve()}
    />,
  )
}

describe("MerchantConsoleProductsPage", () => {
  beforeEach(() => {
    mocks.useProducts.mockReturnValue({
      data: {
        items: [product],
        page: 1,
        page_size: 10,
        total: 1,
        total_pages: 1,
      },
      error: null,
      isPending: false,
      refetch: vi.fn(),
    })
    mocks.useProductCategories.mockReturnValue({
      data: [
        {
          id: 3,
          name: "Clothing",
          product_count: 1,
          created_at: product.created_at,
          updated_at: product.updated_at,
        },
      ],
      error: null,
      isPending: false,
      refetch: vi.fn(),
    })
    mocks.useCreateProduct.mockReturnValue({
      error: null,
      isPending: false,
      mutateAsync: vi.fn(),
      reset: vi.fn(),
    })
    mocks.useUpdateProduct.mockReturnValue(updateMutation)
    mocks.useDeleteProduct.mockReturnValue(deleteMutation)
  })

  it("renders the product list and sends search text to the list query", async () => {
    const user = userEvent.setup()
    renderPage()

    expect(screen.getByRole("link", { name: "Cotton T-shirt" })).toHaveAttribute(
      "href",
      "/merchant/products/7",
    )
    await user.type(
      screen.getByLabelText("Search products by name"),
      "shirt",
    )

    await waitFor(() => {
      expect(mocks.useProducts).toHaveBeenLastCalledWith(
        "token",
        expect.objectContaining({ search: "shirt", page: 1 }),
      )
    })
  })

  it("confirms deletion and invokes the product mutation", async () => {
    const user = userEvent.setup()
    renderPage()

    await user.click(screen.getByRole("button", { name: "Delete" }))
    const dialog = await screen.findByRole("alertdialog")
    expect(
      within(dialog).getByText(
        "Delete “Cotton T-shirt” and all of its SKUs. This cannot be undone.",
      ),
    ).toBeVisible()

    await user.click(within(dialog).getByRole("button", { name: "Delete" }))

    expect(deleteMutation.mutate).toHaveBeenCalledWith(
      7,
      expect.objectContaining({
        onSuccess: expect.any(Function),
        onSettled: expect.any(Function),
      }),
    )
  })

  it("confirms a shelf status change with the complete product payload", async () => {
    const user = userEvent.setup()
    renderPage()

    await user.click(screen.getByRole("button", { name: "Put on shelf" }))
    const dialog = await screen.findByRole("alertdialog")
    expect(
      within(dialog).getByText("Put “Cotton T-shirt” on shelf?"),
    ).toBeVisible()
    await user.click(
      within(dialog).getByRole("button", { name: "Put on shelf" }),
    )

    expect(updateMutation.mutate).toHaveBeenCalledWith(
      {
        productID: 7,
        input: {
          name: "Cotton T-shirt",
          description: "Soft cotton",
          category_id: 3,
          status: "on_shelf",
        },
      },
      expect.objectContaining({
        onSuccess: expect.any(Function),
        onSettled: expect.any(Function),
      }),
    )
  })

  it("opens the create-product form", async () => {
    const user = userEvent.setup()
    renderPage()

    await user.click(screen.getByRole("button", { name: "New product" }))

    expect(
      await screen.findByRole("heading", { name: "New product" }),
    ).toBeVisible()
    expect(screen.getByLabelText("Name")).toBeInTheDocument()
  })
})
