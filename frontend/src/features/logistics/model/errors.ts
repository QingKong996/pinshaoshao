import type { TFunction } from "i18next"
import { ApiError } from "@/shared/api/client"

export function describeShipmentError(
  error: unknown,
  t: TFunction<"logistics">,
) {
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
      return t("errors.notShippable")
    case 502:
      return t("errors.provider")
    case 500:
      return t("errors.server")
    default:
      return t("errors.unknown")
  }
}
