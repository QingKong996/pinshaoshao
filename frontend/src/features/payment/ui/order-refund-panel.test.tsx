import { screen } from "@testing-library/react"
import { HttpResponse, http } from "msw"
import { describe, expect, it } from "vitest"
import { renderWithProviders } from "@/test/render"
import { server } from "@/test/server"
import { OrderRefundPanel } from "./order-refund-panel"

const basePayment = {
  payment_no: "4f2614df-f674-4b5e-a9e4-3fc5e7f62013",
  order_id: 7,
  provider: "mockpay",
  provider_transaction_id: "8ad23f19-24f5-48b8-bbc7-243e1f1e0a98",
  amount_cents: 3998,
  paid_at: "2026-11-24T12:05:00Z",
  created_at: "2026-11-24T12:00:00Z",
  updated_at: "2026-11-24T12:05:00Z",
}

describe("OrderRefundPanel", () => {
  it.each([
    ["pending", "Refund in progress"],
    ["succeeded", "Refunded"],
  ] as const)("renders a %s refund", async (status, label) => {
    server.use(
      http.get("*/api/v1/orders/7/payment", () =>
        HttpResponse.json({
          ...basePayment,
          status: "succeeded",
          refund: {
            refund_no: "a1c6cfe8-e77b-4ac2-b8bf-c6a0f56673f1",
            amount_cents: 3998,
            reason: "order_cancelled_after_payment",
            status,
            provider_refund_no:
              status === "succeeded"
                ? "b6b2d7f3-6f28-4f24-a9df-6a1d3b7cfb19"
                : null,
            refunded_at:
              status === "succeeded" ? "2026-11-24T12:10:00Z" : null,
            failed_at: null,
            created_at: "2026-11-24T12:05:00Z",
            updated_at: "2026-11-24T12:10:00Z",
          },
        }),
      ),
    )

    renderWithProviders(<OrderRefundPanel token="token" orderID={7} />)

    expect(await screen.findByText("Refund")).toBeVisible()
    expect(screen.getByText(label)).toBeVisible()
    expect(screen.getByText("Refund amount")).toBeVisible()
  })
})
