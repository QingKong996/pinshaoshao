import { MoreHorizontal, Pencil, Trash2 } from "lucide-react"
import { useTranslation } from "react-i18next"
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
import { formatDateTime } from "@/lib/format"
import type { Category } from "@/lib/merchant-api"

type CategoryTableProps = {
  categories: Category[]
  onRename: (category: Category) => void
  onDelete: (category: Category) => void
}

export function CategoryTable({
  categories,
  onRename,
  onDelete,
}: CategoryTableProps) {
  const { t } = useTranslation("merchant")

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>{t("categories.columns.name")}</TableHead>
          <TableHead>{t("categories.columns.productCount")}</TableHead>
          <TableHead>{t("categories.columns.createdAt")}</TableHead>
          <TableHead className="w-10">
            <span className="sr-only">{t("common.actions")}</span>
          </TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {categories.map((category) => (
          <TableRow key={category.id}>
            <TableCell className="font-medium">{category.name}</TableCell>
            <TableCell className="text-muted-foreground">
              {t("categories.productCount", {
                count: category.product_count,
              })}
            </TableCell>
            <TableCell className="text-muted-foreground">
              {formatDateTime(category.created_at)}
            </TableCell>
            <TableCell className="text-right">
              <DropdownMenu>
                <DropdownMenuTrigger
                  aria-label={t("common.actions")}
                  className={buttonVariants({
                    variant: "ghost",
                    size: "icon-sm",
                  })}
                >
                  <MoreHorizontal aria-hidden />
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem onClick={() => onRename(category)}>
                    <Pencil aria-hidden />
                    {t("categories.form.titleEdit")}
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    variant="destructive"
                    onClick={() => onDelete(category)}
                  >
                    <Trash2 aria-hidden />
                    {t("common.delete")}
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
