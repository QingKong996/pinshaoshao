import type { ParseKeys, TFunction } from "i18next"
import { ApiError } from "@/lib/api"

export type AuthAction = "login" | "register"

const SERVER_MESSAGE_KEYS: Record<string, ParseKeys> = {
  "invalid credentials": "auth:errors.invalidCredentials",
  "user is banned": "auth:errors.accountBanned",
  "username already exists": "auth:errors.usernameTaken",
  "email already exists": "auth:errors.emailTaken",
  "invalid json format": "auth:errors.invalidForm",
  "authentication required": "auth:errors.sessionExpired",
  "invalid or expired session": "auth:errors.sessionExpired",
  "internal server error": "auth:errors.serverError",
}

export function describeAuthError(
  error: unknown,
  action: AuthAction,
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
      return t("auth:errors.accountBanned")
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
