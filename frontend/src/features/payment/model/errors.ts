import type { TFunction } from "i18next"
import { ApiError } from "@/shared/api/client"

export class PaymentClosedError extends Error {
  constructor() {
    super("payment is closed")
    this.name = "PaymentClosedError"
  }
}

export class CheckoutUnavailableError extends Error {
  constructor() {
    super("checkout URL is unavailable")
    this.name = "CheckoutUnavailableError"
  }
}

export function describeStartPaymentError(
  error: unknown,
  t: TFunction<"payment">,
) {
  if (error instanceof PaymentClosedError) {
    return t("errors.closed")
  }
  if (error instanceof CheckoutUnavailableError) {
    return t("errors.checkoutUnavailable")
  }
  if (!(error instanceof ApiError)) {
    return t("errors.unknown")
  }
  if (error.status === 0) {
    return t("errors.network")
  }

  switch (error.status) {
    case 400:
      return t("errors.invalidRequest")
    case 401:
      return t("errors.unauthorized")
    case 404:
      return t("errors.notFound")
    case 409:
      return t("errors.notPayable")
    case 502:
      return t("errors.gateway")
    case 500:
      return t("errors.server")
    default:
      return t("errors.unknown")
  }
}
