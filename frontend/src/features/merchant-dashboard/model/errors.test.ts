import type { TFunction } from "i18next"
import { describe, expect, it, vi } from "vitest"
import { describeDashboardError } from "@/features/merchant-dashboard/model/errors"
import { ApiError } from "@/shared/api/client"

const t = vi.fn((key: string) => key) as unknown as TFunction<"console">

describe("describeDashboardError", () => {
  it("maps known API failures and falls back for others", () => {
    const cases = [
      [new ApiError(0, "offline"), "errors.network"],
      [new ApiError(401, "expired"), "errors.unauthorized"],
      [new ApiError(403, "banned"), "errors.banned"],
      [new ApiError(404, "missing"), "errors.notFound"],
      [new ApiError(500, "failed"), "errors.server"],
      [new ApiError(422, "invalid"), "errors.unknown"],
      [new Error("unexpected"), "errors.unknown"],
    ] as const

    for (const [error, expected] of cases) {
      expect(describeDashboardError(error, t)).toBe(expected)
    }
  })
})
