export const SUPPORTED_LANGUAGES = ["zh-CN", "en"] as const

export type Language = (typeof SUPPORTED_LANGUAGES)[number]

export const DEFAULT_LANGUAGE: Language = "zh-CN"

const LANGUAGE_STORAGE_KEY = "pinshaoshao.language"

export function isLanguage(value: string | null | undefined): value is Language {
  return SUPPORTED_LANGUAGES.includes(value as Language)
}

export function normalizeLanguage(
  value: string | null | undefined,
): Language {
  return isLanguage(value) ? value : DEFAULT_LANGUAGE
}

function matchLanguage(value: string | null | undefined): Language | null {
  if (!value) {
    return null
  }
  const normalized = value.toLowerCase()
  if (normalized.startsWith("zh")) {
    return "zh-CN"
  }
  if (normalized.startsWith("en")) {
    return "en"
  }
  return null
}

function readStoredLanguage(): Language | null {
  try {
    const stored = localStorage.getItem(LANGUAGE_STORAGE_KEY)
    return isLanguage(stored) ? stored : null
  } catch {
    return null
  }
}

export function resolveInitialLanguage(): Language {
  const stored = readStoredLanguage()
  if (stored) {
    return stored
  }

  for (const candidate of [navigator.language, ...navigator.languages]) {
    const matched = matchLanguage(candidate)
    if (matched) {
      return matched
    }
  }

  return DEFAULT_LANGUAGE
}

export function rememberLanguage(language: Language) {
  try {
    localStorage.setItem(LANGUAGE_STORAGE_KEY, language)
  } catch {
    // the current page still switches, only persistence is unavailable
  }
}

export function applyDocumentLanguage(language: Language) {
  document.documentElement.lang = language
}
