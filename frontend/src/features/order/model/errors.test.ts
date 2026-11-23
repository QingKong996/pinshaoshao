import type { TFunction } from "i18next"
import { describe, expect, it, vi } from "vitest"
import { ApiError } from "@/shared/api/client"
import { describeCheckoutError, describeOrderError } from "./errors"

function createTranslate() {
  return vi.fn((key: string) => key) as unknown as TFunction<"order">
}

describe("describeCheckoutError", () => {
  it("maps API failures to checkout messages", () => {
    const t = createTranslate()
    const cases = [
      [new ApiError(0, "offline"), "errors.network"],
      [new ApiError(400, "bad"), "errors.invalidCheckout"],
      [new ApiError(401, "bad"), "errors.unauthorized"],
      [new ApiError(409, "bad"), "errors.checkoutConflict"],
      [new ApiError(500, "bad"), "errors.server"],
      [new ApiError(418, "bad"), "errors.unknown"],
      [new Error("unexpected"), "errors.unknown"],
    ] as const

    for (const [error, message] of cases) {
      expect(describeCheckoutError(error, t)).toBe(message)
    }
  })
})

describe("describeOrderError", () => {
  it("maps API failures to order messages", () => {
    const t = createTranslate()
    const cases = [
      [new ApiError(0, "offline"), "errors.network"],
      [new ApiError(400, "bad"), "errors.invalidOrder"],
      [new ApiError(401, "bad"), "errors.unauthorized"],
      [new ApiError(404, "bad"), "errors.notFound"],
      [new ApiError(409, "bad"), "errors.cannotCancel"],
      [new ApiError(500, "bad"), "errors.server"],
      [new ApiError(418, "bad"), "errors.unknown"],
      [new Error("unexpected"), "errors.unknown"],
    ] as const

    for (const [error, message] of cases) {
      expect(describeOrderError(error, t)).toBe(message)
    }
  })
})
