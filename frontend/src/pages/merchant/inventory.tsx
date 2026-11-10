import { useMemo, useState } from "react"
import { CheckCircle2, Search, Warehouse } from "lucide-react"
import { useTranslation } from "react-i18next"
import { EmptyState } from "@/components/merchant/empty-state"
import { InventoryTable } from "@/components/merchant/inventory/inventory-table"
import { StockAdjustDialog } from "@/components/merchant/inventory/stock-adjust-dialog"
import { PageHeader } from "@/components/merchant/page-header"
import { QueryError } from "@/components/merchant/query-error"
import { TableSkeleton } from "@/components/merchant/table-skeleton"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Card, CardContent } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { getInventory, type InventoryItem } from "@/lib/merchant-api"
import { useDocumentTitle } from "@/lib/use-document-title"
import { useMerchantSession } from "@/lib/use-merchant-session"
import { useRequest } from "@/lib/use-request"

const FILTERS = ["all", "out"] as const

type InventoryFilter = (typeof FILTERS)[number]

const FILTER_LABEL_KEYS = {
  all: "inventory.filter.all",
  out: "inventory.filter.out",
} as const

export default function InventoryPage() {
  const { t } = useTranslation("merchant")
  useDocumentTitle(t("documentTitle.inventory"))

  const { token } = useMerchantSession()
  const inventory = useRequest(() => getInventory(token), [token])
  const [search, setSearch] = useState("")
  const [filter, setFilter] = useState<InventoryFilter>("all")
  const [adjustTarget, setAdjustTarget] = useState<InventoryItem | null>(null)
  const [adjusted, setAdjusted] = useState(false)

  const items = useMemo(() => inventory.data ?? [], [inventory.data])

  const filteredItems = useMemo(() => {
    const query = search.trim().toLowerCase()
    return items.filter((item) => {
      if (filter === "out" && item.stock !== 0) {
        return false
      }
      if (!query) {
        return true
      }
      return [item.product_name, ...Object.values(item.specs)].some((value) =>
        value.toLowerCase().includes(query),
      )
    })
  }, [filter, items, search])

  function openAdjustDialog(item: InventoryItem) {
    setAdjusted(false)
    setAdjustTarget(item)
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("inventory.title")}
        description={t("inventory.description")}
      />

      {inventory.error ? (
        <QueryError error={inventory.error} onRetry={inventory.reload} />
      ) : null}

      {adjusted ? (
        <Alert>
          <CheckCircle2 aria-hidden />
          <AlertDescription>{t("inventory.adjusted")}</AlertDescription>
        </Alert>
      ) : null}

      <div className="flex flex-wrap items-center gap-2">
        <div className="relative w-full sm:max-w-xs">
          <Search
            className="pointer-events-none absolute top-1/2 left-2.5 size-4 -translate-y-1/2 text-muted-foreground"
            aria-hidden
          />
          <Input
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder={t("inventory.searchPlaceholder")}
            aria-label={t("inventory.searchPlaceholder")}
            className="pl-8"
          />
        </div>
        <Select
          value={filter}
          onValueChange={(value) => {
            if (value) {
              setFilter(value)
            }
          }}
        >
          <SelectTrigger className="w-36" aria-label={t("inventory.filterLabel")}>
            <SelectValue>{t(FILTER_LABEL_KEYS[filter])}</SelectValue>
          </SelectTrigger>
          <SelectContent>
            {FILTERS.map((value) => (
              <SelectItem key={value} value={value}>
                {t(FILTER_LABEL_KEYS[value])}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {inventory.data || !inventory.error ? (
        <Card>
          <CardContent>
            {inventory.loading && !inventory.data ? (
              <TableSkeleton columns={7} />
            ) : null}

            {inventory.data && items.length === 0 ? (
              <EmptyState
                icon={Warehouse}
                title={t("inventory.empty.title")}
                description={t("inventory.empty.description")}
                className="py-10"
              />
            ) : null}

            {inventory.data &&
            items.length > 0 &&
            filteredItems.length === 0 ? (
              <EmptyState
                icon={Search}
                title={t("inventory.emptyFiltered.title")}
                description={t("inventory.emptyFiltered.description")}
                className="py-10"
              />
            ) : null}

            {filteredItems.length > 0 ? (
              <InventoryTable
                items={filteredItems}
                onAdjust={openAdjustDialog}
              />
            ) : null}
          </CardContent>
        </Card>
      ) : null}

      {adjustTarget ? (
        <StockAdjustDialog
          open
          onOpenChange={(open) => {
            if (!open) {
              setAdjustTarget(null)
            }
          }}
          item={adjustTarget}
          onSaved={() => {
            setAdjustTarget(null)
            setAdjusted(true)
            inventory.reload()
          }}
        />
      ) : null}
    </div>
  )
}
