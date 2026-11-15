import { afterEach, describe, expect, it } from "vitest"
import {
  formatDateTime,
  formatDateTimeSeconds,
  formatPriceCents,
  formatPriceRange,
} from "@/shared/lib/format"
import i18n from "@/shared/i18n/instance"

afterEach(async () => {
  await i18n.changeLanguage("en")
})

describe("formatPriceCents", () => {
  it("formats cents as CNY using the active locale", () => {
    expect(formatPriceCents(123456)).toContain("1,234.56")
  })
})

describe("formatPriceRange", () => {
  it("returns null when either endpoint is missing", () => {
    expect(formatPriceRange(null, 100)).toBeNull()
    expect(formatPriceRange(100, null)).toBeNull()
  })

  it("collapses equal endpoints and separates different prices", () => {
    expect(formatPriceRange(1000, 1000)).toBe(formatPriceCents(1000))
    expect(formatPriceRange(1000, 2500)).toBe(
      `${formatPriceCents(1000)} – ${formatPriceCents(2500)}`,
    )
  })
})

describe("date formatting", () => {
  it("falls back to the original value for invalid dates", () => {
    expect(formatDateTime("not-a-date")).toBe("not-a-date")
    expect(formatDateTimeSeconds("not-a-date")).toBe("not-a-date")
  })

  it("formats valid timestamps without relying on the host locale", () => {
    expect(formatDateTime("2026-11-15T08:05:00")).not.toBe(
      "2026-11-15T08:05:00",
    )
    expect(formatDateTimeSeconds("2026-11-15T08:05:09")).toBe(
      "2026-11-15 08:05:09",
    )
  })
})
