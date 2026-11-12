import i18n from "@/shared/i18n/instance"
import { DEFAULT_LANGUAGE } from "@/shared/i18n/config"

const CURRENCY = "CNY"

function resolveLocale(): string {
  return i18n.resolvedLanguage ?? DEFAULT_LANGUAGE
}

export function formatPriceCents(cents: number): string {
  return new Intl.NumberFormat(resolveLocale(), {
    style: "currency",
    currency: CURRENCY,
  }).format(cents / 100)
}

export function formatPriceRange(
  minCents: number | null,
  maxCents: number | null,
): string | null {
  if (minCents === null || maxCents === null) {
    return null
  }
  if (minCents === maxCents) {
    return formatPriceCents(minCents)
  }
  return `${formatPriceCents(minCents)} – ${formatPriceCents(maxCents)}`
}

export function formatDateTime(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return new Intl.DateTimeFormat(resolveLocale(), {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date)
}
