export function formatOrderSpecs(
  specs: Record<string, string>,
  fallback: string,
): string {
  const entries = Object.entries(specs)
  if (entries.length === 0) {
    return fallback
  }
  return entries.map(([key, value]) => `${key}: ${value}`).join(" / ")
}
