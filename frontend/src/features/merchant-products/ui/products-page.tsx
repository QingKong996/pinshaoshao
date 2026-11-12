import { useMemo, useState } from "react"
import {
  AlertCircle,
  ArrowDownToLine,
  ArrowUpFromLine,
  Boxes,
  LoaderCircle,
  MoreHorizontal,
  Pencil,
  Plus,
  Search,
  Trash2,
} from "lucide-react"
import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { PageHeader } from "@/shared/ui/page-header"
import { ProductFormDialog } from "./product-form-dialog"
import { ProductStatusBadge } from "./product-status-badge"
import { QueryError } from "./query-error"
import { Alert, AlertDescription, AlertTitle } from "@/shared/ui/alert"
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/shared/ui/alert-dialog"
import { Button } from "@/shared/ui/button"
import { Card } from "@/shared/ui/card"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/shared/ui/dropdown-menu"
import { Input } from "@/shared/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/shared/ui/select"
import { Skeleton } from "@/shared/ui/skeleton"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/shared/ui/table"
import { formatPriceRange } from "@/shared/lib/format"
import type { Product, CatalogChanged } from "../types"
import {
  useDeleteProduct,
  useUpdateProduct,
} from "../model/mutations"
import { describeConsoleError } from "../model/errors"
import { useProducts, useProductCategories } from "../model/queries"
import {
  ALL,
  UNCATEGORIZED,
  filterProducts,
  type StatusFilter,
} from "../model/filters"
import { useDocumentTitle } from "@/shared/hooks/use-document-title"

const COLUMN_COUNT = 7

type ProductFormState = {
  open: boolean
  product: Product | null
}

export default function MerchantConsoleProductsPage({
  token,
  productHref,
  onCatalogChanged,
}: {
  token: string
  productHref: (productID: number) => string
  onCatalogChanged: CatalogChanged
}) {
  const { t } = useTranslation("console")
  useDocumentTitle(t("documentTitle.products"))

  const products = useProducts(token)
  const categories = useProductCategories(token)
  const updateProductMutation = useUpdateProduct(token, onCatalogChanged)
  const deleteProductMutation = useDeleteProduct(token, onCatalogChanged)

  const [search, setSearch] = useState("")
  const [statusFilter, setStatusFilter] = useState<StatusFilter>(ALL)
  const [categoryFilter, setCategoryFilter] = useState(ALL)
  const [form, setForm] = useState<ProductFormState>({
    open: false,
    product: null,
  })
  const [deleteTarget, setDeleteTarget] = useState<Product | null>(null)

  const actionError =
    updateProductMutation.error ?? deleteProductMutation.error

  const filteredProducts = useMemo(
    () => filterProducts(products.data ?? [], search, statusFilter, categoryFilter),
    [products.data, search, statusFilter, categoryFilter],
  )

  const statusItems = [
    { value: ALL, label: t("products.filters.allStatuses") },
    { value: "on_shelf", label: t("productStatus.on_shelf") },
    { value: "off_shelf", label: t("productStatus.off_shelf") },
  ]
  const categoryItems = [
    { value: ALL, label: t("products.filters.allCategories") },
    { value: UNCATEGORIZED, label: t("products.filters.uncategorized") },
    ...(categories.data ?? []).map((category) => ({
      value: String(category.id),
      label: category.name,
    })),
  ]

  function openCreateDialog() {
    setForm({ open: true, product: null })
  }

  function openEditDialog(product: Product) {
    setForm({ open: true, product })
  }

  function toggleShelfStatus(product: Product) {
    updateProductMutation.mutate({
      productID: product.id,
      input: {
        name: product.name,
        description: product.description,
        category_id: product.category_id,
        status: product.status === "on_shelf" ? "off_shelf" : "on_shelf",
      },
    })
  }

  function confirmDelete() {
    if (!deleteTarget) {
      return
    }
    deleteProductMutation.mutate(deleteTarget.id, {
      onSettled: () => setDeleteTarget(null),
    })
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("products.title")}
        description={t("products.description")}
        actions={
          <Button onClick={openCreateDialog}>
            <Plus aria-hidden />
            {t("products.new")}
          </Button>
        }
      />

      {products.error && !products.data ? (
        <QueryError error={products.error} onRetry={products.refetch} />
      ) : (
        <Card className="gap-0 overflow-hidden p-0">
          <div className="flex flex-col gap-2 border-b p-4 lg:flex-row lg:items-center">
            <div className="relative w-full lg:max-w-xs">
              <Search
                className="pointer-events-none absolute top-1/2 left-2.5 size-4 -translate-y-1/2 text-muted-foreground"
                aria-hidden
              />
              <Input
                value={search}
                onChange={(event) => setSearch(event.target.value)}
                placeholder={t("products.searchPlaceholder")}
                aria-label={t("products.searchPlaceholder")}
                className="pl-8"
              />
            </div>
            <div className="flex flex-wrap items-center gap-2 lg:ml-auto">
              <Select
                items={statusItems}
                value={statusFilter}
                onValueChange={(value) => {
                  if (value !== null) {
                    setStatusFilter(value)
                  }
                }}
              >
                <SelectTrigger
                  className="w-36"
                  aria-label={t("products.table.status")}
                >
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {statusItems.map((item) => (
                    <SelectItem key={item.value} value={item.value}>
                      {item.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <Select
                items={categoryItems}
                value={categoryFilter}
                onValueChange={(value) => {
                  if (value !== null) {
                    setCategoryFilter(value)
                  }
                }}
              >
                <SelectTrigger
                  className="w-44"
                  aria-label={t("products.table.category")}
                >
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {categoryItems.map((item) => (
                    <SelectItem key={item.value} value={item.value}>
                      {item.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          {actionError ? (
            <div className="px-4 pt-4">
              <Alert variant="destructive">
                <AlertCircle aria-hidden />
                <AlertTitle>{t("common.actionError")}</AlertTitle>
                <AlertDescription>
                  {describeConsoleError(actionError, t)}
                </AlertDescription>
              </Alert>
            </div>
          ) : null}

          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="px-4">
                  {t("products.table.product")}
                </TableHead>
                <TableHead className="px-4">
                  {t("products.table.category")}
                </TableHead>
                <TableHead className="px-4">
                  {t("products.table.status")}
                </TableHead>
                <TableHead className="px-4">
                  {t("products.table.skus")}
                </TableHead>
                <TableHead className="px-4">
                  {t("products.table.stock")}
                </TableHead>
                <TableHead className="px-4">
                  {t("products.table.price")}
                </TableHead>
                <TableHead className="px-4 text-right">
                  {t("products.table.actions")}
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {products.isPending ? (
                <ProductTableSkeleton />
              ) : filteredProducts.length === 0 ? (
                <TableRow>
                  <TableCell
                    colSpan={COLUMN_COUNT}
                    className="px-4 py-16 text-center whitespace-normal"
                  >
                    {products.data && products.data.length === 0 ? (
                      <>
                        <p className="font-medium">{t("products.empty.title")}</p>
                        <p className="mt-1 text-sm text-muted-foreground">
                          {t("products.empty.description")}
                        </p>
                      </>
                    ) : (
                      <p className="text-sm text-muted-foreground">
                        {t("products.empty.filtered")}
                      </p>
                    )}
                  </TableCell>
                </TableRow>
              ) : (
                filteredProducts.map((product) => (
                  <TableRow key={product.id}>
                    <TableCell className="px-4 py-3">
                      <div className="flex max-w-72 flex-col gap-0.5">
                        <Link
                          to={productHref(product.id)}
                          className="truncate font-medium hover:underline"
                        >
                          {product.name}
                        </Link>
                        {product.description ? (
                          <span className="truncate text-xs text-muted-foreground">
                            {product.description}
                          </span>
                        ) : null}
                      </div>
                    </TableCell>
                    <TableCell className="px-4 py-3 text-muted-foreground">
                      {product.category_name ??
                        t("products.filters.uncategorized")}
                    </TableCell>
                    <TableCell className="px-4 py-3">
                      <ProductStatusBadge status={product.status} />
                    </TableCell>
                    <TableCell className="px-4 py-3 tabular-nums">
                      {product.sku_count}
                    </TableCell>
                    <TableCell className="px-4 py-3 tabular-nums">
                      {product.total_stock}
                    </TableCell>
                    <TableCell className="px-4 py-3 tabular-nums">
                      {formatPriceRange(
                        product.min_price_cents,
                        product.max_price_cents,
                      ) ?? "—"}
                    </TableCell>
                    <TableCell className="px-4 py-3 text-right">
                      <DropdownMenu>
                        <DropdownMenuTrigger
                          render={
                            <Button
                              variant="ghost"
                              size="icon-sm"
                              aria-label={t("products.table.actions")}
                            />
                          }
                        >
                          <MoreHorizontal aria-hidden />
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem
                            onClick={() => openEditDialog(product)}
                          >
                            <Pencil aria-hidden />
                            {t("products.actions.edit")}
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            render={
                              <Link
                                to={productHref(product.id)}
                              />
                            }
                          >
                            <Boxes aria-hidden />
                            {t("products.actions.manageSkus")}
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem
                            onClick={() => void toggleShelfStatus(product)}
                          >
                            {product.status === "on_shelf" ? (
                              <ArrowDownToLine aria-hidden />
                            ) : (
                              <ArrowUpFromLine aria-hidden />
                            )}
                            {product.status === "on_shelf"
                              ? t("products.actions.takeOffShelf")
                              : t("products.actions.putOnShelf")}
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem
                            variant="destructive"
                            onClick={() => setDeleteTarget(product)}
                          >
                            <Trash2 aria-hidden />
                            {t("products.actions.delete")}
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </Card>
      )}

      <ProductFormDialog
        token={token}
        onCatalogChanged={onCatalogChanged}
        open={form.open}
        onOpenChange={(open) =>
          setForm((current) => ({ ...current, open }))
        }
        product={form.product ?? undefined}
        categories={categories.data ?? []}
      />

      <AlertDialog
        open={deleteTarget !== null}
        onOpenChange={(open) => {
          if (!open && !deleteProductMutation.isPending) {
            setDeleteTarget(null)
          }
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t("products.deleteDialog.title")}</AlertDialogTitle>
            <AlertDialogDescription>
              {t("products.deleteDialog.description", {
                name: deleteTarget?.name ?? "",
              })}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={deleteProductMutation.isPending}>
              {t("common.cancel")}
            </AlertDialogCancel>
            <Button
              variant="destructive"
              disabled={deleteProductMutation.isPending}
              onClick={confirmDelete}
            >
              {deleteProductMutation.isPending ? (
                <LoaderCircle className="animate-spin" aria-hidden />
              ) : null}
              {t("products.deleteDialog.confirm")}
            </Button>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}

function ProductTableSkeleton() {
  return (
    <>
      {Array.from({ length: 5 }, (_, index) => (
        <TableRow key={index}>
          <TableCell className="px-4 py-3">
            <div className="flex flex-col gap-1.5">
              <Skeleton className="h-4 w-40" />
              <Skeleton className="h-3 w-56" />
            </div>
          </TableCell>
          <TableCell className="px-4 py-3">
            <Skeleton className="h-4 w-20" />
          </TableCell>
          <TableCell className="px-4 py-3">
            <Skeleton className="h-5 w-16" />
          </TableCell>
          <TableCell className="px-4 py-3">
            <Skeleton className="h-4 w-8" />
          </TableCell>
          <TableCell className="px-4 py-3">
            <Skeleton className="h-4 w-8" />
          </TableCell>
          <TableCell className="px-4 py-3">
            <Skeleton className="h-4 w-24" />
          </TableCell>
          <TableCell className="px-4 py-3">
            <Skeleton className="ml-auto size-7" />
          </TableCell>
        </TableRow>
      ))}
    </>
  )
}
