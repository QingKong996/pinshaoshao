import type { ReactNode } from "react"
import { QueryClientProvider } from "@tanstack/react-query"
import { renderHook, waitFor } from "@testing-library/react"
import { HttpResponse, http } from "msw"
import { describe, expect, it, vi } from "vitest"
import {
  useInventory,
  useProduct,
  useProductCategories,
  useProductSKUs,
  useProducts,
} from "@/features/merchant-products/model/queries"
import { createTestQueryClient } from "@/test/render"
import { server } from "@/test/server"

function createWrapper() {
  const queryClient = createTestQueryClient()
  function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        {children}
      </QueryClientProvider>
    )
  }
  return { queryClient, Wrapper }
}

describe("merchant product queries", () => {
  it("loads products, detail, categories, SKUs, and inventory", async () => {
    server.use(
      http.get("*/api/v1/merchant/products", () =>
        HttpResponse.json({
          items: [],
          page: 1,
          page_size: 10,
          total: 0,
          total_pages: 0,
        }),
      ),
      http.get("*/api/v1/merchant/products/7/skus", () =>
        HttpResponse.json([]),
      ),
      http.get("*/api/v1/merchant/products/7", () =>
        HttpResponse.json({ id: 7, name: "Cotton T-shirt" }),
      ),
      http.get("*/api/v1/merchant/categories", () => HttpResponse.json([])),
      http.get("*/api/v1/merchant/inventory", () => HttpResponse.json([])),
    )

    const products = renderHook(
      () =>
        useProducts("token", {
          page: 1,
          page_size: 10,
          sort: "created_at",
          order: "desc",
        }),
      { wrapper: createWrapper().Wrapper },
    )
    const product = renderHook(() => useProduct("token", 7), {
      wrapper: createWrapper().Wrapper,
    })
    const categories = renderHook(() => useProductCategories("token"), {
      wrapper: createWrapper().Wrapper,
    })
    const skus = renderHook(() => useProductSKUs("token", 7), {
      wrapper: createWrapper().Wrapper,
    })
    const inventory = renderHook(() => useInventory("token"), {
      wrapper: createWrapper().Wrapper,
    })

    await waitFor(() => {
      expect(products.result.current.isSuccess).toBe(true)
      expect(product.result.current.isSuccess).toBe(true)
      expect(categories.result.current.isSuccess).toBe(true)
      expect(skus.result.current.isSuccess).toBe(true)
      expect(inventory.result.current.isSuccess).toBe(true)
    })
    expect(products.result.current.data?.total).toBe(0)
    expect(product.result.current.data).toMatchObject({ id: 7 })
  })

  it("does not fetch disabled or invalid resources", async () => {
    const request = vi.fn()
    server.use(
      http.get("*/api/v1/merchant/products/0", () => {
        request()
        return HttpResponse.json({})
      }),
      http.get("*/api/v1/merchant/products/1/skus", () => {
        request()
        return HttpResponse.json([])
      }),
      http.get("*/api/v1/merchant/categories", () => {
        request()
        return HttpResponse.json([])
      }),
    )

    const product = renderHook(() => useProduct("token", 0), {
      wrapper: createWrapper().Wrapper,
    })
    const skus = renderHook(() => useProductSKUs("token", 1.5), {
      wrapper: createWrapper().Wrapper,
    })
    const categories = renderHook(
      () => useProductCategories("token", false),
      { wrapper: createWrapper().Wrapper },
    )

    await new Promise((resolve) => setTimeout(resolve, 20))
    expect(product.result.current.fetchStatus).toBe("idle")
    expect(skus.result.current.fetchStatus).toBe("idle")
    expect(categories.result.current.fetchStatus).toBe("idle")
    expect(request).not.toHaveBeenCalled()
  })
})
