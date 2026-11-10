import type { ParseKeys, TFunction } from "i18next"
import { ApiError } from "@/lib/api"

type MerchantKey = ParseKeys<"merchant">

const SERVER_MESSAGE_KEYS: Record<string, MerchantKey> = {
  "authentication required": "errors.unauthorized",
  "invalid or expired shop session": "errors.unauthorized",
  "shop is banned": "errors.banned",
  "invalid json format": "errors.invalidInput",
  "internal server error": "errors.server",
  "product not found": "errors.notFound",
  "category not found": "errors.notFound",
  "product sku not found": "errors.notFound",
  "invalid category": "errors.invalidCategory",
  "category name already exists": "errors.categoryExists",
  "product sku with the same specs already exists": "errors.skuExists",
  "product sku specs must have the same keys as the product's existing skus":
    "errors.specsMismatch",
}

export function describeMerchantError(
  error: unknown,
  t: TFunction<"merchant">,
): string {
  if (!(error instanceof ApiError)) {
    return t("errors.unknown")
  }
  if (error.status === 0) {
    return t("errors.network")
  }

  const key = SERVER_MESSAGE_KEYS[error.message]
  if (key) {
    return t(key)
  }

  switch (error.status) {
    case 400:
      return t("errors.invalidInput")
    case 401:
      return t("errors.unauthorized")
    case 403:
      return t("errors.forbidden")
    case 404:
      return t("errors.notFound")
    case 409:
      return t("errors.conflict")
    case 500:
      return t("errors.server")
    default:
      return t("errors.unknown")
  }
}
