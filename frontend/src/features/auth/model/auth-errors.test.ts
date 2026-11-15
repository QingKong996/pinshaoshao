import type { TFunction } from "i18next"
import { describe, expect, it, vi } from "vitest"
import { describeMerchantAuthError } from "@/features/auth/model/merchant-errors"
import { describeUserAuthError } from "@/features/auth/model/user-errors"
import { ApiError } from "@/shared/api/client"

function createTranslate() {
  return vi.fn((key: string) => key) as unknown as TFunction
}

describe("describeUserAuthError", () => {
  it("maps API failures to the expected messages", () => {
    const t = createTranslate()
    const cases = [
      [new ApiError(0, "offline"), "login", "auth:errors.network"],
      [new ApiError(400, "bad"), "login", "auth:errors.invalidForm"],
      [new ApiError(401, "bad"), "login", "auth:errors.invalidCredentials"],
      [new ApiError(403, "bad"), "login", "auth:errors.accountBanned"],
      [
        new ApiError(409, "bad"),
        "register",
        "auth:errors.usernameOrEmailTaken",
      ],
      [new ApiError(409, "bad"), "login", "auth:errors.conflict"],
      [new ApiError(500, "bad"), "login", "auth:errors.serverError"],
      [new ApiError(418, "bad"), "login", "auth:errors.requestFailed"],
      [new Error("unexpected"), "login", "auth:errors.unknown"],
    ] as const

    for (const [error, action, message] of cases) {
      expect(describeUserAuthError(error, action, t)).toBe(message)
    }
  })
})

describe("describeMerchantAuthError", () => {
  it("uses merchant-specific banned and conflict messages", () => {
    const t = createTranslate()

    expect(describeMerchantAuthError(new ApiError(403, "bad"), "login", t)).toBe(
      "auth:errors.shopBanned",
    )
    expect(
      describeMerchantAuthError(new ApiError(409, "bad"), "register", t),
    ).toBe("auth:errors.usernameOrEmailTaken")
  })

  it("falls back for unknown and non-API failures", () => {
    const t = createTranslate()

    expect(
      describeMerchantAuthError(new ApiError(429, "bad"), "login", t),
    ).toBe("auth:errors.requestFailed")
    expect(describeMerchantAuthError("bad", "login", t)).toBe(
      "auth:errors.unknown",
    )
  })
})
