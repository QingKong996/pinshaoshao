import type { ParseKeys, TFunction } from "i18next"
import { ApiError } from "@/api/client"

export type MerchantAuthAction = "login" | "register"

const SERVER_MESSAGE_KEYS: Record<string, ParseKeys> = {
  "invalid credentials": "auth:errors.invalidCredentials",
  "shop is banned": "auth:errors.shopBanned",
  "shop username already exists": "auth:errors.shopUsernameTaken",
  "shop email already exists": "auth:errors.shopEmailTaken",
  "invalid json format": "auth:errors.invalidForm",
  "authentication required": "auth:errors.sessionExpired",
  "invalid or expired shop session": "auth:errors.sessionExpired",
  "internal server error": "auth:errors.serverError",
}

export function describeMerchantAuthError(
  error: unknown,
  action: MerchantAuthAction,
  t: TFunction,
): string {
  if (!(error instanceof ApiError)) {
    return t("auth:errors.unknown")
  }
  if (error.status === 0) {
    return t("auth:errors.network")
  }

  const key = SERVER_MESSAGE_KEYS[error.message]
  if (key) {
    return t(key)
  }

  switch (error.status) {
    case 400:
      return t("auth:errors.invalidForm")
    case 401:
      return t("auth:errors.invalidCredentials")
    case 403:
      return t("auth:errors.shopBanned")
    case 409:
      return t(
        action === "register"
          ? "auth:errors.usernameOrEmailTaken"
          : "auth:errors.conflict",
      )
    case 500:
      return t("auth:errors.serverError")
    default:
      return t("auth:errors.requestFailed")
  }
}
