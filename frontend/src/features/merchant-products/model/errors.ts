import type { ParseKeys, TFunction } from "i18next"
import { ApiError } from "@/shared/api/client"

type ConsoleKey = ParseKeys<"console">

export const PRODUCT_NOT_FOUND_ERROR = new ApiError(404, "product not found")

const SERVER_MESSAGE_KEYS: Record<string, ConsoleKey> = {
  "authentication required": "errors.unauthorized",
  "invalid or expired shop session": "errors.unauthorized",
  "shop is banned": "errors.banned",
  "shop not found": "errors.notFound",
  "product not found": "errors.productNotFound",
  "invalid category": "errors.invalidCategory",
  "invalid product sku specs": "errors.productSKUSpecs",
  "product sku specs must have the same keys as the product's existing skus":
    "errors.productSKUSpecsMismatch",
  "product sku with the same specs already exists":
    "errors.productSKUAlreadyExists",
  "internal server error": "errors.server",
}

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

  const key = SERVER_MESSAGE_KEYS[error.message]
  if (key) {
    return t(key)
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
