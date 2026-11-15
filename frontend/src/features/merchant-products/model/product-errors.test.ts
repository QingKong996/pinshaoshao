import type { TFunction } from "i18next"
import { describe, expect, it, vi } from "vitest"
import { describeConsoleError } from "@/features/merchant-products/model/errors"
import { ApiError } from "@/shared/api/client"

const t = vi.fn((key: string) => key) as unknown as TFunction<"console">

describe("describeConsoleError", () => {
  it("maps HTTP statuses to user-facing messages", () => {
    const cases = [
      [new ApiError(0, "offline"), "errors.network"],
      [new ApiError(401, "expired"), "errors.unauthorized"],
      [new ApiError(403, "banned"), "errors.banned"],
      [new ApiError(404, "missing"), "errors.notFound"],
      [new ApiError(500, "failed"), "errors.server"],
      [new ApiError(418, "teapot"), "errors.unknown"],
      [new Error("unexpected"), "errors.unknown"],
    ] as const

    for (const [error, expected] of cases) {
      expect(describeConsoleError(error, t)).toBe(expected)
    }
  })

  it("distinguishes duplicate category names from other conflicts", () => {
    expect(
      describeConsoleError(
        new ApiError(409, "category name already exists"),
        t,
      ),
    ).toBe("errors.categoryNameTaken")
    expect(describeConsoleError(new ApiError(409, "other conflict"), t)).toBe(
      "errors.unknown",
    )
  })
})
