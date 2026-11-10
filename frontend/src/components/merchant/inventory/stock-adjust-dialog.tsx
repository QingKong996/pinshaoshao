import { useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { AlertCircle, LoaderCircle } from "lucide-react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { formatSpecs } from "@/components/merchant/skus/spec-format"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Field, FieldError, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { updateProductSKU, type InventoryItem } from "@/lib/merchant-api"
import { describeMerchantError } from "@/lib/merchant-errors"
import {
  merchantStockSchema,
  translateMerchantValidationMessage,
  type MerchantStockValues,
} from "@/lib/merchant-schemas"
import { useMerchantSession } from "@/lib/use-merchant-session"

type StockAdjustDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  item: InventoryItem
  onSaved: () => void
}

export function StockAdjustDialog({
  open,
  onOpenChange,
  item,
  onSaved,
}: StockAdjustDialogProps) {
  const { t } = useTranslation("merchant")
  const { token } = useMerchantSession()
  const [pending, setPending] = useState(false)
  const [saveError, setSaveError] = useState<unknown>(null)

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<MerchantStockValues>({
    resolver: zodResolver(merchantStockSchema),
    defaultValues: { stock: String(item.stock) },
  })

  const specs = formatSpecs(item.specs)
  const displayName = specs
    ? `${item.product_name} · ${specs}`
    : item.product_name

  async function onSubmit(values: MerchantStockValues) {
    setPending(true)
    setSaveError(null)
    try {
      await updateProductSKU(token, item.product_id, item.sku_id, {
        stock: Number(values.stock),
      })
      onSaved()
    } catch (error) {
      setSaveError(error)
      setPending(false)
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(nextOpen) => {
        if (!nextOpen && pending) {
          return
        }
        onOpenChange(nextOpen)
      }}
    >
      <DialogContent closeLabel={t("common.close")}>
        <DialogHeader>
          <DialogTitle>{t("inventory.dialog.title")}</DialogTitle>
          <DialogDescription>
            {t("inventory.dialog.description", { name: displayName })}
          </DialogDescription>
        </DialogHeader>

        <form
          noValidate
          aria-busy={pending}
          onSubmit={handleSubmit(onSubmit)}
          className="flex flex-col gap-4"
        >
          {saveError ? (
            <Alert variant="destructive">
              <AlertCircle aria-hidden />
              <AlertDescription>
                {describeMerchantError(saveError, t)}
              </AlertDescription>
            </Alert>
          ) : null}

          <Field data-invalid={errors.stock ? true : undefined}>
            <FieldLabel htmlFor="inventory-stock">
              {t("inventory.dialog.stockLabel")}
            </FieldLabel>
            <Input
              id="inventory-stock"
              inputMode="numeric"
              autoComplete="off"
              aria-invalid={errors.stock ? true : undefined}
              aria-describedby={
                errors.stock ? "inventory-stock-error" : undefined
              }
              className="h-10"
              {...register("stock")}
            />
            {errors.stock ? (
              <FieldError id="inventory-stock-error">
                {translateMerchantValidationMessage(errors.stock.message, t)}
              </FieldError>
            ) : null}
          </Field>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={pending}
            >
              {t("common.cancel")}
            </Button>
            <Button type="submit" disabled={pending} className="gap-2">
              {pending ? (
                <LoaderCircle className="animate-spin" aria-hidden />
              ) : null}
              {t("inventory.dialog.submit")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
