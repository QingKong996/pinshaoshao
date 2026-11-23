export function formatProductSpecs(
  specs: Record<string, string>,
  fallback: string,
) {
  const entries = Object.entries(specs).sort(([left], [right]) =>
    left.localeCompare(right),
  )
  if (entries.length === 0) {
    return fallback
  }
  return entries.map(([key, value]) => `${key}: ${value}`).join(" / ")
}
