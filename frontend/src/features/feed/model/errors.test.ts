import type { TFunction } from "i18next"
import { describe, expect, it, vi } from "vitest"
import { ApiError } from "@/shared/api/client"
import { describeFeedError } from "./errors"

function createTranslate() {
  return vi.fn((key: string) => key) as unknown as TFunction<"feed">
}

describe("describeFeedError", () => {
  it("maps API failures to feed messages", () => {
    const t = createTranslate()
    const cases = [
      [new ApiError(0, "offline"), "errors.network"],
      [new ApiError(400, "bad"), "errors.invalidQuery"],
      [new ApiError(401, "bad"), "errors.unauthorized"],
      [new ApiError(500, "bad"), "errors.server"],
      [new ApiError(418, "bad"), "errors.unknown"],
      [new Error("unexpected"), "errors.unknown"],
    ] as const

    for (const [error, message] of cases) {
      expect(describeFeedError(error, t)).toBe(message)
    }
  })
})
