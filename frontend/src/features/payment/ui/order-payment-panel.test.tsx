import { screen, waitFor } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { HttpResponse, http } from "msw"
import { describe, expect, it, vi } from "vitest"
import { renderWithProviders } from "@/test/render"
import { server } from "@/test/server"
import { OrderPaymentPanel } from "./order-payment-panel"

const pendingPayment = {
  payment_no: "4f2614df-f674-4b5e-a9e4-3fc5e7f62013",
  order_id: 7,
  provider: "mockpay",
  provider_transaction_id: null,
  status: "pending",
  amount_cents: 3998,
  paid_at: null,
  created_at: "2026-11-24T12:00:00Z",
  updated_at: "2026-11-24T12:00:00Z",
}

const panelProps = {
  token: "token",
  orderID: 7,
  orderStatus: "pending_payment" as const,
  amountCents: 3998,
  expiresAt: "2026-11-24T12:30:00Z",
  paidAt: null,
}

describe("OrderPaymentPanel", () => {
  it("creates a payment and opens the provider checkout", async () => {
    const user = userEvent.setup()
    const onCheckout = vi.fn()
    const onPaymentSettled = vi.fn()
    const createBody = vi.fn<(body: unknown) => void>()
    const payPayment = vi.fn()

    server.use(
      http.post("*/api/v1/payments", async ({ request }) => {
        expect(request.headers.get("Authorization")).toBe("Bearer token")
        createBody(await request.json())
        return HttpResponse.json(pendingPayment, { status: 201 })
      }),
      http.post(
        "*/api/v1/payments/4f2614df-f674-4b5e-a9e4-3fc5e7f62013/pay",
        () => {
          payPayment()
          return HttpResponse.json({
            ...pendingPayment,
            provider_transaction_id: "8ad23f19-24f5-48b8-bbc7-243e1f1e0a98",
            checkout_url:
              "http://localhost:8080/mockpay/checkout/8ad23f19-24f5-48b8-bbc7-243e1f1e0a98",
          })
        },
      ),
    )

    renderWithProviders(
      <OrderPaymentPanel
        {...panelProps}
        onCheckout={onCheckout}
        onPaymentSettled={onPaymentSettled}
      />,
    )

    await user.click(screen.getByRole("button", { name: "Pay now" }))

    await waitFor(() => {
      expect(onCheckout).toHaveBeenCalledWith(
        "http://localhost:8080/mockpay/checkout/8ad23f19-24f5-48b8-bbc7-243e1f1e0a98",
      )
    })
    expect(createBody).toHaveBeenCalledWith({ order_id: 7 })
    expect(payPayment).toHaveBeenCalledOnce()
    expect(onPaymentSettled).not.toHaveBeenCalled()
  })

  it("marks an already-succeeded payment without creating checkout", async () => {
    const user = userEvent.setup()
    const onCheckout = vi.fn()
    const onPaymentSettled = vi.fn()
    const payPayment = vi.fn()

    server.use(
      http.post("*/api/v1/payments", () =>
        HttpResponse.json({
          ...pendingPayment,
          status: "succeeded",
          provider_transaction_id:
            "8ad23f19-24f5-48b8-bbc7-243e1f1e0a98",
          paid_at: "2026-11-24T12:05:00Z",
        }),
      ),
      http.post("*/api/v1/payments/:paymentNo/pay", () => {
        payPayment()
        return HttpResponse.json({})
      }),
    )

    renderWithProviders(
      <OrderPaymentPanel
        {...panelProps}
        onCheckout={onCheckout}
        onPaymentSettled={onPaymentSettled}
      />,
    )

    await user.click(screen.getByRole("button", { name: "Pay now" }))

    expect(await screen.findByText("Payment completed")).toBeInTheDocument()
    expect(onPaymentSettled).toHaveBeenCalledOnce()
    expect(onCheckout).not.toHaveBeenCalled()
    expect(payPayment).not.toHaveBeenCalled()
  })

  it("shows when an order is no longer payable", async () => {
    const user = userEvent.setup()

    server.use(
      http.post("*/api/v1/payments", () =>
        HttpResponse.json(
          { error: "order is not payable" },
          { status: 409 },
        ),
      ),
    )

    renderWithProviders(<OrderPaymentPanel {...panelProps} />)

    await user.click(screen.getByRole("button", { name: "Pay now" }))

    expect(
      await screen.findByText(
        "This order is no longer payable. Refresh for the latest status.",
      ),
    ).toBeInTheDocument()
  })

  it("shows a closed session error and dismisses it", async () => {
    const user = userEvent.setup()

    server.use(
      http.post("*/api/v1/payments", () =>
        HttpResponse.json(
          { ...pendingPayment, status: "closed" },
          { status: 200 },
        ),
      ),
    )

    renderWithProviders(<OrderPaymentPanel {...panelProps} />)

    await user.click(screen.getByRole("button", { name: "Pay now" }))
    expect(
      await screen.findByText("This payment session is closed."),
    ).toBeInTheDocument()
    await user.click(screen.getByRole("button", { name: "Dismiss" }))
    expect(
      screen.queryByText("This payment session is closed."),
    ).not.toBeInTheDocument()
  })

  it("shows when the provider returns no checkout URL", async () => {
    const user = userEvent.setup()

    server.use(
      http.post("*/api/v1/payments", () =>
        HttpResponse.json(pendingPayment, { status: 201 }),
      ),
      http.post("*/api/v1/payments/:paymentNo/pay", () =>
        HttpResponse.json({
          ...pendingPayment,
          provider_transaction_id:
            "8ad23f19-24f5-48b8-bbc7-243e1f1e0a98",
        }),
      ),
    )

    renderWithProviders(<OrderPaymentPanel {...panelProps} />)

    await user.click(screen.getByRole("button", { name: "Pay now" }))
    expect(
      await screen.findByText(
        "The checkout session is unavailable. Try again.",
      ),
    ).toBeInTheDocument()
  })

  it("marks payment settled when close discovers a paid payment", async () => {
    const user = userEvent.setup()
    const onPaymentSettled = vi.fn()

    server.use(
      http.post("*/api/v1/payments", () =>
        HttpResponse.json(pendingPayment, { status: 201 }),
      ),
      http.post("*/api/v1/payments/:paymentNo/close", () =>
        HttpResponse.json({
          ...pendingPayment,
          status: "succeeded",
          provider_transaction_id:
            "8ad23f19-24f5-48b8-bbc7-243e1f1e0a98",
          paid_at: "2026-11-24T12:05:00Z",
        }),
      ),
    )

    renderWithProviders(
      <OrderPaymentPanel
        {...panelProps}
        onPaymentSettled={onPaymentSettled}
      />,
    )

    await user.click(
      screen.getByRole("button", {
        name: "Close payment and cancel order",
      }),
    )
    await user.click(
      screen.getByRole("button", { name: "Close and cancel order" }),
    )

    expect(await screen.findByText("Payment completed")).toBeVisible()
    expect(onPaymentSettled).toHaveBeenCalledOnce()
  })

  it("shows provider and close failures", async () => {
    const user = userEvent.setup()
    const handlers = [
      http.post("*/api/v1/payments", () =>
        HttpResponse.json({ error: "unavailable" }, { status: 502 }),
      ),
    ]
    server.use(...handlers)

    const { unmount } = renderWithProviders(
      <OrderPaymentPanel {...panelProps} />,
    )
    await user.click(screen.getByRole("button", { name: "Pay now" }))
    expect(
      await screen.findByText(
        "The payment service is temporarily unavailable. Try again later.",
      ),
    ).toBeVisible()
    unmount()

    server.use(
      http.post("*/api/v1/payments", () =>
        HttpResponse.json(pendingPayment, { status: 201 }),
      ),
      http.post("*/api/v1/payments/:paymentNo/close", () =>
        HttpResponse.json({ error: "unavailable" }, { status: 500 }),
      ),
    )
    renderWithProviders(<OrderPaymentPanel {...panelProps} />)
    await user.click(
      screen.getByRole("button", {
        name: "Close payment and cancel order",
      }),
    )
    await user.click(
      screen.getByRole("button", { name: "Close and cancel order" }),
    )

    expect(
      await screen.findByText("Could not close payment"),
    ).toBeVisible()
    expect(
      await screen.findByText(
        "The server cannot start payment right now. Try again later.",
      ),
    ).toBeVisible()
  })

  it("renders a paid order directly", () => {
    renderWithProviders(
      <OrderPaymentPanel
        {...panelProps}
        orderStatus="paid"
        paidAt="2026-11-24T12:05:00Z"
      />,
    )

    expect(screen.getByText("Payment completed")).toBeVisible()
    expect(screen.getByText(/Paid /)).toBeVisible()
  })

  it("closes payment and notifies when the order is cancelled", async () => {
    const user = userEvent.setup()
    const onOrderCancelled = vi.fn()
    const createPayment = vi.fn()
    const closePayment = vi.fn()

    server.use(
      http.post("*/api/v1/payments", () => {
        createPayment()
        return HttpResponse.json(pendingPayment, { status: 201 })
      }),
      http.post(
        "*/api/v1/payments/4f2614df-f674-4b5e-a9e4-3fc5e7f62013/close",
        () => {
          closePayment()
          return HttpResponse.json({
            ...pendingPayment,
            status: "closed",
          })
        },
      ),
    )

    renderWithProviders(
      <OrderPaymentPanel
        {...panelProps}
        onOrderCancelled={onOrderCancelled}
      />,
    )

    await user.click(
      screen.getByRole("button", {
        name: "Close payment and cancel order",
      }),
    )
    expect(await screen.findByText("Close this payment?")).toBeInTheDocument()
    await user.click(
      screen.getByRole("button", { name: "Close and cancel order" }),
    )

    await waitFor(() => {
      expect(onOrderCancelled).toHaveBeenCalledOnce()
    })
    expect(createPayment).toHaveBeenCalledOnce()
    expect(closePayment).toHaveBeenCalledOnce()
    expect(await screen.findByText("Order cancelled")).toBeInTheDocument()
  })

  it("shows when a payment can no longer be closed", async () => {
    const user = userEvent.setup()

    server.use(
      http.post("*/api/v1/payments", () =>
        HttpResponse.json(pendingPayment, { status: 201 }),
      ),
      http.post("*/api/v1/payments/:paymentNo/close", () =>
        HttpResponse.json(
          { error: "payment conflicts with the existing transaction" },
          { status: 409 },
        ),
      ),
    )

    renderWithProviders(<OrderPaymentPanel {...panelProps} />)

    await user.click(
      screen.getByRole("button", {
        name: "Close payment and cancel order",
      }),
    )
    await user.click(
      screen.getByRole("button", { name: "Close and cancel order" }),
    )

    expect(
      await screen.findByText(
        "This payment can no longer be closed. Refresh for the latest status.",
      ),
    ).toBeInTheDocument()
  })
})
