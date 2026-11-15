import { screen, waitFor, within } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { beforeEach, describe, expect, it, vi } from "vitest"
import MerchantConsoleCategoriesPage from "@/features/merchant-products/ui/categories-page"
import type { ProductCategory } from "@/features/merchant-products/types"
import { renderWithProviders } from "@/test/render"

const categories: ProductCategory[] = [
  {
    id: 3,
    name: "Clothing",
    product_count: 1,
    created_at: "2026-11-15T08:00:00Z",
    updated_at: "2026-11-15T09:00:00Z",
  },
  {
    id: 4,
    name: "Accessories",
    product_count: 0,
    created_at: "2026-11-15T08:00:00Z",
    updated_at: "2026-11-15T09:00:00Z",
  },
]

const mocks = vi.hoisted(() => ({
  useProductCategories: vi.fn(),
  useCreateProductCategory: vi.fn(),
  useUpdateProductCategory: vi.fn(),
  useDeleteProductCategory: vi.fn(),
}))

vi.mock("@/features/merchant-products/model/queries", () => ({
  useProductCategories: mocks.useProductCategories,
}))

vi.mock("@/features/merchant-products/model/mutations", () => ({
  useCreateProductCategory: mocks.useCreateProductCategory,
  useUpdateProductCategory: mocks.useUpdateProductCategory,
  useDeleteProductCategory: mocks.useDeleteProductCategory,
}))

const deleteMutation = {
  error: null,
  isPending: false,
  mutate: vi.fn(),
}
const createMutation = {
  error: null,
  isPending: false,
  mutateAsync: vi.fn().mockResolvedValue(undefined),
  reset: vi.fn(),
}
const updateMutation = {
  error: null,
  isPending: false,
  mutateAsync: vi.fn().mockResolvedValue(undefined),
  reset: vi.fn(),
}

function renderPage() {
  return renderWithProviders(
    <MerchantConsoleCategoriesPage
      token="token"
      onCatalogChanged={() => Promise.resolve()}
    />,
  )
}

describe("MerchantConsoleCategoriesPage", () => {
  beforeEach(() => {
    mocks.useProductCategories.mockReturnValue({
      data: categories,
      error: null,
      isPending: false,
      refetch: vi.fn(),
    })
    mocks.useCreateProductCategory.mockReturnValue(createMutation)
    mocks.useUpdateProductCategory.mockReturnValue(updateMutation)
    mocks.useDeleteProductCategory.mockReturnValue(deleteMutation)
  })

  it("filters categories locally by name", async () => {
    const user = userEvent.setup()
    renderPage()

    expect(screen.getByText("Clothing")).toBeVisible()
    expect(screen.getByText("Accessories")).toBeVisible()

    await user.type(
      screen.getByLabelText("Search categories by name"),
      "cloth",
    )

    await waitFor(() => {
      expect(screen.getByText("Clothing")).toBeVisible()
      expect(screen.queryByText("Accessories")).not.toBeInTheDocument()
    })
  })

  it("creates a category from the dialog", async () => {
    const user = userEvent.setup()
    renderPage()

    await user.click(screen.getByRole("button", { name: "New category" }))
    const dialog = await screen.findByRole("dialog")
    await user.type(
      within(dialog).getByLabelText("Category name"),
      "Home",
    )
    await user.click(within(dialog).getByRole("button", { name: "Create" }))

    expect(createMutation.mutateAsync).toHaveBeenCalledWith({ name: "Home" })
  })

  it("confirms category deletion", async () => {
    const user = userEvent.setup()
    renderPage()

    await user.click(screen.getAllByRole("button", { name: "Delete" })[0])
    const dialog = await screen.findByRole("alertdialog")
    expect(within(dialog).getByText(/Delete “Clothing”/)).toBeVisible()
    await user.click(within(dialog).getByRole("button", { name: "Delete" }))

    expect(deleteMutation.mutate).toHaveBeenCalledWith(
      3,
      expect.objectContaining({ onSettled: expect.any(Function) }),
    )
  })
})
