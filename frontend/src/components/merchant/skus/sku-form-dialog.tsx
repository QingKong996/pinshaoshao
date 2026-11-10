import { useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { AlertCircle, LoaderCircle, Plus, X } from "lucide-react"
import { useFieldArray, useForm, useWatch } from "react-hook-form"
import { useTranslation } from "react-i18next"
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
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Switch } from "@/components/ui/switch"
import { centsToYuan, yuanToCents } from "@/lib/format"
import { createProductSKU, updateProductSKU, type SKU } from "@/lib/merchant-api"
import { describeMerchantError } from "@/lib/merchant-errors"
import {
  merchantSKUSchema,
  SPEC_MAX_COUNT,
  translateMerchantValidationMessage,
  type MerchantSKUValues,
} from "@/lib/merchant-schemas"
import { useMerchantSession } from "@/lib/use-merchant-session"

type SKUFormDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  productID: number
  sku: SKU | null
  templateKeys: string[]
  onSaved: () => void
}

function buildDefaultValues(
  sku: SKU | null,
  templateKeys: string[],
): MerchantSKUValues {
  if (sku) {
    return {
      specs: Object.entries(sku.specs).map(([key, value]) => ({ key, value })),
      price: centsToYuan(sku.price_cents),
      stock: String(sku.stock),
      isEnabled: sku.is_enabled,
    }
  }

  if (templateKeys.length > 0) {
    return {
      specs: templateKeys.map((key) => ({ key, value: "" })),
      price: "",
      stock: "",
      isEnabled: true,
    }
  }

  return {
    specs: [{ key: "", value: "" }],
    price: "",
    stock: "",
    isEnabled: true,
  }
}

export function SKUFormDialog({
  open,
  onOpenChange,
  productID,
  sku,
  templateKeys,
  onSaved,
}: SKUFormDialogProps) {
  const { t } = useTranslation("merchant")
  const { token } = useMerchantSession()
  const fixedSpecs = sku !== null || templateKeys.length > 0
  const [pending, setPending] = useState(false)
  const [saveError, setSaveError] = useState<unknown>(null)

  const {
    control,
    register,
    handleSubmit,
    setValue,
    formState: { errors },
  } = useForm<MerchantSKUValues>({
    resolver: zodResolver(merchantSKUSchema),
    defaultValues: buildDefaultValues(sku, templateKeys),
  })

  const { fields, append, remove } = useFieldArray({
    control,
    name: "specs",
  })
  const isEnabled = useWatch({ control, name: "isEnabled" })
  const specsError = translateMerchantValidationMessage(
    errors.specs?.message,
    t,
  )

  async function onSubmit(values: MerchantSKUValues) {
    setPending(true)
    setSaveError(null)

    const specs = Object.fromEntries(
      values.specs.map((spec) => [spec.key, spec.value]),
    )
    const body = {
      specs,
      price_cents: yuanToCents(values.price),
      stock: Number(values.stock),
      is_enabled: values.isEnabled,
    }

    try {
      if (sku) {
        await updateProductSKU(token, productID, sku.id, body)
      } else {
        await createProductSKU(token, productID, body)
      }
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
      <DialogContent className="sm:max-w-lg" closeLabel={t("common.close")}>
        <DialogHeader>
          <DialogTitle>
            {sku ? t("skus.form.titleEdit") : t("skus.form.titleCreate")}
          </DialogTitle>
          <DialogDescription>
            {fixedSpecs ? t("skus.form.specsHint") : t("skus.form.firstSKUsHint")}
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

          <FieldGroup>
            <Field data-invalid={specsError ? true : undefined}>
              <FieldLabel>{t("skus.form.specsLabel")}</FieldLabel>
              <div className="flex flex-col gap-2">
                {fields.map((field, index) => {
                  const rowErrors = errors.specs?.[index]
                  const keyErrorId = `sku-spec-key-error-${index}`
                  const valueErrorId = `sku-spec-value-error-${index}`
                  const keyLabel = `${t("skus.form.specKeyPlaceholder")} ${index + 1}`
                  const valueLabel = `${t("skus.form.specValuePlaceholder")} ${index + 1}`
                  return (
                    <div key={field.id} className="flex flex-col gap-1">
                      <div className="flex items-start gap-2">
                        {fixedSpecs ? (
                          <Input
                            value={field.key}
                            disabled
                            readOnly
                            aria-label={keyLabel}
                            className="flex-1"
                          />
                        ) : (
                          <Input
                            placeholder={t("skus.form.specKeyPlaceholder")}
                            className="flex-1"
                            aria-label={keyLabel}
                            aria-invalid={
                              rowErrors?.key ? true : undefined
                            }
                            aria-describedby={
                              rowErrors?.key ? keyErrorId : undefined
                            }
                            {...register(`specs.${index}.key`)}
                          />
                        )}
                        <Input
                          placeholder={t("skus.form.specValuePlaceholder")}
                          className="flex-1"
                          aria-label={valueLabel}
                          aria-invalid={rowErrors?.value ? true : undefined}
                          aria-describedby={
                            rowErrors?.value ? valueErrorId : undefined
                          }
                          {...register(`specs.${index}.value`)}
                        />
                        {fixedSpecs ? null : (
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon-sm"
                            aria-label={t("skus.form.removeSpec")}
                            onClick={() => remove(index)}
                            disabled={fields.length <= 1}
                          >
                            <X aria-hidden />
                          </Button>
                        )}
                      </div>
                      {rowErrors?.key ? (
                        <FieldError id={keyErrorId}>
                          {translateMerchantValidationMessage(
                            rowErrors.key.message,
                            t,
                          )}
                        </FieldError>
                      ) : null}
                      {rowErrors?.value ? (
                        <FieldError id={valueErrorId}>
                          {translateMerchantValidationMessage(
                            rowErrors.value.message,
                            t,
                          )}
                        </FieldError>
                      ) : null}
                    </div>
                  )
                })}
              </div>
              {fixedSpecs ? null : (
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  className="w-fit gap-1.5"
                  onClick={() => append({ key: "", value: "" })}
                  disabled={fields.length >= SPEC_MAX_COUNT}
                >
                  <Plus aria-hidden />
                  {t("skus.form.addSpec")}
                </Button>
              )}
              {specsError ? <FieldError>{specsError}</FieldError> : null}
            </Field>

            <div className="grid grid-cols-2 gap-4">
              <Field data-invalid={errors.price ? true : undefined}>
                <FieldLabel htmlFor="sku-price">
                  {t("skus.form.priceLabel")}
                </FieldLabel>
                <Input
                  id="sku-price"
                  inputMode="decimal"
                  placeholder={t("skus.form.pricePlaceholder")}
                  aria-invalid={errors.price ? true : undefined}
                  aria-describedby={
                    errors.price ? "sku-price-error" : undefined
                  }
                  {...register("price")}
                />
                {errors.price ? (
                  <FieldError id="sku-price-error">
                    {translateMerchantValidationMessage(
                      errors.price.message,
                      t,
                    )}
                  </FieldError>
                ) : null}
              </Field>

              <Field data-invalid={errors.stock ? true : undefined}>
                <FieldLabel htmlFor="sku-stock">
                  {t("skus.form.stockLabel")}
                </FieldLabel>
                <Input
                  id="sku-stock"
                  inputMode="numeric"
                  placeholder={t("skus.form.stockPlaceholder")}
                  aria-invalid={errors.stock ? true : undefined}
                  aria-describedby={
                    errors.stock ? "sku-stock-error" : undefined
                  }
                  {...register("stock")}
                />
                {errors.stock ? (
                  <FieldError id="sku-stock-error">
                    {translateMerchantValidationMessage(
                      errors.stock.message,
                      t,
                    )}
                  </FieldError>
                ) : null}
              </Field>
            </div>

            <Field>
              <FieldLabel htmlFor="sku-enabled">
                {t("skus.form.enabledLabel")}
              </FieldLabel>
              <div className="flex items-center gap-2">
                <Switch
                  id="sku-enabled"
                  checked={isEnabled}
                  aria-describedby="sku-enabled-hint"
                  onCheckedChange={(checked) =>
                    setValue("isEnabled", checked)
                  }
                />
                <FieldDescription id="sku-enabled-hint">
                  {t("skus.form.enabledHint")}
                </FieldDescription>
              </div>
            </Field>
          </FieldGroup>

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
              {sku ? t("common.save") : t("common.add")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
