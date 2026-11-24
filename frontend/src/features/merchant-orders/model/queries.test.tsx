import type { ReactNode } from "react"
import { QueryClientProvider } from "@tanstack/react-query"
import { renderHook, waitFor } from "@testing-library/react"
import { HttpResponse, http } from "msw"
import { describe, expect, it } from "vitest"
import { useMerchantOrders } from "@/features/merchant-orders/model/queries"
import type { MerchantOrderListResult } from "@/features/merchant-orders/types"
import { createTestQueryClient } from "@/test/render"
import { server } from "@/test/server"

const orderList: MerchantOrderListResult = {
  items: [
    {
      id: 4,
      order_id: 12,
      status: "shipped",
      total_amount_cents: 3998,
      item_count: 2,
      receiver_name: "Receiver",
      receiver_phone: "1234567890",
      shipping_address: "Shanghai",
      paid_at: "2026-11-20T12:05:00Z",
      created_at: "2026-11-20T12:00:00Z",
    },
  ],
  page: 2,
  page_size: 5,
  total: 6,
  total_pages: 2,
}

describe("merchant order queries", () => {
  it("loads the authenticated order list with filters", async () => {
    let search: Record<string, string> = {}
    let authorization: string | null = null

    server.use(
      http.get("*/api/v1/merchant/orders", ({ request }) => {
        const url = new URL(request.url)
        search = Object.fromEntries(url.searchParams)
        authorization = request.headers.get("Authorization")
        return HttpResponse.json(orderList)
      }),
    )

    const queryClient = createTestQueryClient()
    function Wrapper({ children }: { children: ReactNode }) {
      return (
        <QueryClientProvider client={queryClient}>
          {children}
        </QueryClientProvider>
      )
    }

    const orders = renderHook(
      () =>
        useMerchantOrders("merchant-token", {
          page: 2,
          page_size: 5,
          status: "shipped",
        }),
      { wrapper: Wrapper },
    )

    await waitFor(() => {
      expect(orders.result.current.isSuccess).toBe(true)
    })
    expect(search).toEqual({
      page: "2",
      page_size: "5",
      status: "shipped",
    })
    expect(authorization).toBe("Bearer merchant-token")
    expect(orders.result.current.data).toEqual(orderList)
  })
})
