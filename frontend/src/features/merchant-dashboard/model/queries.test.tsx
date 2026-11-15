import type { ReactNode } from "react"
import { QueryClientProvider } from "@tanstack/react-query"
import { renderHook, waitFor } from "@testing-library/react"
import { HttpResponse, http } from "msw"
import { describe, expect, it, vi } from "vitest"
import {
  invalidateMerchantDashboard,
  useDashboard,
  useShopProfile,
} from "@/features/merchant-dashboard/model/queries"
import { dashboardQueryKeys } from "@/features/merchant-dashboard/model/query-keys"
import { createTestQueryClient } from "@/test/render"
import { server } from "@/test/server"

describe("merchant dashboard queries", () => {
  it("loads dashboard and shop profile data", async () => {
    server.use(
      http.get("*/api/v1/merchant/dashboard", () =>
        HttpResponse.json({
          product_count: 3,
          on_shelf_product_count: 2,
          sku_count: 6,
          category_count: 2,
          total_stock: 20,
          inventory_value_cents: 39900,
          out_of_stock_sku_count: 1,
        }),
      ),
      http.get("*/api/v1/merchant/shop", () =>
        HttpResponse.json({
          id: 1,
          username: "merchant",
          name: "Example Shop",
        }),
      ),
    )
    const queryClient = createTestQueryClient()
    function Wrapper({ children }: { children: ReactNode }) {
      return (
        <QueryClientProvider client={queryClient}>
          {children}
        </QueryClientProvider>
      )
    }

    const dashboard = renderHook(() => useDashboard("token"), {
      wrapper: Wrapper,
    })
    const shop = renderHook(() => useShopProfile("token"), {
      wrapper: Wrapper,
    })

    await waitFor(() => {
      expect(dashboard.result.current.isSuccess).toBe(true)
      expect(shop.result.current.isSuccess).toBe(true)
    })
    expect(dashboard.result.current.data?.product_count).toBe(3)
    expect(shop.result.current.data?.name).toBe("Example Shop")
  })

  it("invalidates only the dashboard query", async () => {
    const queryClient = createTestQueryClient()
    const invalidate = vi
      .spyOn(queryClient, "invalidateQueries")
      .mockResolvedValue()

    await invalidateMerchantDashboard(queryClient)

    expect(invalidate).toHaveBeenCalledWith({
      queryKey: dashboardQueryKeys.dashboard(),
    })
  })
})
