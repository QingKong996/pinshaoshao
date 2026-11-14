import type { TFunction } from "i18next"
import { ApiError } from "@/shared/api/client"

export function describeDashboardError(
  error: unknown,
  translate: TFunction<"console">,
) {
  if (!(error instanceof ApiError)) {
    return translate("errors.unknown")
  }
  if (error.status === 0) {
    return translate("errors.network")
  }

  switch (error.status) {
    case 401:
      return translate("errors.unauthorized")
    case 403:
      return translate("errors.banned")
    case 404:
      return translate("errors.notFound")
    case 500:
      return translate("errors.server")
    default:
      return translate("errors.unknown")
  }
}
