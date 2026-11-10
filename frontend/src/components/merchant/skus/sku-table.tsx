import { Ellipsis, Pencil, Trash2 } from "lucide-react"
import { useTranslation } from "react-i18next"
import { SpecBadges } from "@/components/merchant/skus/spec-badges"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { formatDateTime, formatPriceCents } from "@/lib/format"
import type { SKU } from "@/lib/merchant-api"

type SKUTableProps = {
  skus: SKU[]
  onEdit: (sku: SKU) => void
  onDelete: (sku: SKU) => void
}

export function SKUTable({ skus, onEdit, onDelete }: SKUTableProps) {
  const { t } = useTranslation("merchant")

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>{t("skus.columns.specs")}</TableHead>
          <TableHead>{t("skus.columns.price")}</TableHead>
          <TableHead>{t("skus.columns.stock")}</TableHead>
          <TableHead>{t("skus.columns.enabled")}</TableHead>
          <TableHead>{t("skus.columns.updatedAt")}</TableHead>
          <TableHead className="text-right">{t("common.actions")}</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {skus.map((sku) => (
          <TableRow key={sku.id}>
            <TableCell>
              <SpecBadges specs={sku.specs} />
            </TableCell>
            <TableCell className="tabular-nums">
              {formatPriceCents(sku.price_cents)}
            </TableCell>
            <TableCell className="tabular-nums">{sku.stock}</TableCell>
            <TableCell>
              <Badge variant={sku.is_enabled ? "secondary" : "outline"}>
                {sku.is_enabled ? t("status.enabled") : t("status.disabled")}
              </Badge>
            </TableCell>
            <TableCell>{formatDateTime(sku.updated_at)}</TableCell>
            <TableCell className="text-right">
              <DropdownMenu>
                <DropdownMenuTrigger
                  render={
                    <Button
                      variant="ghost"
                      size="icon-sm"
                      aria-label={t("skus.rowActions")}
                    />
                  }
                >
                  <Ellipsis aria-hidden />
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-36">
                  <DropdownMenuItem onClick={() => onEdit(sku)}>
                    <Pencil aria-hidden />
                    {t("skus.actions.edit")}
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    variant="destructive"
                    onClick={() => onDelete(sku)}
                  >
                    <Trash2 aria-hidden />
                    {t("skus.actions.delete")}
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}
