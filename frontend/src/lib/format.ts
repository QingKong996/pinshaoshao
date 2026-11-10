import i18n from "@/i18n"

const CURRENCY = "CNY"
const EMPTY_VALUE = "—"

const currencyFormatters = new Map<string, Intl.NumberFormat>()
const dateTimeFormatters = new Map<string, Intl.DateTimeFormat>()

function activeLocale(): string {
  return i18n.resolvedLanguage ?? "zh-CN"
}

function currencyFormatter(locale: string): Intl.NumberFormat {
  let formatter = currencyFormatters.get(locale)
  if (!formatter) {
    formatter = new Intl.NumberFormat(locale, {
      style: "currency",
      currency: CURRENCY,
    })
    currencyFormatters.set(locale, formatter)
  }
  return formatter
}

function dateTimeFormatter(locale: string): Intl.DateTimeFormat {
  let formatter = dateTimeFormatters.get(locale)
  if (!formatter) {
    formatter = new Intl.DateTimeFormat(locale, {
      dateStyle: "medium",
      timeStyle: "short",
    })
    dateTimeFormatters.set(locale, formatter)
  }
  return formatter
}

export function formatPriceCents(cents: number | null | undefined): string {
  if (cents === null || cents === undefined) {
    return EMPTY_VALUE
  }
  return currencyFormatter(activeLocale()).format(cents / 100)
}

export function formatPriceRange(
  minCents: number | null,
  maxCents: number | null,
): string {
  if (minCents === null || maxCents === null) {
    return EMPTY_VALUE
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
  return dateTimeFormatter(activeLocale()).format(date)
}

export function yuanToCents(value: string): number {
  return Math.round(Number(value) * 100)
}

export function centsToYuan(cents: number): string {
  return (cents / 100).toFixed(2)
}
