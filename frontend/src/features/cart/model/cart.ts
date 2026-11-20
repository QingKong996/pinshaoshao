import type { CartItem } from "../types"

export function formatCartSpecs(
  specs: Record<string, string>,
  emptyLabel: string,
) {
  const entries = Object.entries(specs).sort(([left], [right]) =>
    left.localeCompare(right),
  )
  if (entries.length === 0) {
    return emptyLabel
  }
  return entries.map(([key, value]) => `${key}: ${value}`).join(" / ")
}

export function summarizeSelection(
  items: CartItem[],
  selectedIDs: ReadonlySet<number>,
) {
  return items.reduce(
    (summary, item) => {
      if (
        item.availability !== "available" ||
        !selectedIDs.has(item.id)
      ) {
        return summary
      }
      summary.count += 1
      summary.subtotalCents += item.line_total_cents
      return summary
    },
    { count: 0, subtotalCents: 0 },
  )
}
