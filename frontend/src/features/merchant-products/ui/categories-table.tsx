import { useMemo } from "react"
import {
  createColumnHelper,
  tableFeatures,
  useTable,
} from "@tanstack/react-table"
import { cn } from "cn"
import { Pencil, Trash2 } from "lucide-react"
import { useTranslation } from "react-i18next"
import { formatDateTimeSeconds } from "@/shared/lib/format"
import { Button } from "@/shared/ui/button"
import { Skeleton } from "@/shared/ui/skeleton"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/shared/ui/table"
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/shared/ui/tooltip"
import type { ProductCategory } from "../types"

const categoryFeatures = tableFeatures({
  columnMeta: {} as {
    cellClassName?: string
    headerClassName?: string
  },
})
const categoryColumnHelper = createColumnHelper<
  typeof categoryFeatures,
  ProductCategory
>()

export function CategoriesTable({
  categories,
  isPending,
  isCatalogEmpty,
  onEdit,
  onDelete,
}: {
  categories: ProductCategory[]
  isPending: boolean
  isCatalogEmpty: boolean
  onEdit: (category: ProductCategory) => void
  onDelete: (category: ProductCategory) => void
}) {
  const { t } = useTranslation("console")

  const columns = useMemo(
    () =>
      categoryColumnHelper.columns([
        categoryColumnHelper.accessor("name", {
          header: t("categories.table.name"),
          cell: ({ getValue }) => (
            <span className="font-medium">{getValue()}</span>
          ),
        }),
        categoryColumnHelper.accessor("product_count", {
          header: t("categories.table.productCount"),
          meta: { cellClassName: "tabular-nums" },
        }),
        categoryColumnHelper.accessor("created_at", {
          header: t("categories.table.createdAt"),
          meta: {
            cellClassName: "text-muted-foreground tabular-nums",
          },
          cell: ({ getValue }) => formatDateTimeSeconds(getValue()),
        }),
        categoryColumnHelper.accessor("updated_at", {
          header: t("categories.table.updatedAt"),
          meta: {
            cellClassName: "text-muted-foreground tabular-nums",
          },
          cell: ({ getValue }) => formatDateTimeSeconds(getValue()),
        }),
        categoryColumnHelper.display({
          id: "actions",
          header: t("categories.table.actions"),
          meta: {
            headerClassName: "text-right",
            cellClassName: "text-right",
          },
          cell: ({ row }) => (
            <div className="flex items-center justify-end gap-2">
              <Tooltip>
                <TooltipTrigger
                  render={
                    <Button
                      variant="ghost"
                      size="icon-sm"
                      aria-label={t("categories.actions.edit")}
                      onClick={() => onEdit(row.original)}
                    />
                  }
                >
                  <Pencil aria-hidden />
                </TooltipTrigger>
                <TooltipContent>
                  {t("categories.actions.edit")}
                </TooltipContent>
              </Tooltip>
              <Tooltip>
                <TooltipTrigger
                  render={
                    <Button
                      variant="ghost"
                      size="icon-sm"
                      aria-label={t("categories.actions.delete")}
                      className="text-destructive hover:bg-destructive/10 hover:text-destructive"
                      onClick={() => onDelete(row.original)}
                    />
                  }
                >
                  <Trash2 aria-hidden />
                </TooltipTrigger>
                <TooltipContent>
                  {t("categories.actions.delete")}
                </TooltipContent>
              </Tooltip>
            </div>
          ),
        }),
      ]),
    [onDelete, onEdit, t],
  )

  const table = useTable({
    features: categoryFeatures,
    columns,
    data: categories,
    getRowId: (row) => String(row.id),
  })
  const rows = table.getRowModel().rows
  const columnCount = table.getAllLeafColumns().length

  return (
    <Table>
      <TableHeader>
        {table.getHeaderGroups().map((headerGroup) => (
          <TableRow key={headerGroup.id}>
            {headerGroup.headers.map((header) => (
              <TableHead
                key={header.id}
                className={cn(
                  "px-4",
                  header.column.columnDef.meta?.headerClassName,
                )}
              >
                {header.isPlaceholder ? null : (
                  <table.FlexRender header={header} />
                )}
              </TableHead>
            ))}
          </TableRow>
        ))}
      </TableHeader>
      <TableBody>
        {isPending ? (
          <CategoryTableSkeleton />
        ) : rows.length === 0 ? (
          <TableRow>
            <TableCell
              colSpan={columnCount}
              className="px-4 py-16 text-center whitespace-normal"
            >
              {isCatalogEmpty ? (
                <>
                  <p className="font-medium">{t("categories.empty.title")}</p>
                  <p className="mt-1 text-sm text-muted-foreground">
                    {t("categories.empty.description")}
                  </p>
                </>
              ) : (
                <p className="text-sm text-muted-foreground">
                  {t("categories.empty.filtered")}
                </p>
              )}
            </TableCell>
          </TableRow>
        ) : (
          rows.map((row) => (
            <TableRow key={row.id}>
              {row.getAllCells().map((cell) => (
                <TableCell
                  key={cell.id}
                  className={cn(
                    "px-4 py-3",
                    cell.column.columnDef.meta?.cellClassName,
                  )}
                >
                  <table.FlexRender cell={cell} />
                </TableCell>
              ))}
            </TableRow>
          ))
        )}
      </TableBody>
    </Table>
  )
}

function CategoryTableSkeleton() {
  return (
    <>
      {Array.from({ length: 4 }, (_, index) => (
        <TableRow key={index}>
          <TableCell className="px-4 py-3">
            <Skeleton className="h-4 w-32" />
          </TableCell>
          <TableCell className="px-4 py-3">
            <Skeleton className="h-4 w-10" />
          </TableCell>
          <TableCell className="px-4 py-3">
            <Skeleton className="h-4 w-32" />
          </TableCell>
          <TableCell className="px-4 py-3">
            <Skeleton className="h-4 w-32" />
          </TableCell>
          <TableCell className="px-4 py-3">
            <div className="flex justify-end gap-2">
              <Skeleton className="size-7" />
              <Skeleton className="size-7" />
            </div>
          </TableCell>
        </TableRow>
      ))}
    </>
  )
}
