import { useState } from "react"
import { Truck } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Button } from "@/shared/ui/button"
import type { Shipment } from "../types"
import { ShipmentDispatchDialog } from "./shipment-dispatch-dialog"

export function ShipSubOrderButton({
  token,
  subOrderID,
  receiverName,
  onCreated,
}: {
  token: string
  subOrderID: number
  receiverName?: string
  onCreated?: (shipment: Shipment) => void
}) {
  const { t } = useTranslation("logistics")
  const [open, setOpen] = useState(false)

  return (
    <>
      <Button
        variant="outline"
        size="sm"
        onClick={() => setOpen(true)}
      >
        <Truck aria-hidden />
        {t("actions.ship")}
      </Button>
      <ShipmentDispatchDialog
        token={token}
        open={open}
        onOpenChange={setOpen}
        subOrderID={subOrderID}
        receiverName={receiverName}
        onCreated={onCreated}
      />
    </>
  )
}
