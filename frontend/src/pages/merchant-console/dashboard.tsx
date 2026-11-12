import {
  Boxes,
  CircleDollarSign,
  Package,
  Warehouse,
} from "lucide-react"
import { useTranslation } from "react-i18next"
import { PageHeader } from "@/components/merchant-console/page-header"
import { QueryError } from "@/components/merchant-console/query-error"
import { StatCard } from "@/components/merchant-console/stat-card"
import { Badge } from "@/components/ui/badge"
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { formatPriceCents } from "@/lib/format"
import {
  getDashboard,
  getShopProfile,
  type Dashboard,
  type ShopProfile,
} from "@/api/merchant-console"
import { useDocumentTitle } from "@/hooks/use-document-title"
import { useMerchantSession } from "@/hooks/use-merchant-session"
import { useRequest, type RequestState } from "@/hooks/use-request"

export default function MerchantConsoleDashboardPage() {
  const { t } = useTranslation("console")
  useDocumentTitle(t("documentTitle.dashboard"))

  const { token } = useMerchantSession()
  const dashboard = useRequest(() => getDashboard(token), token)
  const shop = useRequest(() => getShopProfile(token), token)

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("dashboard.title")}
        description={t("dashboard.description")}
      />

      {dashboard.error ? (
        <QueryError error={dashboard.error} onRetry={dashboard.reload} />
      ) : null}

      {dashboard.loading && !dashboard.data ? (
        <DashboardStatsSkeleton />
      ) : null}

      {dashboard.data ? <DashboardStats data={dashboard.data} /> : null}

      <ShopProfileCard shop={shop} />
    </div>
  )
}

function DashboardStats({ data }: { data: Dashboard }) {
  const { t } = useTranslation("console")

  return (
    <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
      <StatCard
        label={t("dashboard.stats.products")}
        value={data.product_count}
        hint={t("dashboard.stats.productsHint", {
          onShelf: data.on_shelf_product_count,
        })}
        icon={Package}
      />
      <StatCard
        label={t("dashboard.stats.skus")}
        value={data.sku_count}
        hint={t("dashboard.stats.skusHint", {
          count: data.category_count,
        })}
        icon={Boxes}
      />
      <StatCard
        label={t("dashboard.stats.stock")}
        value={data.total_stock}
        hint={t("dashboard.stats.stockHint", {
          count: data.out_of_stock_sku_count,
        })}
        icon={Warehouse}
      />
      <StatCard
        label={t("dashboard.stats.value")}
        value={formatPriceCents(data.inventory_value_cents)}
        hint={t("dashboard.stats.valueHint")}
        icon={CircleDollarSign}
      />
    </div>
  )
}

function DashboardStatsSkeleton() {
  return (
    <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
      {Array.from({ length: 4 }, (_, index) => (
        <Card key={index} size="sm">
          <CardHeader>
            <Skeleton className="h-4 w-20" />
          </CardHeader>
          <CardContent className="flex flex-col gap-2">
            <Skeleton className="h-7 w-24" />
            <Skeleton className="h-3 w-32" />
          </CardContent>
        </Card>
      ))}
    </div>
  )
}

function ShopProfileCard({ shop }: { shop: RequestState<ShopProfile> }) {
  const { t } = useTranslation("console")

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("dashboard.shopCard.title")}</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-3">
        {shop.loading && !shop.data ? (
          <>
            <Skeleton className="h-5 w-32" />
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-2/3" />
          </>
        ) : null}
        {shop.error && !shop.data ? (
          <QueryError error={shop.error} onRetry={shop.reload} />
        ) : null}
        {shop.data ? (
          <>
            <div className="flex flex-col gap-1">
              <span className="font-medium">
                {shop.data.name || shop.data.username}
              </span>
              <span className="text-sm text-muted-foreground">
                {shop.data.description ||
                  t("dashboard.shopCard.noDescription")}
              </span>
            </div>
            <div className="flex items-center gap-2 text-sm">
              <span className="text-muted-foreground">
                {t("dashboard.shopCard.statusLabel")}
              </span>
              <Badge
                variant={
                  shop.data.status === "active" ? "secondary" : "destructive"
                }
              >
                {t(`status.${shop.data.status}`)}
              </Badge>
            </div>
          </>
        ) : null}
      </CardContent>
    </Card>
  )
}
