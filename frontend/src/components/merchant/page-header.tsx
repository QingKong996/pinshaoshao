import type { ReactNode } from "react"
import { ArrowLeft } from "lucide-react"
import { Link } from "react-router"
import { cn } from "@/lib/utils"

type PageHeaderProps = {
  title: ReactNode
  description?: ReactNode
  actions?: ReactNode
  className?: string
} & (
  | { backTo: string; backLabel: string }
  | { backTo?: undefined; backLabel?: undefined }
)

export function PageHeader({
  title,
  description,
  actions,
  backTo,
  backLabel,
  className,
}: PageHeaderProps) {
  return (
    <div
      className={cn(
        "flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between",
        className,
      )}
    >
      <div className="flex min-w-0 flex-col gap-1.5">
        {backTo ? (
          <Link
            to={backTo}
            className="-ml-2 mb-1 inline-flex w-fit items-center gap-1.5 rounded-md px-2 py-1 text-sm text-muted-foreground transition-colors hover:text-foreground focus-visible:ring-3 focus-visible:ring-ring/50"
          >
            <ArrowLeft className="size-4" aria-hidden />
            {backLabel}
          </Link>
        ) : null}
        <h1 className="font-heading text-2xl font-semibold tracking-tight text-balance">
          {title}
        </h1>
        {description ? (
          <p className="max-w-2xl text-sm leading-relaxed text-muted-foreground">
            {description}
          </p>
        ) : null}
      </div>
      {actions ? (
        <div className="flex shrink-0 flex-wrap items-center gap-2">{actions}</div>
      ) : null}
    </div>
  )
}
