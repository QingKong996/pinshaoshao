import { useTranslation } from "react-i18next"
import { Badge } from "@/components/ui/badge"
import type { ProductStatus } from "@/api/merchant-console"

export function ProductStatusBadge({ status }: { status: ProductStatus }) {
  const { t } = useTranslation("console")

  return (
    <Badge variant={status === "on_shelf" ? "secondary" : "outline"}>
      {t(`productStatus.${status}`)}
    </Badge>
  )
}
