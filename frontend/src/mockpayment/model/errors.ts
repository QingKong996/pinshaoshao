import type { TFunction } from "i18next"
import { ApiError } from "@/shared/api/client"

function statusMessage(
  error: unknown,
  t: TFunction<"mockpayment">,
  messages: Partial<Record<number, string>>,
) {
  if (!(error instanceof ApiError)) {
    return t("errors.unknown")
  }
  if (error.status === 0) {
    return t("errors.network")
  }
  return messages[error.status] ?? t("errors.unknown")
}

export function describeCreatePaymentError(
  error: unknown,
  t: TFunction<"mockpayment">,
) {
  return statusMessage(error, t, {
    400: t("errors.invalidRequest"),
    409: t("errors.duplicateOrder"),
    500: t("errors.server"),
  })
}

export function describePaymentError(
  error: unknown,
  t: TFunction<"mockpayment">,
) {
  return statusMessage(error, t, {
    400: t("errors.invalidPayment"),
    404: t("errors.notFound"),
    410: t("errors.expired"),
    500: t("errors.server"),
  })
}

export function describeClosePaymentError(
  error: unknown,
  t: TFunction<"mockpayment">,
) {
  return statusMessage(error, t, {
    404: t("errors.notFound"),
    409: t("errors.notClosable"),
    500: t("errors.server"),
  })
}

export function describeNotificationError(
  error: unknown,
  t: TFunction<"mockpayment">,
) {
  return statusMessage(error, t, {
    404: t("errors.notificationNotFound"),
    409: t("errors.notificationNotRetryable"),
    500: t("errors.server"),
  })
}
