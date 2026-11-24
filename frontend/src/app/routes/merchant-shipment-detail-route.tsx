import { useParams } from "react-router"
import { useTranslation } from "react-i18next"
import { useMerchantSession } from "@/features/auth"
import { ShipmentDetailPage } from "@/features/logistics"
import { MERCHANT_CONSOLE_PATHS } from "../merchant-navigation"

export function MerchantShipmentDetailRoute() {
  const { t } = useTranslation("logistics")
  const { token } = useMerchantSession()
  const { shipmentNo } = useParams()

  return (
    <ShipmentDetailPage
      token={token}
      shipmentNo={shipmentNo ?? ""}
      scope="merchant"
      backTo={MERCHANT_CONSOLE_PATHS.logistics}
      backLabel={t("actions.backToShipments")}
    />
  )
}
