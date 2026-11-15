import { describe, expect, it, vi } from "vitest"
import {
  DEFAULT_LANGUAGE,
  applyDocumentLanguage,
  isLanguage,
  normalizeLanguage,
  rememberLanguage,
  resolveInitialLanguage,
} from "@/shared/i18n/config"

describe("language configuration", () => {
  it("validates and normalizes supported language values", () => {
    expect(isLanguage("en")).toBe(true)
    expect(isLanguage("zh-CN")).toBe(true)
    expect(isLanguage("fr")).toBe(false)
    expect(normalizeLanguage("fr")).toBe(DEFAULT_LANGUAGE)
  })

  it("uses navigator languages when nothing is stored", () => {
    vi.spyOn(window.navigator, "language", "get").mockReturnValue("en-GB")

    expect(resolveInitialLanguage()).toBe("en")
  })

  it("prefers a valid stored language", () => {
    rememberLanguage("zh-CN")
    vi.spyOn(window.navigator, "language", "get").mockReturnValue("en-US")

    expect(resolveInitialLanguage()).toBe("zh-CN")
  })

  it("ignores invalid stored values and unsupported navigator languages", () => {
    rememberLanguage("en")
    localStorage.setItem(localStorage.key(0)!, "fr")
    vi.spyOn(window.navigator, "language", "get").mockReturnValue("fr-FR")
    vi.spyOn(window.navigator, "languages", "get").mockReturnValue([])

    expect(resolveInitialLanguage()).toBe(DEFAULT_LANGUAGE)
  })

  it("updates the document language", () => {
    applyDocumentLanguage("en")

    expect(document.documentElement.lang).toBe("en")
  })
})
