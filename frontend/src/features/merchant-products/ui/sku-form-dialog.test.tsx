import { fireEvent, render, screen, waitFor } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { beforeEach, describe, expect, it, vi } from "vitest"
import { SKUFormDialog } from "@/features/merchant-products/ui/sku-form-dialog"
import type { ProductSKU } from "@/features/merchant-products/types"

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

vi.mock("@/features/merchant-products/model/mutations", () => ({
  useCreateProductSKU: () => createMutation,
  useUpdateProductSKU: () => updateMutation,
}))

const sku: ProductSKU = {
  id: 11,
  product_id: 7,
  specs: { Size: "M" },
  price_cents: 1990,
  is_enabled: true,
  created_at: "2026-11-15T08:00:00Z",
  updated_at: "2026-11-15T09:00:00Z",
}

describe("SKUFormDialog", () => {
  beforeEach(() => {
    createMutation.mutateAsync.mockClear()
    updateMutation.mutateAsync.mockClear()
  })

  it("creates a SKU with normalized specs and price", async () => {
    const user = userEvent.setup()
    const onOpenChange = vi.fn()
    render(
      <SKUFormDialog
        token="token"
        onCatalogChanged={() => Promise.resolve()}
        open
        onOpenChange={onOpenChange}
        productID={7}
        skus={[]}
      />,
    )

    expect(
      screen.getByRole("heading", { name: "Add SKU" }),
    ).toBeVisible()
    await user.type(screen.getByLabelText("Spec name"), " Size ")
    await user.type(screen.getByLabelText("Spec value"), " M ")
    await user.type(screen.getByLabelText("Price (CNY)"), "19.90")
    await user.click(screen.getByRole("button", { name: "Add" }))

    await waitFor(() =>
      expect(createMutation.mutateAsync).toHaveBeenCalledWith({
        productID: 7,
        input: {
          specs: { Size: "M" },
          price_cents: 1990,
          is_enabled: true,
        },
      }),
    )
    await waitFor(() => expect(onOpenChange).toHaveBeenCalledWith(false))
  })

  it("updates an existing SKU and submits the changed values", async () => {
    const user = userEvent.setup()
    render(
      <SKUFormDialog
        token="token"
        onCatalogChanged={() => Promise.resolve()}
        open
        onOpenChange={vi.fn()}
        productID={7}
        sku={sku}
        skus={[sku]}
      />,
    )

    expect(screen.getByRole("heading", { name: "Edit SKU" })).toBeVisible()
    expect(screen.getByLabelText("Spec name")).toHaveValue("Size")
    const priceInput = screen.getByLabelText("Price (CNY)")
    fireEvent.change(priceInput, { target: { value: "25.50" } })
    expect(priceInput).toHaveValue("25.50")
    await user.click(screen.getByRole("button", { name: "Save" }))

    await waitFor(() =>
      expect(updateMutation.mutateAsync).toHaveBeenCalledWith({
        productID: 7,
        skuID: 11,
        input: {
          specs: { Size: "M" },
          price_cents: 2550,
          is_enabled: true,
        },
      }),
    )
  })
})
