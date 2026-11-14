import { useTranslation } from "react-i18next"
import { Badge } from "@/shared/ui/badge"
import type { ProductStatus } from "../types"

export function ProductStatusBadge({ status }: { status: ProductStatus }) {
  const { t } = useTranslation("console")
  const isOnShelf = status === "on_shelf"

  return (
    <Badge
      className={
        isOnShelf
          ? "border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-900/60 dark:bg-emerald-950/40 dark:text-emerald-300"
          : "border-rose-200 bg-rose-50 text-rose-700 dark:border-rose-900/60 dark:bg-rose-950/40 dark:text-rose-300"
      }
    >
      {t(`productStatus.${status}`)}
    </Badge>
  )
}
