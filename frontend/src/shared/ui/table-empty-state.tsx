import { TableCell, TableRow } from "@/shared/ui/table"
import { cn } from "@/shared/lib/utils"

export function TableEmptyState({
  colSpan,
  title,
  description,
  compact = false,
}: {
  colSpan: number
  title?: string
  description: string
  compact?: boolean
}) {
  return (
    <TableRow className="hover:bg-transparent">
      <TableCell
        colSpan={colSpan}
        className={cn(
          "px-4 text-center whitespace-normal",
          compact ? "py-8" : "py-16",
        )}
      >
        <div className="flex flex-col items-center gap-1">
          {title ? <p className="font-medium">{title}</p> : null}
          <p className="text-sm text-muted-foreground">{description}</p>
        </div>
      </TableCell>
    </TableRow>
  )
}
