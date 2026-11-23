import type { TFunction } from "i18next"
import { ApiError } from "@/shared/api/client"

export function describeCheckoutError(
  error: unknown,
  t: TFunction<"order">,
): string {
  if (!(error instanceof ApiError)) {
    return t("errors.unknown")
  }
  if (error.status === 0) {
    return t("errors.network")
  }

  switch (error.status) {
    case 400:
      return t("errors.invalidCheckout")
    case 401:
      return t("errors.unauthorized")
    case 409:
      return t("errors.checkoutConflict")
    case 500:
      return t("errors.server")
    default:
      return t("errors.unknown")
  }
}

export function describeOrderError(
  error: unknown,
  t: TFunction<"order">,
): string {
  if (!(error instanceof ApiError)) {
    return t("errors.unknown")
  }
  if (error.status === 0) {
    return t("errors.network")
  }

  switch (error.status) {
    case 400:
      return t("errors.invalidOrder")
    case 401:
      return t("errors.unauthorized")
    case 404:
      return t("errors.notFound")
    case 409:
      return t("errors.cannotCancel")
    case 500:
      return t("errors.server")
    default:
      return t("errors.unknown")
  }
}
