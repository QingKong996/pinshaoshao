import { useEffect } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
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
import {
  initialValues,
  stockFormSchema,
  translateValidationMessage,
  type FormValues,
} from "../model/stock-form"
import { formatSpecs } from "./format-specs"
import { FormDialog } from "./form-dialog"

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
    <FormDialog
      open={item !== null}
      onOpenChange={(open) => {
        if (!open) {
          closeDialog()
        }
      }}
      title={t("stockForm.title")}
      description={
        item
          ? t("stockForm.description", {
              product: item.product_name,
              specs: formatSpecs(
                item.specs,
                t("inventory.defaultSpecs"),
              ),
            })
          : undefined
      }
      formId="stock-form"
      onSubmit={handleSubmit(onSubmit)}
      error={updateStockMutation.error}
      isSubmitting={isSubmitting}
      submitLabel={t("stockForm.submit.save")}
      submittingLabel={t("stockForm.submit.saving")}
      className="sm:max-w-sm"
    >
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
    </FormDialog>
  )
}
