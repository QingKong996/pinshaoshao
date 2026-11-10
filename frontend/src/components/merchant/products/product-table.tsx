import { MoreHorizontal, Pencil, Trash2 } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Badge } from "@/components/ui/badge"
import { buttonVariants } from "@/components/ui/button"
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
import { formatDateTime, formatPriceRange } from "@/lib/format"
import type { Product } from "@/lib/merchant-api"

type ProductTableProps = {
  products: Product[]
  onEdit: (product: Product) => void
  onDelete: (product: Product) => void
}

export function ProductTable({ products, onEdit, onDelete }: ProductTableProps) {
  const { t } = useTranslation("merchant")

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>{t("products.columns.product")}</TableHead>
          <TableHead>{t("products.columns.category")}</TableHead>
          <TableHead>{t("products.columns.status")}</TableHead>
          <TableHead className="text-right">
            {t("products.columns.sku")}
          </TableHead>
          <TableHead className="text-right">
            {t("products.columns.stock")}
          </TableHead>
          <TableHead className="text-right">
            {t("products.columns.price")}
          </TableHead>
          <TableHead>{t("products.columns.updatedAt")}</TableHead>
          <TableHead className="w-10">
            <span className="sr-only">{t("common.actions")}</span>
          </TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {products.map((product) => (
          <TableRow key={product.id}>
            <TableCell className="max-w-56">
              <div className="flex min-w-0 flex-col">
                <div className="truncate font-medium">{product.name}</div>
                <div className="truncate text-xs text-muted-foreground">
                  {product.description || t("common.none")}
                </div>
              </div>
            </TableCell>
            <TableCell>
              {product.category_name ? (
                <Badge variant="outline">{product.category_name}</Badge>
              ) : (
                <span className="text-muted-foreground">
                  {t("common.uncategorized")}
                </span>
              )}
            </TableCell>
            <TableCell>
              <Badge
                variant={product.status === "on_shelf" ? "secondary" : "outline"}
              >
                {product.status === "on_shelf"
                  ? t("status.onShelf")
                  : t("status.offShelf")}
              </Badge>
            </TableCell>
            <TableCell className="text-right tabular-nums">
              {product.sku_count}
            </TableCell>
            <TableCell className="text-right tabular-nums">
              {product.total_stock}
            </TableCell>
            <TableCell className="text-right tabular-nums">
              {formatPriceRange(
                product.min_price_cents,
                product.max_price_cents,
              )}
            </TableCell>
            <TableCell className="text-muted-foreground">
              {formatDateTime(product.updated_at)}
            </TableCell>
            <TableCell className="text-right">
              <DropdownMenu>
                <DropdownMenuTrigger
                  aria-label={t("products.rowActions")}
                  className={buttonVariants({
                    variant: "ghost",
                    size: "icon-sm",
                  })}
                >
                  <MoreHorizontal aria-hidden />
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem onClick={() => onEdit(product)}>
                    <Pencil aria-hidden />
                    {t("products.actions.edit")}
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    variant="destructive"
                    onClick={() => onDelete(product)}
                  >
                    <Trash2 aria-hidden />
                    {t("products.actions.delete")}
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
