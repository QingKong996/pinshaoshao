import i18n from "@/i18n"
import { DEFAULT_LANGUAGE } from "@/i18n/config"

const CURRENCY = "CNY"

export function formatPriceCents(cents: number): string {
  const locale = i18n.resolvedLanguage ?? DEFAULT_LANGUAGE
  return new Intl.NumberFormat(locale, {
    style: "currency",
    currency: CURRENCY,
  }).format(cents / 100)
}
