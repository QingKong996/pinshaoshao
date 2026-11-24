import { useState } from "react"
import { Plus } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Button } from "@/shared/ui/button"
import type { Shipment } from "../types"
import { ShipmentDispatchDialog } from "./shipment-dispatch-dialog"

export function ShipmentDispatchButton({
  token,
  onCreated,
}: {
  token: string
  onCreated?: (shipment: Shipment) => void
}) {
  const { t } = useTranslation("logistics")
  const [open, setOpen] = useState(false)

  return (
    <>
      <Button onClick={() => setOpen(true)}>
        <Plus aria-hidden />
        {t("merchant.newShipment")}
      </Button>
      <ShipmentDispatchDialog
        token={token}
        open={open}
        onOpenChange={setOpen}
        onCreated={onCreated}
      />
    </>
  )
}
