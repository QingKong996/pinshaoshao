import type { ReactNode } from "react"
import { QueryClientProvider } from "@tanstack/react-query"
import { act, renderHook, waitFor } from "@testing-library/react"
import { HttpResponse, http } from "msw"
import { describe, expect, it } from "vitest"
import { createTestQueryClient } from "@/test/render"
import { server } from "@/test/server"
import { orderQueryKeys } from "./query-keys"
import { useCancelOrder } from "./mutations"

const pendingOrder = {
  id: 3,
  status: "pending_payment",
  total_amount_cents: 1999,
  receiver_name: "Receiver",
  receiver_phone: "1234567890",
  shipping_address: "Shanghai",
  expires_at: "2026-11-20T12:30:00Z",
  paid_at: null,
  created_at: "2026-11-20T12:00:00Z",
  updated_at: "2026-11-20T12:00:00Z",
  sub_orders: [],
}

describe("order mutations", () => {
  it("stores the cancelled order and invalidates the order list", async () => {
    const queryClient = createTestQueryClient({ gcTime: 60_000 })
    queryClient.setQueryData(
      orderQueryKeys.order("token", pendingOrder.id),
      pendingOrder,
    )
    queryClient.setQueryData(orderQueryKeys.orders("token"), {
      pages: [],
      pageParams: [],
    })
    server.use(
      http.post("*/api/v1/orders/3/cancel", () =>
        HttpResponse.json({
          ...pendingOrder,
          status: "cancelled",
        }),
      ),
    )

    const { result } = renderHook(
      () => useCancelOrder("token", pendingOrder.id),
      { wrapper: createWrapper(queryClient) },
    )

    act(() => {
      result.current.mutate()
    })

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true)
    })
    expect(
      queryClient.getQueryData(
        orderQueryKeys.order("token", pendingOrder.id),
      ),
    ).toMatchObject({ id: 3, status: "cancelled" })
    await waitFor(() => {
      expect(
        queryClient.getQueryState(
          orderQueryKeys.order("token", pendingOrder.id),
        )?.isInvalidated,
      ).toBe(true)
    })
    await waitFor(() => {
      expect(
        queryClient.getQueryState(orderQueryKeys.orders("token"))
          ?.isInvalidated,
      ).toBe(true)
    })
  })
})

function createWrapper(
  queryClient: ReturnType<typeof createTestQueryClient>,
) {
  return function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        {children}
      </QueryClientProvider>
    )
  }
}
