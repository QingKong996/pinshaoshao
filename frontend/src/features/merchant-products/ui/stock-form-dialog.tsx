import { useEffect } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { AlertCircle, LoaderCircle } from "lucide-react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { Alert, AlertDescription, AlertTitle } from "@/shared/ui/alert"
import { Button } from "@/shared/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/shared/ui/dialog"
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/shared/ui/field"
import { Input } from "@/shared/ui/input"
import type { CatalogChanged, InventoryItem } from "../types"
import { useUpdateInventoryStock } from "../model/mutations"
import { describeConsoleError } from "../model/errors"
import {
  initialValues,
  stockFormSchema,
  translateValidationMessage,
  type FormValues,
} from "../model/stock-form"

export function StockFormDialog({
  token,
  onCatalogChanged,
  item,
  onClose,
}: {
  token: string
  onCatalogChanged: CatalogChanged
  item: InventoryItem | null
  onClose: () => void
}) {
  const { t } = useTranslation("console")
  const updateStockMutation = useUpdateInventoryStock(
    token,
    onCatalogChanged,
  )
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<FormValues>({
    resolver: zodResolver(stockFormSchema),
    defaultValues: { stock: "" },
  })

  useEffect(() => {
    if (item) {
      reset(initialValues(item.stock))
    }
  }, [item, reset])

  function closeDialog() {
    updateStockMutation.reset()
    onClose()
  }

  async function onSubmit(values: FormValues) {
    if (!item) {
      return
    }
    try {
      await updateStockMutation.mutateAsync({
        productID: item.product_id,
        skuID: item.sku_id,
        stock: Number(values.stock),
      })
      closeDialog()
    } catch {
      // the mutation state renders the error
    }
  }

  return (
    <Dialog
      open={item !== null}
      onOpenChange={(open) => {
        if (!open && !isSubmitting) {
          closeDialog()
        }
      }}
    >
      <DialogContent className="sm:max-w-sm">
        <DialogHeader>
          <DialogTitle>{t("stockForm.title")}</DialogTitle>
          <DialogDescription>
            {item
              ? t("stockForm.description", {
                  product: item.product_name,
                  specs: formatSpecs(item, t("inventory.defaultSpecs")),
                })
              : null}
          </DialogDescription>
        </DialogHeader>

        <form
          id="stock-form"
          noValidate
          onSubmit={handleSubmit(onSubmit)}
          className="flex flex-col gap-4"
        >
          {updateStockMutation.error ? (
            <Alert variant="destructive">
              <AlertCircle aria-hidden />
              <AlertTitle>{t("common.saveError")}</AlertTitle>
              <AlertDescription>
                {describeConsoleError(updateStockMutation.error, t)}
              </AlertDescription>
            </Alert>
          ) : null}

          <FieldGroup>
            <Field data-invalid={errors.stock ? true : undefined}>
              <FieldLabel htmlFor="stock-form-stock">
                {t("stockForm.fields.stock")}
              </FieldLabel>
              <Input
                id="stock-form-stock"
                type="number"
                inputMode="numeric"
                min={0}
                step={1}
                autoFocus
                aria-invalid={errors.stock ? true : undefined}
                {...register("stock")}
              />
              {item ? (
                <FieldDescription>
                  {t("stockForm.currentStock", { stock: item.stock })}
                </FieldDescription>
              ) : null}
              {errors.stock ? (
                <FieldError>
                  {translateValidationMessage(errors.stock.message, t)}
                </FieldError>
              ) : null}
            </Field>
          </FieldGroup>
        </form>

        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            disabled={isSubmitting}
            onClick={closeDialog}
          >
            {t("common.cancel")}
          </Button>
          <Button type="submit" form="stock-form" disabled={isSubmitting}>
            {isSubmitting ? (
              <>
                <LoaderCircle className="animate-spin" aria-hidden />
                {t("stockForm.submit.saving")}
              </>
            ) : (
              t("stockForm.submit.save")
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function formatSpecs(
  item: InventoryItem,
  emptyLabel: string,
) {
  const entries = Object.entries(item.specs).sort(([left], [right]) =>
    left.localeCompare(right),
  )
  if (entries.length === 0) {
    return emptyLabel
  }
  return entries.map(([key, value]) => `${key}: ${value}`).join(" / ")
}
