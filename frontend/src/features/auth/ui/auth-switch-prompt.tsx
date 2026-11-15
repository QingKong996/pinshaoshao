import { Link } from "react-router"

export function AuthSwitchPrompt({
  prompt,
  linkLabel,
  to,
}: {
  prompt: string
  linkLabel: string
  to: string
}) {
  return (
    <>
      {prompt}
      <Link
        to={to}
        className="ml-1 rounded-sm font-medium text-foreground underline-offset-4 transition-colors hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/50"
      >
        {linkLabel}
      </Link>
    </>
  )
}
