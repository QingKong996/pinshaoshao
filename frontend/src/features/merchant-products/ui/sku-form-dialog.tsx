import { useMemo } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { AlertCircle, LoaderCircle, Plus, Trash2 } from "lucide-react"
import { Controller, useFieldArray, useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { Alert, AlertDescription, AlertTitle } from "@/shared/ui/alert"
import { Button } from "@/shared/ui/button"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/shared/ui/dialog"
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
import { describeConsoleError } from "../model/errors"
import {
  MAX_SPECS,
  buildFormSchema,
  translateValidationMessage,
  fixedSpecKeys,
  initialSpecs,
  skuToInput,
  type FormValues,
} from "../model/sku-form"

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
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>
            {sku ? t("skuForm.editTitle") : t("skuForm.createTitle")}
          </DialogTitle>
        </DialogHeader>

        <form
          id="sku-form"
          noValidate
          onSubmit={handleSubmit(onSubmit)}
          className="flex flex-col gap-4"
        >
          {saveError ? (
            <Alert variant="destructive">
              <AlertCircle aria-hidden />
              <AlertTitle>{t("common.saveError")}</AlertTitle>
              <AlertDescription>
                {describeConsoleError(saveError, t)}
              </AlertDescription>
            </Alert>
          ) : null}

          <FieldGroup>
            <Field>
              <div className="flex flex-wrap items-center justify-between gap-2">
                <FieldLabel>{t("skuForm.fields.specs")}</FieldLabel>
                {specsEditable ? null : (
                  <span className="text-xs text-muted-foreground">
                    {t("skuForm.specsLockedHint")}
                  </span>
                )}
              </div>

              {specsEditable ? (
                <div className="flex flex-col gap-2">
                  {fields.map((field, index) => {
                    const keyError = errors.specs?.[index]?.key?.message
                    const valueError = errors.specs?.[index]?.value?.message

                    return (
                      <div key={field.id} className="flex items-start gap-2">
                        <Field
                          className="flex-1"
                          data-invalid={keyError ? true : undefined}
                        >
                          <Input
                            autoComplete="off"
                            placeholder={t("skuForm.specPlaceholder.key")}
                            aria-label={t("skuForm.fields.specKey")}
                            aria-invalid={keyError ? true : undefined}
                            {...register(`specs.${index}.key`)}
                          />
                          {keyError ? (
                            <FieldError>
                              {translateValidationMessage(keyError, t)}
                            </FieldError>
                          ) : null}
                        </Field>
                        <Field
                          className="flex-1"
                          data-invalid={valueError ? true : undefined}
                        >
                          <Input
                            autoComplete="off"
                            placeholder={t("skuForm.specPlaceholder.value")}
                            aria-label={t("skuForm.fields.specValue")}
                            aria-invalid={valueError ? true : undefined}
                            {...register(`specs.${index}.value`)}
                          />
                          {valueError ? (
                            <FieldError>
                              {translateValidationMessage(valueError, t)}
                            </FieldError>
                          ) : null}
                        </Field>
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon-sm"
                          aria-label={t("skuForm.removeSpec")}
                          disabled={fields.length <= 1}
                          onClick={() => remove(index)}
                        >
                          <Trash2 aria-hidden />
                        </Button>
                      </div>
                    )
                  })}

                  {fields.length < MAX_SPECS ? (
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      className="w-fit"
                      onClick={() => append({ key: "", value: "" })}
                    >
                      <Plus aria-hidden />
                      {t("skuForm.addSpec")}
                    </Button>
                  ) : null}
                </div>
              ) : (
                <div className="flex flex-col gap-3">
                  {fields.map((field, index) => {
                    const valueError = errors.specs?.[index]?.value?.message
                    const inputID = `sku-form-spec-${index}`

                    return (
                      <Field
                        key={field.id}
                        data-invalid={valueError ? true : undefined}
                      >
                        <FieldLabel htmlFor={inputID}>{field.key}</FieldLabel>
                        <Input
                          id={inputID}
                          autoComplete="off"
                          aria-invalid={valueError ? true : undefined}
                          {...register(`specs.${index}.value`)}
                        />
                        {valueError ? (
                          <FieldError>
                            {translateValidationMessage(valueError, t)}
                          </FieldError>
                        ) : null}
                      </Field>
                    )
                  })}
                </div>
              )}
            </Field>

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
        </form>

        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            onClick={() => handleOpenChange(false)}
          >
            {t("common.cancel")}
          </Button>
          <Button type="submit" form="sku-form" disabled={isSubmitting}>
            {isSubmitting ? (
              <>
                <LoaderCircle className="animate-spin" aria-hidden />
                {t("skuForm.submit.saving")}
              </>
            ) : sku ? (
              t("skuForm.submit.save")
            ) : (
              t("skuForm.submit.create")
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
