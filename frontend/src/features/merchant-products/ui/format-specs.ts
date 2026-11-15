export function formatSpecs(
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
