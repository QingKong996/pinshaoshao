import { useTranslation } from "react-i18next"
import { Badge } from "@/shared/ui/badge"
import type { ProductStatus } from "../types"

export function ProductStatusBadge({ status }: { status: ProductStatus }) {
  const { t } = useTranslation("console")

  return (
    <Badge variant={status === "on_shelf" ? "secondary" : "outline"}>
      {t(`productStatus.${status}`)}
    </Badge>
  )
}
