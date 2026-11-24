import type { TFunction } from "i18next"
import { ApiError } from "@/shared/api/client"

function statusMessage(
  error: unknown,
  t: TFunction<"mocklogistics">,
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

export function describeCreateShipmentError(
  error: unknown,
  t: TFunction<"mocklogistics">,
) {
  return statusMessage(error, t, {
    400: t("errors.invalidRequest"),
    409: t("errors.duplicateOrder"),
    500: t("errors.server"),
  })
}

export function describeShipmentError(
  error: unknown,
  t: TFunction<"mocklogistics">,
) {
  return statusMessage(error, t, {
    400: t("errors.invalidShipment"),
    404: t("errors.notFound"),
    500: t("errors.server"),
  })
}

export function describeShipmentActionError(
  error: unknown,
  t: TFunction<"mocklogistics">,
) {
  return statusMessage(error, t, {
    400: t("errors.invalidShipment"),
    404: t("errors.notFound"),
    409: t("errors.invalidTransition"),
    500: t("errors.server"),
  })
}

export function describeNotificationError(
  error: unknown,
  t: TFunction<"mocklogistics">,
) {
  return statusMessage(error, t, {
    400: t("errors.invalidRequest"),
    404: t("errors.notificationNotFound"),
    409: t("errors.notificationNotRetryable"),
    500: t("errors.server"),
  })
}
