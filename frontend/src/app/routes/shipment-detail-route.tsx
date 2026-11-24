import { useParams } from "react-router"
import { useTranslation } from "react-i18next"
import { useAuth } from "@/features/auth"
import {
  ShipmentDetailPage,
  ShipmentSignedOut,
} from "@/features/logistics"
import { SiteFooter } from "../layouts/site-footer"
import { SiteHeader } from "../layouts/site-header"

export function ShipmentDetailRoute() {
  const { t } = useTranslation("logistics")
  const { session } = useAuth()
  const { shipmentNo } = useParams()

  return (
    <div className="flex min-h-svh flex-col">
      <SiteHeader />
      <main className="mx-auto w-full max-w-6xl flex-1 px-4 py-8 sm:px-6">
        {session ? (
          <ShipmentDetailPage
            token={session.token}
            shipmentNo={shipmentNo ?? ""}
            scope="buyer"
            backTo="/orders"
            backLabel={t("actions.backToOrders")}
          />
        ) : (
          <ShipmentSignedOut />
        )}
      </main>
      <SiteFooter />
    </div>
  )
}
