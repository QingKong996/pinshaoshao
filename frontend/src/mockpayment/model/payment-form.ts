import type { ParseKeys, TFunction } from "i18next"
import { z } from "zod"
import type { CreateMockPaymentInput } from "../dto/payment"

const VALIDATION_MESSAGES = {
  merchantOrderNoRequired: "validation.merchantOrderNoRequired",
  merchantOrderNoLength: "validation.merchantOrderNoLength",
  amountRequired: "validation.amountRequired",
  amountInvalid: "validation.amountInvalid",
  currencyLength: "validation.currencyLength",
  subjectRequired: "validation.subjectRequired",
  subjectLength: "validation.subjectLength",
  notifyRequired: "validation.notifyRequired",
  notifyLength: "validation.notifyLength",
  notifyInvalid: "validation.notifyInvalid",
  returnLength: "validation.returnLength",
  returnInvalid: "validation.returnInvalid",
  expiresInvalid: "validation.expiresInvalid",
} as const

function isHttpUrl(value: string) {
  try {
    const url = new URL(value)
    return (
      (url.protocol === "http:" || url.protocol === "https:") &&
      url.hostname.length > 0 &&
      url.username.length === 0 &&
      url.password.length === 0 &&
      url.hash.length === 0
    )
  } catch {
    return false
  }
}

export const mockPaymentFormSchema = z.object({
  merchant_order_no: z
    .string()
    .trim()
    .min(1, VALIDATION_MESSAGES.merchantOrderNoRequired)
    .max(64, VALIDATION_MESSAGES.merchantOrderNoLength),
  amount_cents: z
    .string()
    .trim()
    .min(1, VALIDATION_MESSAGES.amountRequired)
    .regex(/^\d+$/, VALIDATION_MESSAGES.amountInvalid)
    .refine((value) => {
      const amount = Number(value)
      return Number.isSafeInteger(amount) && amount > 0
    }, VALIDATION_MESSAGES.amountInvalid),
  currency: z
    .string()
    .trim()
    .regex(/^[A-Za-z]{3}$/, VALIDATION_MESSAGES.currencyLength),
  subject: z
    .string()
    .trim()
    .min(1, VALIDATION_MESSAGES.subjectRequired)
    .max(200, VALIDATION_MESSAGES.subjectLength),
  notify_url: z
    .string()
    .trim()
    .min(1, VALIDATION_MESSAGES.notifyRequired)
    .max(2048, VALIDATION_MESSAGES.notifyLength)
    .refine(isHttpUrl, VALIDATION_MESSAGES.notifyInvalid),
  return_url: z
    .string()
    .trim()
    .max(2048, VALIDATION_MESSAGES.returnLength)
    .refine(
      (value) => value.length === 0 || isHttpUrl(value),
      VALIDATION_MESSAGES.returnInvalid,
    ),
  expires_in_seconds: z
    .string()
    .trim()
    .regex(/^\d+$/, VALIDATION_MESSAGES.expiresInvalid)
    .refine((value) => {
      const seconds = Number(value)
      return Number.isInteger(seconds) && seconds >= 60 && seconds <= 86_400
    }, VALIDATION_MESSAGES.expiresInvalid),
})

export type MockPaymentFormValues = z.infer<typeof mockPaymentFormSchema>

export function createMockPaymentInitialValues(): MockPaymentFormValues {
  return {
    merchant_order_no: `sandbox-${Date.now()}`,
    amount_cents: "",
    currency: "CNY",
    subject: "Sandbox order",
    notify_url: "http://localhost:8080/api/v1/payment/webhook",
    return_url: "",
    expires_in_seconds: "1800",
  }
}

export function toCreateMockPaymentInput(
  values: MockPaymentFormValues,
): CreateMockPaymentInput {
  const returnURL = values.return_url.trim()
  return {
    merchant_order_no: values.merchant_order_no.trim(),
    amount_cents: Number(values.amount_cents),
    currency: values.currency.trim().toUpperCase(),
    subject: values.subject.trim(),
    notify_url: values.notify_url.trim(),
    ...(returnURL ? { return_url: returnURL } : {}),
    expires_in_seconds: Number(values.expires_in_seconds),
  }
}

export function translateMockPaymentValidationMessage(
  message: string | undefined,
  t: TFunction<"mockpayment">,
) {
  if (!message) {
    return undefined
  }
  return t(message as ParseKeys<"mockpayment">)
}
