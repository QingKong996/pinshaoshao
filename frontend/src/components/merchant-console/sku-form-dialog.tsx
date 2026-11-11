import { useMemo } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import type { ParseKeys, TFunction } from "i18next"
import { AlertCircle, LoaderCircle, Plus, Trash2 } from "lucide-react"
import { Controller, useFieldArray, useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { z } from "zod"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Switch } from "@/components/ui/switch"
import {
  type ProductSKU,
  type ProductSKUInput,
} from "@/api/merchant-console"
import {
  useCreateProductSKU,
  useUpdateProductSKU,
} from "@/merchant-console/mutations"
import { describeConsoleError } from "@/merchant-console/errors"
import { useMerchantSession } from "@/hooks/use-merchant-session"

const MAX_SPECS = 5
const PRICE_PATTERN = /^\d+(\.\d{1,2})?$/
const STOCK_PATTERN = /^\d+$/

const VALIDATION_MESSAGES = {
  specKeyRequired: "skuForm.errors.specKeyRequired",
  specValueRequired: "skuForm.errors.specValueRequired",
  specDuplicate: "skuForm.errors.specDuplicate",
  specLimit: "skuForm.errors.specLimit",
  priceRequired: "skuForm.errors.priceRequired",
  priceInvalid: "skuForm.errors.priceInvalid",
  stockRequired: "skuForm.errors.stockRequired",
  stockInvalid: "skuForm.errors.stockInvalid",
} as const

const skuFormSchema = z.object({
  specs: z.array(z.object({ key: z.string(), value: z.string() })),
  price: z.string(),
  stock: z.string(),
  is_enabled: z.boolean(),
})

type FormValues = z.infer<typeof skuFormSchema>

function buildFormSchema(specsEditable: boolean) {
  return skuFormSchema.superRefine((values, ctx) => {
    const seenKeys = new Set<string>()

    values.specs.forEach((spec, index) => {
      const key = spec.key.trim()
      if (specsEditable) {
        if (!key) {
          ctx.addIssue({
            code: "custom",
            message: VALIDATION_MESSAGES.specKeyRequired,
            path: ["specs", index, "key"],
          })
        } else if (seenKeys.has(key)) {
          ctx.addIssue({
            code: "custom",
            message: VALIDATION_MESSAGES.specDuplicate,
            path: ["specs", index, "key"],
          })
        }
        if (key) {
          seenKeys.add(key)
        }
      }
      if (!spec.value.trim()) {
        ctx.addIssue({
          code: "custom",
          message: VALIDATION_MESSAGES.specValueRequired,
          path: ["specs", index, "value"],
        })
      }
    })

    if (specsEditable && values.specs.length > MAX_SPECS) {
      ctx.addIssue({
        code: "custom",
        message: VALIDATION_MESSAGES.specLimit,
        path: ["specs", MAX_SPECS],
      })
    }

    const price = values.price.trim()
    if (!price) {
      ctx.addIssue({
        code: "custom",
        message: VALIDATION_MESSAGES.priceRequired,
        path: ["price"],
      })
    } else if (!PRICE_PATTERN.test(price)) {
      ctx.addIssue({
        code: "custom",
        message: VALIDATION_MESSAGES.priceInvalid,
        path: ["price"],
      })
    }

    const stock = values.stock.trim()
    if (!stock) {
      ctx.addIssue({
        code: "custom",
        message: VALIDATION_MESSAGES.stockRequired,
        path: ["stock"],
      })
    } else if (!STOCK_PATTERN.test(stock)) {
      ctx.addIssue({
        code: "custom",
        message: VALIDATION_MESSAGES.stockInvalid,
        path: ["stock"],
      })
    }
  })
}

function translateValidationMessage(
  message: string | undefined,
  t: TFunction<"console">,
) {
  if (!message) {
    return undefined
  }
  return message === VALIDATION_MESSAGES.specLimit
    ? t(message as ParseKeys<"console">, { max: MAX_SPECS })
    : t(message as ParseKeys<"console">)
}

function sortSpecEntries(specs: Record<string, string>) {
  return Object.entries(specs).sort(([left], [right]) =>
    left.localeCompare(right),
  )
}

function fixedSpecKeys(sku: ProductSKU | undefined, skus: ProductSKU[]) {
  const source = sku && skus.length > 1 ? [sku] : skus
  const keys = new Set<string>()

  for (const item of source) {
    for (const key of Object.keys(item.specs)) {
      keys.add(key)
    }
  }

  return [...keys].sort((left, right) => left.localeCompare(right))
}

function initialSpecs(
  sku: ProductSKU | undefined,
  specsEditable: boolean,
  keys: string[],
) {
  if (specsEditable) {
    // An existing SKU may legitimately have no specs (the product's default
    // variant), so only a new SKU starts with an empty pair.
    return sku
      ? sortSpecEntries(sku.specs).map(([key, value]) => ({ key, value }))
      : [{ key: "", value: "" }]
  }
  return keys.map((key) => ({ key, value: sku?.specs[key] ?? "" }))
}

export type SKUFormDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  productID: number
  sku?: ProductSKU
  skus: ProductSKU[]
}

export function SKUFormDialog({
  open,
  onOpenChange,
  productID,
  sku,
  skus,
}: SKUFormDialogProps) {
  const { t } = useTranslation("console")
  const { token } = useMerchantSession()
  const createProductSKUMutation = useCreateProductSKU(token)
  const updateProductSKUMutation = useUpdateProductSKU(token)
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
    const input: ProductSKUInput = {
      specs: Object.fromEntries(
        values.specs.map((spec) => [spec.key.trim(), spec.value.trim()]),
      ),
      price_cents: Math.round(Number(values.price.trim()) * 100),
      stock: Number(values.stock.trim()),
      is_enabled: values.is_enabled,
    }

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
