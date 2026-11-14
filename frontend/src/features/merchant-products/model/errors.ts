import type { TFunction } from "i18next"
import { ApiError } from "@/shared/api/client"

export const PRODUCT_NOT_FOUND_ERROR = new ApiError(404, "product not found")

export function describeConsoleError(
  error: unknown,
  t: TFunction<"console">,
): string {
  if (!(error instanceof ApiError)) {
    return t("errors.unknown")
  }
  if (error.status === 0) {
    return t("errors.network")
  }

  switch (error.status) {
    case 401:
      return t("errors.unauthorized")
    case 403:
      return t("errors.banned")
    case 404:
      return t("errors.notFound")
    case 500:
      return t("errors.server")
    default:
      return t("errors.unknown")
  }
}
