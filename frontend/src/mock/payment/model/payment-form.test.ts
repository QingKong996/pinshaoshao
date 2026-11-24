import type { TFunction } from "i18next"
import { describe, expect, it, vi } from "vitest"
import {
  mockPaymentFormSchema,
  toCreateMockPaymentInput,
  translateMockPaymentValidationMessage,
  type MockPaymentFormValues,
} from "./payment-form"

const validValues: MockPaymentFormValues = {
  merchant_order_no: "  order-1  ",
  amount_cents: "3998",
  currency: "cny",
  subject: "  T-shirt  ",
  notify_url: "https://example.com/webhook",
  return_url: "",
  expires_in_seconds: "1800",
}

describe("mockPaymentFormSchema", () => {
  it("normalizes form values into the API payload", () => {
    const parsed = mockPaymentFormSchema.parse(validValues)

    expect(parsed).toMatchObject({
      merchant_order_no: "order-1",
      subject: "T-shirt",
    })
    expect(toCreateMockPaymentInput(parsed)).toEqual({
      merchant_order_no: "order-1",
      amount_cents: 3998,
      currency: "CNY",
      subject: "T-shirt",
      notify_url: "https://example.com/webhook",
      expires_in_seconds: 1800,
    })
  })

  it.each([
    { field: "merchant_order_no" as const, value: " " },
    { field: "merchant_order_no" as const, value: "a".repeat(65) },
    { field: "amount_cents" as const, value: "0" },
    { field: "amount_cents" as const, value: "-1" },
    { field: "amount_cents" as const, value: "1.5" },
    { field: "amount_cents" as const, value: "9007199254740992" },
    { field: "currency" as const, value: "CN" },
    { field: "subject" as const, value: " " },
    { field: "subject" as const, value: "a".repeat(201) },
    { field: "notify_url" as const, value: "ftp://example.com/hook" },
    {
      field: "notify_url" as const,
      value: "https://user@example.com/hook",
    },
    {
      field: "notify_url" as const,
      value: "https://example.com/hook#fragment",
    },
    { field: "return_url" as const, value: "ftp://example.com/return" },
    { field: "expires_in_seconds" as const, value: "59" },
    { field: "expires_in_seconds" as const, value: "86401" },
    { field: "expires_in_seconds" as const, value: "soon" },
  ])("rejects invalid $field value", ({ field, value }) => {
    expect(
      mockPaymentFormSchema.safeParse({
        ...validValues,
        [field]: value,
      }).success,
    ).toBe(false)
  })

  it("accepts boundary values and an optional return URL", () => {
    expect(
      mockPaymentFormSchema.safeParse({
        ...validValues,
        merchant_order_no: "a".repeat(64),
        amount_cents: "1",
        subject: "a".repeat(200),
        return_url: "https://example.com/return?order=1",
        expires_in_seconds: "60",
      }).success,
    ).toBe(true)
    expect(
      mockPaymentFormSchema.safeParse({
        ...validValues,
        expires_in_seconds: "86400",
      }).success,
    ).toBe(true)
  })
})

describe("translateMockPaymentValidationMessage", () => {
  it("translates known validation keys and ignores missing messages", () => {
    const t = vi.fn((key: string) => `translated:${key}`) as unknown as TFunction<
      "mockpayment"
    >

    expect(translateMockPaymentValidationMessage(undefined, t)).toBeUndefined()
    expect(
      translateMockPaymentValidationMessage(
        "validation.merchantOrderNoRequired",
        t,
      ),
    ).toBe("translated:validation.merchantOrderNoRequired")
    expect(t).toHaveBeenCalledOnce()
  })
})
