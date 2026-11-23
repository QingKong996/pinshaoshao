import type { ReactNode } from "react"
import { QueryClientProvider } from "@tanstack/react-query"
import { renderHook, waitFor } from "@testing-library/react"
import { HttpResponse, http } from "msw"
import { describe, expect, it } from "vitest"
import { createTestQueryClient } from "@/test/render"
import { server } from "@/test/server"
import { useOrder, useOrders } from "./queries"

const order = {
  id: 9,
  status: "pending_payment",
  total_amount_cents: 3998,
  receiver_name: "Receiver",
  receiver_phone: "1234567890",
  shipping_address: "Shanghai",
  expires_at: "2026-11-20T12:30:00Z",
  paid_at: null,
  created_at: "2026-11-20T12:00:00Z",
  updated_at: "2026-11-20T12:00:00Z",
  sub_orders: [],
}

describe("order queries", () => {
  it("loads a paginated order list", async () => {
    server.use(
      http.get("*/api/v1/orders", ({ request }) => {
        const url = new URL(request.url)
        expect(url.searchParams.get("page")).toBe("2")
        expect(url.searchParams.get("page_size")).toBe("5")
        return HttpResponse.json({
          items: [],
          page: 2,
          page_size: 5,
          total: 5,
          total_pages: 2,
        })
      }),
    )

    const { result } = renderHook(() => useOrders("token", 2, 5), {
      wrapper: createWrapper(),
    })

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true)
    })
    expect(result.current.data).toMatchObject({
      page: 2,
      page_size: 5,
      total: 5,
    })
  })

  it("loads a single order", async () => {
    server.use(
      http.get("*/api/v1/orders/9", () => HttpResponse.json(order)),
    )

    const { result } = renderHook(() => useOrder("token", 9), {
      wrapper: createWrapper(),
    })

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true)
    })
    expect(result.current.data).toEqual(order)
  })

  it("does not request invalid orders", () => {
    const { result } = renderHook(
      () => ({
        noToken: useOrder("", 9),
        invalidID: useOrder("token", 0),
      }),
      { wrapper: createWrapper() },
    )

    expect(result.current.noToken.fetchStatus).toBe("idle")
    expect(result.current.invalidID.fetchStatus).toBe("idle")
  })
})

function createWrapper() {
  const queryClient = createTestQueryClient()

  return function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        {children}
      </QueryClientProvider>
    )
  }
}
