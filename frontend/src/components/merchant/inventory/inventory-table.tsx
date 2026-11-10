import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { SpecBadges } from "@/components/merchant/skus/spec-badges"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { formatDateTime, formatPriceCents } from "@/lib/format"
import type { InventoryItem } from "@/lib/merchant-api"
import { cn } from "@/lib/utils"

type InventoryTableProps = {
  items: InventoryItem[]
  onAdjust: (item: InventoryItem) => void
}

export function InventoryTable({ items, onAdjust }: InventoryTableProps) {
  const { t } = useTranslation("merchant")

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>{t("inventory.columns.product")}</TableHead>
          <TableHead>{t("inventory.columns.category")}</TableHead>
          <TableHead>{t("inventory.columns.specs")}</TableHead>
          <TableHead>{t("inventory.columns.price")}</TableHead>
          <TableHead>{t("inventory.columns.stock")}</TableHead>
          <TableHead>{t("inventory.columns.updatedAt")}</TableHead>
          <TableHead className="text-right">{t("common.actions")}</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {items.map((item) => {
          const isOut = item.stock === 0
          return (
            <TableRow key={item.sku_id}>
              <TableCell>
                <span className="flex items-center gap-2">
                  <Link
                    to={`/merchant/products/${item.product_id}`}
                    className="font-medium underline-offset-4 hover:underline"
                  >
                    {item.product_name}
                  </Link>
                  {item.is_enabled ? null : (
                    <Badge variant="outline">{t("status.disabled")}</Badge>
                  )}
                </span>
              </TableCell>
              <TableCell>
                {item.category_name || t("common.uncategorized")}
              </TableCell>
              <TableCell>
                <SpecBadges specs={item.specs} />
              </TableCell>
              <TableCell className="tabular-nums">
                {formatPriceCents(item.price_cents)}
              </TableCell>
              <TableCell>
                <span className="flex items-center gap-2">
                  <span
                    className={cn(
                      "font-medium tabular-nums",
                      isOut && "text-destructive",
                    )}
                  >
                    {item.stock}
                  </span>
                  {isOut ? (
                    <Badge variant="destructive">
                      {t("inventory.badge.out")}
                    </Badge>
                  ) : null}
                </span>
              </TableCell>
              <TableCell>{formatDateTime(item.updated_at)}</TableCell>
              <TableCell className="text-right">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => onAdjust(item)}
                >
                  {t("inventory.adjust")}
                </Button>
              </TableCell>
            </TableRow>
          )
        })}
      </TableBody>
    </Table>
  )
}
