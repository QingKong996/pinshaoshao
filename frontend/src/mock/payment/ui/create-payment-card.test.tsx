import { Route, Routes } from "react-router"
import { screen, waitFor } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { HttpResponse, http } from "msw"
import { describe, expect, it } from "vitest"
import { renderWithProviders } from "@/test/render"
import { server } from "@/test/server"
import type { CreateMockPaymentInput } from "../dto/payment"
import { CreatePaymentCard } from "./create-payment-card"

describe("CreatePaymentCard", () => {
  it("posts form values and opens the created payment", async () => {
    const user = userEvent.setup()
    let posted: CreateMockPaymentInput | undefined
    server.use(
      http.post("*/mockpay/api/v1/payments", async ({ request }) => {
        posted = (await request.json()) as CreateMockPaymentInput
        return HttpResponse.json(
          {
            payment_no: "4f2614df-f674-4b5e-a9e4-3fc5e7f62013",
            merchant_order_no: posted.merchant_order_no,
            status: "pending",
            amount_cents: posted.amount_cents,
            currency: posted.currency,
            subject: posted.subject,
            checkout_url:
              "http://localhost:8080/mockpay/checkout/4f2614df-f674-4b5e-a9e4-3fc5e7f62013",
            expires_at: "2026-11-24T12:30:00Z",
            paid_at: null,
            closed_at: null,
            created_at: "2026-11-24T12:00:00Z",
          },
          { status: 201 },
        )
      }),
    )

    renderWithProviders(
      <Routes>
        <Route path="/mockpay" element={<CreatePaymentCard />} />
        <Route
          path="/mockpay/payments/:paymentNo"
          element={<div data-testid="payment-detail-route" />}
        />
      </Routes>,
      { initialEntries: ["/mockpay"] },
    )

    await user.clear(screen.getByLabelText("Merchant order number"))
    await user.type(screen.getByLabelText("Merchant order number"), "order-ui-1")
    await user.type(screen.getByLabelText("Amount (cents)"), "3998")
    await user.click(screen.getByRole("button", { name: "Create payment" }))

    await waitFor(() => {
      expect(posted).toBeDefined()
    })
    expect(posted).toMatchObject({
      merchant_order_no: "order-ui-1",
      amount_cents: 3998,
      currency: "CNY",
      subject: "Sandbox order",
      notify_url: "http://localhost:8080/api/v1/payment/webhook",
      expires_in_seconds: 1800,
    })
    expect(posted).not.toHaveProperty("return_url")
    expect(await screen.findByTestId("payment-detail-route")).toBeInTheDocument()
  })
})
