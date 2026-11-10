import {
  Boxes,
  CircleDollarSign,
  Package,
  Warehouse,
} from "lucide-react"
import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { PageHeader } from "@/components/merchant/page-header"
import { QueryError } from "@/components/merchant/query-error"
import { StatCard } from "@/components/merchant/stat-card"
import { Badge } from "@/components/ui/badge"
import { buttonVariants } from "@/components/ui/button"
import {
  Card,
  CardAction,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { getDashboard, getShopProfile } from "@/lib/merchant-api"
import { useDocumentTitle } from "@/lib/use-document-title"
import { useMerchantSession } from "@/lib/use-merchant-session"
import { useRequest } from "@/lib/use-request"
import { formatPriceCents } from "@/lib/format"
import { cn } from "@/lib/utils"

export default function DashboardPage() {
  const { t } = useTranslation("merchant")
  useDocumentTitle(t("documentTitle.dashboard"))

  const { token } = useMerchantSession()
  const dashboard = useRequest(() => getDashboard(token), [token])
  const shop = useRequest(() => getShopProfile(token), [token])

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
        <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
          {Array.from({ length: 4 }).map((_, index) => (
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
      ) : null}

      {dashboard.data ? (
        <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
          <StatCard
            label={t("dashboard.stats.products")}
            value={dashboard.data.product_count}
            hint={t("dashboard.stats.productsHint", {
              onShelf: dashboard.data.on_shelf_product_count,
            })}
            icon={Package}
          />
          <StatCard
            label={t("dashboard.stats.skus")}
            value={dashboard.data.sku_count}
            hint={t("dashboard.stats.skusHint", {
              count: dashboard.data.category_count,
            })}
            icon={Boxes}
          />
          <StatCard
            label={t("dashboard.stats.stock")}
            value={dashboard.data.total_stock}
            hint={t("dashboard.stats.stockHint", {
              count: dashboard.data.out_of_stock_sku_count,
            })}
            icon={Warehouse}
          />
          <StatCard
            label={t("dashboard.stats.value")}
            value={formatPriceCents(dashboard.data.inventory_value_cents)}
            hint={t("dashboard.stats.valueHint")}
            icon={CircleDollarSign}
          />
        </div>
      ) : null}

      {dashboard.data ? (
        <Card>
          <CardHeader>
            <CardTitle>{t("dashboard.shopCard.title")}</CardTitle>
            <CardAction>
              <Link
                to="/merchant/shop"
                className={cn(
                  buttonVariants({ variant: "ghost", size: "sm" }),
                )}
              >
                {t("dashboard.shopCard.edit")}
              </Link>
            </CardAction>
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
                      shop.data.status === "active"
                        ? "secondary"
                        : "destructive"
                    }
                  >
                    {t(`status.${shop.data.status}`)}
                  </Badge>
                </div>
              </>
            ) : null}
          </CardContent>
        </Card>
      ) : null}
    </div>
  )
}
