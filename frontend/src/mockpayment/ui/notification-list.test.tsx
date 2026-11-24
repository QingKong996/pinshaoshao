import { screen, waitFor } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { HttpResponse, http } from "msw"
import { describe, expect, it } from "vitest"
import { renderWithProviders } from "@/test/render"
import { server } from "@/test/server"
import { NotificationList } from "./notification-list"

const paymentNo = "4f2614df-f674-4b5e-a9e4-3fc5e7f62013"

const exhaustedNotification = {
  id: 1,
  event_id: "a1c6cfe8-e77b-4ac2-b8bf-c6a0f56673f1",
  event_type: "payment.paid",
  target_url: "https://example.com/webhook",
  status: "exhausted",
  attempt_count: 3,
  max_attempts: 3,
  next_attempt_at: "2026-11-24T12:00:00Z",
  last_attempt_at: "2026-11-24T12:00:00Z",
  delivered_at: null,
  last_error: "webhook returned 503",
  created_at: "2026-11-24T12:00:00Z",
  updated_at: "2026-11-24T12:00:00Z",
}

const pendingNotification = {
  ...exhaustedNotification,
  status: "pending",
  attempt_count: 0,
  last_attempt_at: null,
  last_error: null,
}

describe("NotificationList", () => {
  it("retries only exhausted notifications and refreshes the list", async () => {
    const user = userEvent.setup()
    let retryCalls = 0
    let retried = false

    server.use(
      http.get(
        `*/mockpay/api/v1/payments/${paymentNo}/notifications`,
        () =>
          HttpResponse.json({
            items: [retried ? pendingNotification : exhaustedNotification],
          }),
      ),
      http.post(
        `*/mockpay/api/v1/payments/${paymentNo}/notifications/retry`,
        () => {
          retried = true
          retryCalls += 1
          return HttpResponse.json(pendingNotification)
        },
      ),
    )

    renderWithProviders(
      <NotificationList paymentNo={paymentNo} paymentStatus="paid" />,
    )

    expect(await screen.findByText("Exhausted")).toBeInTheDocument()
    expect(screen.getByText("webhook returned 503")).toBeInTheDocument()

    await user.click(screen.getByRole("button", { name: "Retry" }))

    await waitFor(() => {
      expect(retryCalls).toBe(1)
    })
    expect(await screen.findByText("Pending")).toBeInTheDocument()
    expect(
      screen.queryByRole("button", { name: "Retry" }),
    ).not.toBeInTheDocument()
  })
})
