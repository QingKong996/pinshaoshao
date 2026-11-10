export function formatSpecs(specs: Record<string, string>): string {
  return Object.entries(specs)
    .map(([key, value]) => `${key}: ${value}`)
    .join(" · ")
}
