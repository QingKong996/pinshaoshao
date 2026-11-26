import { expect, test, type Page } from "@playwright/test"

const orderID = 7
const paymentNo = "4f2614df-f674-4b5e-a9e4-3fc5e7f62013"
const providerNo = "8ad23f19-24f5-48b8-bbc7-243e1f1e0a98"
const checkoutURL = `http://127.0.0.1:8080/mockpay/checkout/${providerNo}`

type OrderStatus = "pending_payment" | "paid" | "cancelled"

function orderResponse(status: OrderStatus) {
  const paid = status === "paid"
  const cancelled = status === "cancelled"
  const subOrderStatus = cancelled
    ? "cancelled"
    : paid
      ? "awaiting_shipment"
      : "pending_payment"
  return {
    id: orderID,
    status,
    total_amount_cents: 3998,
    receiver_name: "E2E Buyer",
    receiver_phone: "13800000000",
    shipping_address: "E2E Street 1",
    expires_at: "2026-11-24T12:30:00Z",
    paid_at: paid ? "2026-11-24T12:05:00Z" : null,
    created_at: "2026-11-24T12:00:00Z",
    updated_at: "2026-11-24T12:05:00Z",
    sub_orders: [
      {
        id: 11,
        shop_id: 3,
        shop_name: "E2E Shop",
        status: subOrderStatus,
        total_amount_cents: 3998,
        items: [
          {
            id: 21,
            product_id: 10,
            sku_id: 5,
            product_name: "Cotton T-shirt",
            sku_specs: { Size: "M" },
            unit_price_cents: 1999,
            quantity: 2,
            line_amount_cents: 3998,
          },
        ],
      },
    ],
  }
}

function pendingPayment() {
  return {
    payment_no: paymentNo,
    order_id: orderID,
    provider: "mockpay",
    provider_transaction_id: null,
    status: "pending",
    amount_cents: 3998,
    paid_at: null,
    created_at: "2026-11-24T12:00:00Z",
    updated_at: "2026-11-24T12:00:00Z",
  }
}

async function mockPaymentFlow(page: Page) {
  let orderStatus: OrderStatus = "pending_payment"
  let created = false
  let sessionStarted = false
  let cancelled = false

  await page.route("**/api/v1/**", async (route) => {
    const request = route.request()
    const pathname = new URL(request.url()).pathname

    if (request.method() === "POST" && pathname.endsWith("/login")) {
      await route.fulfill({ json: { token: "buyer-session" } })
      return
    }
    if (request.method() === "GET" && pathname.endsWith("/feed")) {
      await route.fulfill({
        json: {
          items: [],
          page: 1,
          page_size: 20,
          total: 0,
          total_pages: 0,
        },
      })
      return
    }
    if (
      request.method() === "GET" &&
      pathname.endsWith(`/orders/${orderID}`)
    ) {
      await route.fulfill({ json: orderResponse(orderStatus) })
      return
    }
    if (
      request.method() === "POST" &&
      pathname.endsWith(`/orders/${orderID}/cancel`)
    ) {
      cancelled = true
      orderStatus = "cancelled"
      await route.fulfill({ json: orderResponse(orderStatus) })
      return
    }
    if (request.method() === "POST" && pathname.endsWith("/payments")) {
      const body = request.postDataJSON() as { order_id?: number }
      expect(body.order_id).toBe(orderID)
      created = true
      await route.fulfill({ json: pendingPayment(), status: 201 })
      return
    }
    if (
      request.method() === "POST" &&
      pathname.endsWith(`/payments/${paymentNo}/pay`)
    ) {
      sessionStarted = true
      await route.fulfill({
        json: {
          ...pendingPayment(),
          provider_transaction_id: providerNo,
          checkout_url: checkoutURL,
        },
      })
      return
    }

    await route.fulfill({
      status: 404,
      json: { error: `unhandled ${request.method()} ${pathname}` },
    })
  })

  await page.route("http://127.0.0.1:8080/mockpay/checkout/**", async (route) => {
    await route.fulfill({
      contentType: "text/html",
      body: `<!doctype html>
<html lang="en">
  <body>
    <h1>MockPay checkout</h1>
    <a href="http://127.0.0.1:4173/orders/${orderID}">Return to merchant</a>
  </body>
</html>`,
    })
  })

  return {
    markPaid() {
      orderStatus = "paid"
    },
    wasCreated() {
      return created
    },
    wasSessionStarted() {
      return sessionStarted
    },
    wasCancelled() {
      return cancelled
    },
  }
}

test("buyer pays through MockPay checkout and returns to the paid order", async ({
  page,
}) => {
  const flow = await mockPaymentFlow(page)

  await page.goto("/signin")
  await page.getByLabel("Username").fill("buyer")
  await page.getByLabel("Password", { exact: true }).fill("password")
  await page.getByRole("button", { name: "Sign in", exact: true }).click()
  await page.waitForURL(/\/$/)

  await page.goto(`/orders/${orderID}`)
  await expect(
    page.getByRole("heading", { name: `Order #${orderID}` }),
  ).toBeVisible()

  await page.getByRole("button", { name: "Pay now" }).click()
  await expect(page).toHaveURL(`${checkoutURL}`)
  await expect(
    page.getByRole("heading", { name: "MockPay checkout" }),
  ).toBeVisible()
  expect(flow.wasCreated()).toBe(true)
  expect(flow.wasSessionStarted()).toBe(true)

  flow.markPaid()
  await page.getByRole("link", { name: "Return to merchant" }).click()
  await expect(page).toHaveURL(new RegExp(`/orders/${orderID}$`))
  await expect(page.getByText("Payment completed")).toBeVisible()
  await expect(page.getByText("Paid", { exact: true }).first()).toBeVisible()
})

test("buyer cancels an unpaid order and releases the payment panel", async ({
  page,
}) => {
  const flow = await mockPaymentFlow(page)

  await page.goto("/signin")
  await page.getByLabel("Username").fill("buyer")
  await page.getByLabel("Password", { exact: true }).fill("password")
  await page.getByRole("button", { name: "Sign in", exact: true }).click()
  await page.waitForURL(/\/$/)

  await page.goto(`/orders/${orderID}`)
  await expect(
    page.getByRole("heading", { name: `Order #${orderID}` }),
  ).toBeVisible()
  await expect(page.getByRole("button", { name: "Pay now" })).toBeVisible()

  await page
    .getByRole("button", { name: "Cancel order", exact: true })
    .click()
  await expect(page.getByText("Cancel this order?")).toBeVisible()
  await page
    .getByRole("button", { name: "Cancel order", exact: true })
    .last()
    .click()

  await expect(page.getByText("Cancelled", { exact: true }).first()).toBeVisible()
  await expect(
    page.getByRole("button", { name: "Cancel order", exact: true }),
  ).toHaveCount(0)
  await expect(page.getByRole("button", { name: "Pay now" })).toHaveCount(0)
  expect(flow.wasCancelled()).toBe(true)
})
