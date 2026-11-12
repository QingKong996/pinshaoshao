import type { ParseKeys, TFunction } from "i18next"
import { ApiError } from "@/shared/api/client"

const SERVER_MESSAGE_KEYS: Record<string, ParseKeys<"console">> = {
  "authentication required": "errors.unauthorized",
  "invalid or expired shop session": "errors.unauthorized",
  "shop is banned": "errors.banned",
  "shop not found": "errors.notFound",
  "internal server error": "errors.server",
}

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
  const key = SERVER_MESSAGE_KEYS[error.message]
  if (key) {
    return translate(key)
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
