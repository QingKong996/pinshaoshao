import { useCallback, useMemo, useState } from "react"
import {
  createColumnHelper,
  tableFeatures,
  useTable,
} from "@tanstack/react-table"
import { cn } from "cn"
import { AlertCircle, Pencil, Plus, Trash2 } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Alert, AlertDescription, AlertTitle } from "@/shared/ui/alert"
import { Badge } from "@/shared/ui/badge"
import { Button } from "@/shared/ui/button"
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/shared/ui/card"
import { Skeleton } from "@/shared/ui/skeleton"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/shared/ui/table"
import { formatPriceCents } from "@/shared/lib/format"
import type { ProductSKU, CatalogChanged } from "../types"
import { useDeleteProductSKU } from "../model/mutations"
import { describeConsoleError } from "../model/errors"
import { useProductSKUs } from "../model/queries"
import { DeleteSKUDialog } from "./delete-sku-dialog"
import { formatSpecsLabel } from "./format-specs"
import { QueryError } from "./query-error"
import { SKUFormDialog } from "./sku-form-dialog"

const EMPTY_SKUS: ProductSKU[] = []
const skuFeatures = tableFeatures({
  columnMeta: {} as {
    cellClassName?: string
    headerClassName?: string
  },
})
const skuColumnHelper = createColumnHelper<
  typeof skuFeatures,
  ProductSKU
>()

export function SKUSection({
  token,
  productID,
  onCatalogChanged,
}: {
  token: string
  productID: number
  onCatalogChanged: CatalogChanged
}) {
  const { t } = useTranslation("console")

  const skus = useProductSKUs(token, productID)
  const deleteProductSKUMutation = useDeleteProductSKU(token, onCatalogChanged)

  const [skuDialogOpen, setSkuDialogOpen] = useState(false)
  const [editingSKU, setEditingSKU] = useState<ProductSKU | null>(null)
  const [deletingSKU, setDeletingSKU] = useState<ProductSKU | null>(null)

  const actionError = deleteProductSKUMutation.error

  function openCreateSKU() {
    setEditingSKU(null)
    setSkuDialogOpen(true)
  }

  const openEditSKU = useCallback((sku: ProductSKU) => {
    setEditingSKU(sku)
    setSkuDialogOpen(true)
  }, [])

  function closeSKUDialog() {
    setSkuDialogOpen(false)
    setEditingSKU(null)
  }

  function handleDeleteSKU() {
    if (!deletingSKU) {
      return
    }
    deleteProductSKUMutation.mutate(
      { productID, skuID: deletingSKU.id },
      { onSettled: () => setDeletingSKU(null) },
    )
  }

  const columns = useMemo(
    () =>
      skuColumnHelper.columns([
        skuColumnHelper.accessor("id", {
          id: "specs",
          header: t("productDetail.skus.table.specs"),
          meta: { cellClassName: "whitespace-normal" },
          cell: ({ row }) =>
            formatSpecsLabel(
              row.original,
              t("productDetail.skus.defaultSpecs"),
            ),
        }),
        skuColumnHelper.accessor("price_cents", {
          header: t("productDetail.skus.table.price"),
          meta: { cellClassName: "tabular-nums" },
          cell: ({ getValue }) => formatPriceCents(getValue()),
        }),
        skuColumnHelper.accessor("stock", {
          header: t("productDetail.skus.table.stock"),
          meta: { cellClassName: "tabular-nums" },
        }),
        skuColumnHelper.accessor("is_enabled", {
          header: t("productDetail.skus.table.enabled"),
          cell: ({ getValue }) => {
            const isEnabled = getValue()
            return (
              <Badge variant={isEnabled ? "secondary" : "outline"}>
                {isEnabled
                  ? t("productDetail.skus.enabled")
                  : t("productDetail.skus.disabled")}
              </Badge>
            )
          },
        }),
        skuColumnHelper.display({
          id: "actions",
          header: t("productDetail.skus.table.actions"),
          meta: {
            headerClassName: "text-right",
            cellClassName: "text-right",
          },
          cell: ({ row }) => (
            <div className="flex justify-end gap-1">
              <Button
                variant="ghost"
                size="icon-sm"
                aria-label={t("productDetail.skus.actions.edit")}
                onClick={() => openEditSKU(row.original)}
              >
                <Pencil aria-hidden />
              </Button>
              <Button
                variant="ghost"
                size="icon-sm"
                aria-label={t("productDetail.skus.actions.delete")}
                onClick={() => setDeletingSKU(row.original)}
              >
                <Trash2 aria-hidden />
              </Button>
            </div>
          ),
        }),
      ]),
    [t, openEditSKU],
  )

  const table = useTable({
    features: skuFeatures,
    columns,
    data: skus.data ?? EMPTY_SKUS,
    getRowId: (row) => String(row.id),
  })
  const rows = table.getRowModel().rows
  const columnCount = table.getAllLeafColumns().length

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>{t("productDetail.skus.title")}</CardTitle>
          <CardDescription>
            {t("productDetail.skus.description")}
          </CardDescription>
          <CardAction>
            <Button
              size="sm"
              onClick={openCreateSKU}
              disabled={!skus.data}
            >
              <Plus aria-hidden />
              {t("productDetail.skus.add")}
            </Button>
          </CardAction>
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          {actionError ? (
            <Alert variant="destructive">
              <AlertCircle aria-hidden />
              <AlertTitle>{t("common.actionError")}</AlertTitle>
              <AlertDescription>
                {describeConsoleError(actionError, t)}
              </AlertDescription>
            </Alert>
          ) : null}

          {skus.error ? (
            <QueryError error={skus.error} onRetry={skus.refetch} />
          ) : null}

          <Table>
            <TableHeader>
              {table.getHeaderGroups().map((headerGroup) => (
                <TableRow key={headerGroup.id}>
                  {headerGroup.headers.map((header) => (
                    <TableHead
                      key={header.id}
                      className={cn(
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
              {skus.isPending ? <SKUTableSkeleton /> : null}

              {skus.data && rows.length === 0 ? (
                <TableRow className="hover:bg-transparent">
                  <TableCell colSpan={columnCount} className="py-8">
                    <div className="flex flex-col items-center gap-1 text-center">
                      <span className="font-medium">
                        {t("productDetail.skus.empty.title")}
                      </span>
                      <span className="text-sm text-muted-foreground">
                        {t("productDetail.skus.empty.description")}
                      </span>
                    </div>
                  </TableCell>
                </TableRow>
              ) : null}

              {rows.map((row) => (
                <TableRow key={row.id}>
                  {row.getAllCells().map((cell) => (
                    <TableCell
                      key={cell.id}
                      className={cn(
                        cell.column.columnDef.meta?.cellClassName,
                      )}
                    >
                      <table.FlexRender cell={cell} />
                    </TableCell>
                  ))}
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {skuDialogOpen ? (
        <SKUFormDialog
          token={token}
          onCatalogChanged={onCatalogChanged}
          key={editingSKU ? `edit-${editingSKU.id}` : "create"}
          open
          onOpenChange={(open) => {
            if (!open) {
              closeSKUDialog()
            }
          }}
          productID={productID}
          sku={editingSKU ?? undefined}
          skus={skus.data ?? []}
        />
      ) : null}

      <DeleteSKUDialog
        sku={deletingSKU}
        pending={deleteProductSKUMutation.isPending}
        onOpenChange={(open) => {
          if (!open && !deleteProductSKUMutation.isPending) {
            setDeletingSKU(null)
          }
        }}
        onConfirm={handleDeleteSKU}
      />
    </>
  )
}

function SKUTableSkeleton() {
  return (
    <>
      {Array.from({ length: 3 }, (_, index) => (
        <TableRow key={index}>
          <TableCell>
            <Skeleton className="h-4 w-32" />
          </TableCell>
          <TableCell>
            <Skeleton className="h-4 w-16" />
          </TableCell>
          <TableCell>
            <Skeleton className="h-4 w-10" />
          </TableCell>
          <TableCell>
            <Skeleton className="h-5 w-16" />
          </TableCell>
          <TableCell>
            <Skeleton className="ml-auto h-6 w-16" />
          </TableCell>
        </TableRow>
      ))}
    </>
  )
}
