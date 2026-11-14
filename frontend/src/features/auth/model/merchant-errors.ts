import type { TFunction } from "i18next"
import { ApiError } from "@/shared/api/client"

export type MerchantAuthAction = "login" | "register"

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
