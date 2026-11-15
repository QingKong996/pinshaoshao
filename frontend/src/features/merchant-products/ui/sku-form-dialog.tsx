import { useMemo } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { Controller, useFieldArray, useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/shared/ui/field"
import { Input } from "@/shared/ui/input"
import { Switch } from "@/shared/ui/switch"
import {
  type ProductSKU,
  type CatalogChanged,
} from "../types"
import {
  useCreateProductSKU,
  useUpdateProductSKU,
} from "../model/mutations"
import {
  buildFormSchema,
  translateValidationMessage,
  fixedSpecKeys,
  initialSpecs,
  skuToInput,
  type FormValues,
} from "../model/sku-form"
import { FormDialog } from "./form-dialog"
import { SKUSpecFields } from "./sku-spec-fields"

export type SKUFormDialogProps = {
  token: string
  onCatalogChanged: CatalogChanged
  open: boolean
  onOpenChange: (open: boolean) => void
  productID: number
  sku?: ProductSKU
  skus: ProductSKU[]
}

export function SKUFormDialog({
  token,
  onCatalogChanged,
  open,
  onOpenChange,
  productID,
  sku,
  skus,
}: SKUFormDialogProps) {
  const { t } = useTranslation("console")
  const createProductSKUMutation = useCreateProductSKU(token, onCatalogChanged)
  const updateProductSKUMutation = useUpdateProductSKU(token, onCatalogChanged)
  const saveError = sku
    ? updateProductSKUMutation.error
    : createProductSKUMutation.error

  // Spec keys are shared by every SKU of a product, so they stay fixed once
  // more than one SKU exists.
  const specsEditable = sku ? skus.length <= 1 : skus.length === 0
  const specKeys = useMemo(() => fixedSpecKeys(sku, skus), [sku, skus])
  const resolver = useMemo(
    () => zodResolver(buildFormSchema(specsEditable)),
    [specsEditable],
  )

  const {
    register,
    control,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<FormValues>({
    resolver,
    defaultValues: {
      specs: initialSpecs(sku, specsEditable, specKeys),
      price: sku ? String(sku.price_cents / 100) : "",
      stock: sku ? String(sku.stock) : "",
      is_enabled: sku?.is_enabled ?? true,
    },
  })

  const { fields, append, remove } = useFieldArray({ control, name: "specs" })

  function handleOpenChange(nextOpen: boolean) {
    if (!nextOpen) {
      createProductSKUMutation.reset()
      updateProductSKUMutation.reset()
    }
    onOpenChange(nextOpen)
  }

  async function onSubmit(values: FormValues) {
    const input = skuToInput(values)

    try {
      if (sku) {
        await updateProductSKUMutation.mutateAsync({
          productID,
          skuID: sku.id,
          input,
        })
      } else {
        await createProductSKUMutation.mutateAsync({ productID, input })
      }
      handleOpenChange(false)
    } catch {
      // the mutation state renders the error
    }
  }

  return (
    <FormDialog
      open={open}
      onOpenChange={handleOpenChange}
      title={sku ? t("skuForm.editTitle") : t("skuForm.createTitle")}
      formId="sku-form"
      onSubmit={handleSubmit(onSubmit)}
      error={saveError}
      isSubmitting={isSubmitting}
      submitLabel={
        sku ? t("skuForm.submit.save") : t("skuForm.submit.create")
      }
      submittingLabel={t("skuForm.submit.saving")}
      className="sm:max-w-lg"
    >
      <FieldGroup>
        <SKUSpecFields
          fields={fields}
          errors={errors.specs}
          register={register}
          specsEditable={specsEditable}
          onAdd={() => append({ key: "", value: "" })}
          onRemove={remove}
        />

        <div className="grid gap-4 sm:grid-cols-2">
          <Field data-invalid={errors.price ? true : undefined}>
            <FieldLabel htmlFor="sku-form-price">
              {t("skuForm.fields.price")}
            </FieldLabel>
            <Input
              id="sku-form-price"
              inputMode="decimal"
              autoComplete="off"
              aria-invalid={errors.price ? true : undefined}
              {...register("price")}
            />
            {errors.price ? (
              <FieldError>
                {translateValidationMessage(errors.price.message, t)}
              </FieldError>
            ) : null}
          </Field>

          <Field data-invalid={errors.stock ? true : undefined}>
            <FieldLabel htmlFor="sku-form-stock">
              {t("skuForm.fields.stock")}
            </FieldLabel>
            <Input
              id="sku-form-stock"
              inputMode="numeric"
              autoComplete="off"
              aria-invalid={errors.stock ? true : undefined}
              {...register("stock")}
            />
            {errors.stock ? (
              <FieldError>
                {translateValidationMessage(errors.stock.message, t)}
              </FieldError>
            ) : null}
          </Field>
        </div>

        <Field orientation="horizontal">
          <FieldLabel htmlFor="sku-form-enabled">
            {t("skuForm.fields.enabled")}
          </FieldLabel>
          <Controller
            control={control}
            name="is_enabled"
            render={({ field }) => (
              <Switch
                id="sku-form-enabled"
                aria-label={t("skuForm.fields.enabled")}
                checked={field.value}
                onCheckedChange={field.onChange}
              />
            )}
          />
        </Field>
      </FieldGroup>
    </FormDialog>
  )
}
