import { ArrowLeft, Pencil } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { PageHeader } from "@/shared/ui/page-header"
import { Button } from "@/shared/ui/button"
import type { Product } from "../types"
import { ProductStatusBadge } from "./product-status-badge"

export function ProductDetailBackLink({ backHref }: { backHref: string }) {
  const { t } = useTranslation("console")

  return (
    <Link
      to={backHref}
      className="inline-flex w-fit items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
    >
      <ArrowLeft className="size-4" aria-hidden />
      {t("productDetail.back")}
    </Link>
  )
}

export function ProductDetailHeader({
  product,
  onEdit,
}: {
  product: Product
  onEdit: () => void
}) {
  const { t } = useTranslation("console")

  return (
    <PageHeader
      title={
        <span className="flex flex-wrap items-center gap-2">
          {product.name}
          <ProductStatusBadge status={product.status} />
        </span>
      }
      description={product.description || t("productDetail.noDescription")}
      actions={
        <Button onClick={onEdit}>
          <Pencil aria-hidden />
          {t("productDetail.edit")}
        </Button>
      }
    />
  )
}
