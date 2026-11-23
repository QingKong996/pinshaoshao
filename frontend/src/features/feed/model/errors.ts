import type { TFunction } from "i18next"
import { ApiError } from "@/shared/api/client"

export function describeFeedError(
  error: unknown,
  t: TFunction<"feed">,
): string {
  if (!(error instanceof ApiError)) {
    return t("errors.unknown")
  }
  if (error.status === 0) {
    return t("errors.network")
  }

  switch (error.status) {
    case 400:
      return t("errors.invalidQuery")
    case 401:
      return t("errors.unauthorized")
    case 500:
      return t("errors.server")
    default:
      return t("errors.unknown")
  }
}
