import { ChevronLeftIcon, ChevronRightIcon } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Button } from "@/shared/ui/button"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/shared/ui/select"
import { PAGE_SIZE_OPTIONS } from "../model/filters"

export function ProductsPagination({
  page,
  pageSize,
  total,
  totalPages,
  onPageChange,
  onPageSizeChange,
}: {
  page: number
  pageSize: number
  total: number
  totalPages: number
  onPageChange: (page: number) => void
  onPageSizeChange: (pageSize: number) => void
}) {
  const { t } = useTranslation("console")
  const visibleTotalPages = Math.max(1, totalPages)
  const visiblePage = Math.min(page, visibleTotalPages)
  const firstItem =
    total === 0 ? 0 : (visiblePage - 1) * pageSize + 1
  const lastItem = Math.min(visiblePage * pageSize, total)
  const pageSizeItems = PAGE_SIZE_OPTIONS.map((size) => ({
    value: String(size),
    label: String(size),
  }))

  return (
    <div className="flex flex-col gap-3 border-t px-4 py-3 text-sm text-muted-foreground sm:flex-row sm:items-center">
      <p aria-live="polite">
        {t("products.pagination.summary", {
          start: firstItem,
          end: lastItem,
          total,
        })}
      </p>
      <div className="flex flex-wrap items-center gap-2 sm:ml-auto">
        <span>{t("products.pagination.rowsPerPage")}</span>
        <Select
          items={pageSizeItems}
          value={String(pageSize)}
          onValueChange={(value) => {
            if (value !== null) {
              onPageSizeChange(Number(value))
            }
          }}
        >
          <SelectTrigger
            size="sm"
            className="w-16"
            aria-label={t("products.pagination.rowsPerPage")}
          >
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {pageSizeItems.map((item) => (
              <SelectItem key={item.value} value={item.value}>
                {item.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Button
          variant="outline"
          size="icon-sm"
          aria-label={t("products.pagination.previous")}
          disabled={visiblePage <= 1}
          onClick={() => onPageChange(visiblePage - 1)}
        >
          <ChevronLeftIcon aria-hidden />
        </Button>
        <span className="min-w-16 text-center tabular-nums">
          {t("products.pagination.pageOf", {
            page: visiblePage,
            totalPages: visibleTotalPages,
          })}
        </span>
        <Button
          variant="outline"
          size="icon-sm"
          aria-label={t("products.pagination.next")}
          disabled={totalPages === 0 || visiblePage >= totalPages}
          onClick={() => onPageChange(visiblePage + 1)}
        >
          <ChevronRightIcon aria-hidden />
        </Button>
      </div>
    </div>
  )
}
