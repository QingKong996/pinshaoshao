import type { ReactNode } from "react"
import { QueryClientProvider } from "@tanstack/react-query"
import { act, renderHook } from "@testing-library/react"
import { HttpResponse, http } from "msw"
import { describe, expect, it, vi } from "vitest"
import {
  useCreateProduct,
  useCreateProductCategory,
  useCreateProductSKU,
  useDeleteProduct,
  useDeleteProductCategory,
  useDeleteProductSKU,
  useUpdateInventoryStock,
  useUpdateProduct,
  useUpdateProductCategory,
  useUpdateProductSKU,
} from "@/features/merchant-products/model/mutations"
import { productQueryKeys } from "@/features/merchant-products/model/query-keys"
import { createTestQueryClient } from "@/test/render"
import { server } from "@/test/server"

const authToken = ["merchant", "token"].join("-")
const productInput = {
  name: "Cotton T-shirt",
  description: "Soft cotton",
  category_id: 3,
  status: "off_shelf" as const,
}
const skuInput = {
  specs: { Size: "M" },
  price_cents: 1990,
  is_enabled: true,
}

function setup() {
  const queryClient = createTestQueryClient()
  const invalidateQueries = vi
    .spyOn(queryClient, "invalidateQueries")
    .mockResolvedValue()
  const onCatalogChanged = vi.fn().mockResolvedValue(undefined)
  function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        {children}
      </QueryClientProvider>
    )
  }
  return { Wrapper, invalidateQueries, onCatalogChanged }
}

describe("merchant product mutations", () => {
  it("invalidates product, category, and inventory data for product writes", async () => {
    server.use(
      http.all("*/api/v1/merchant/*", ({ request }) =>
        request.method === "DELETE"
          ? new HttpResponse(null, { status: 204 })
          : HttpResponse.json({ id: 7 }),
      ),
    )
    const { Wrapper, invalidateQueries, onCatalogChanged } = setup()
    const create = renderHook(
      () => useCreateProduct(authToken, onCatalogChanged),
      { wrapper: Wrapper },
    )
    const update = renderHook(
      () => useUpdateProduct(authToken, onCatalogChanged),
      { wrapper: Wrapper },
    )
    const remove = renderHook(
      () => useDeleteProduct(authToken, onCatalogChanged),
      { wrapper: Wrapper },
    )

    await act(async () => {
      await create.result.current.mutateAsync(productInput)
      await update.result.current.mutateAsync({
        productID: 7,
        input: productInput,
      })
      await remove.result.current.mutateAsync(7)
    })

    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: productQueryKeys.products(),
    })
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: productQueryKeys.categories(),
    })
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: productQueryKeys.inventory(),
    })
    expect(onCatalogChanged).toHaveBeenCalledTimes(3)
  })

  it("invalidates product and inventory data for SKU writes", async () => {
    server.use(
      http.all("*/api/v1/merchant/*", ({ request }) =>
        request.method === "DELETE"
          ? new HttpResponse(null, { status: 204 })
          : HttpResponse.json({ id: 11 }),
      ),
    )
    const { Wrapper, invalidateQueries, onCatalogChanged } = setup()
    const create = renderHook(
      () => useCreateProductSKU(authToken, onCatalogChanged),
      { wrapper: Wrapper },
    )
    const update = renderHook(
      () => useUpdateProductSKU(authToken, onCatalogChanged),
      { wrapper: Wrapper },
    )
    const remove = renderHook(
      () => useDeleteProductSKU(authToken, onCatalogChanged),
      { wrapper: Wrapper },
    )

    await act(async () => {
      await create.result.current.mutateAsync({
        productID: 7,
        input: skuInput,
      })
      await update.result.current.mutateAsync({
        productID: 7,
        skuID: 11,
        input: skuInput,
      })
      await remove.result.current.mutateAsync({
        productID: 7,
        skuID: 11,
      })
    })

    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: productQueryKeys.products(),
    })
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: productQueryKeys.inventory(),
    })
    expect(invalidateQueries).not.toHaveBeenCalledWith({
      queryKey: productQueryKeys.categories(),
    })
    expect(onCatalogChanged).toHaveBeenCalledTimes(3)
  })

  it("invalidates catalog data for category writes", async () => {
    server.use(
      http.all("*/api/v1/merchant/*", ({ request }) =>
        request.method === "DELETE"
          ? new HttpResponse(null, { status: 204 })
          : HttpResponse.json({ id: 3 }),
      ),
    )
    const { Wrapper, invalidateQueries, onCatalogChanged } = setup()
    const create = renderHook(
      () => useCreateProductCategory(authToken, onCatalogChanged),
      { wrapper: Wrapper },
    )
    const update = renderHook(
      () => useUpdateProductCategory(authToken, onCatalogChanged),
      { wrapper: Wrapper },
    )
    const remove = renderHook(
      () => useDeleteProductCategory(authToken, onCatalogChanged),
      { wrapper: Wrapper },
    )

    await act(async () => {
      await create.result.current.mutateAsync({ name: "Clothing" })
      await update.result.current.mutateAsync({
        categoryID: 3,
        input: { name: "Apparel" },
      })
      await remove.result.current.mutateAsync(3)
    })

    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: productQueryKeys.products(),
    })
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: productQueryKeys.categories(),
    })
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: productQueryKeys.inventory(),
    })
    expect(onCatalogChanged).toHaveBeenCalledTimes(3)
  })

  it("invalidates inventory and products after a stock adjustment", async () => {
    server.use(
      http.patch("*/api/v1/merchant/inventory/11", () =>
        HttpResponse.json({ id: 11 }),
      ),
    )
    const { Wrapper, invalidateQueries, onCatalogChanged } = setup()
    const update = renderHook(
      () => useUpdateInventoryStock(authToken, onCatalogChanged),
      { wrapper: Wrapper },
    )

    await act(async () => {
      await update.result.current.mutateAsync({
        skuID: 11,
        totalStock: 9,
        version: 2,
      })
    })

    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: productQueryKeys.inventory(),
    })
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: productQueryKeys.products(),
    })
    expect(onCatalogChanged).toHaveBeenCalledOnce()
  })
})
