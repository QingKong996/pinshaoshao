import { Plus, Trash2 } from "lucide-react"
import { type FieldErrors, type UseFormRegister } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { Button } from "@/shared/ui/button"
import {
  Field,
  FieldError,
  FieldLabel,
  FieldTitle,
} from "@/shared/ui/field"
import { Input } from "@/shared/ui/input"
import {
  MAX_SPECS,
  translateValidationMessage,
  type FormValues,
} from "../model/sku-form"

type SpecField = {
  id: string
  key: string
}

type SKUSpecFieldsProps = {
  fields: SpecField[]
  errors: FieldErrors<FormValues>["specs"]
  register: UseFormRegister<FormValues>
  specsEditable: boolean
  onAdd: () => void
  onRemove: (index: number) => void
}

export function SKUSpecFields({
  fields,
  errors,
  register,
  specsEditable,
  onAdd,
  onRemove,
}: SKUSpecFieldsProps) {
  const { t } = useTranslation("console")

  if (!specsEditable) {
    return (
      <div className="flex flex-col gap-3">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <FieldTitle>{t("skuForm.fields.specs")}</FieldTitle>
          <span className="text-xs text-muted-foreground">
            {t("skuForm.specsLockedHint")}
          </span>
        </div>
        {fields.map((field, index) => {
          const valueError = errors?.[index]?.value?.message
          const inputID = `sku-form-spec-${index}`
          const errorID = `${inputID}-error`

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
                aria-describedby={valueError ? errorID : undefined}
                {...register(`specs.${index}.value`)}
              />
              {valueError ? (
                <FieldError id={errorID}>
                  {translateValidationMessage(valueError, t)}
                </FieldError>
              ) : null}
            </Field>
          )
        })}
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-2">
      <FieldTitle>{t("skuForm.fields.specs")}</FieldTitle>
      {fields.map((field, index) => {
        const keyError = errors?.[index]?.key?.message
        const valueError = errors?.[index]?.value?.message
        const keyID = `sku-form-spec-key-${index}`
        const valueID = `sku-form-spec-value-${index}`

        return (
          <div key={field.id} className="flex items-start gap-2">
            <Field
              className="flex-1"
              data-invalid={keyError ? true : undefined}
            >
              <Input
                id={keyID}
                autoComplete="off"
                placeholder={t("skuForm.specPlaceholder.key")}
                aria-label={t("skuForm.fields.specKey")}
                aria-invalid={keyError ? true : undefined}
                aria-describedby={
                  keyError ? `${keyID}-error` : undefined
                }
                {...register(`specs.${index}.key`)}
              />
              {keyError ? (
                <FieldError id={`${keyID}-error`}>
                  {translateValidationMessage(keyError, t)}
                </FieldError>
              ) : null}
            </Field>
            <Field
              className="flex-1"
              data-invalid={valueError ? true : undefined}
            >
              <Input
                id={valueID}
                autoComplete="off"
                placeholder={t("skuForm.specPlaceholder.value")}
                aria-label={t("skuForm.fields.specValue")}
                aria-invalid={valueError ? true : undefined}
                aria-describedby={
                  valueError ? `${valueID}-error` : undefined
                }
                {...register(`specs.${index}.value`)}
              />
              {valueError ? (
                <FieldError id={`${valueID}-error`}>
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
              onClick={() => onRemove(index)}
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
          onClick={onAdd}
        >
          <Plus aria-hidden />
          {t("skuForm.addSpec")}
        </Button>
      ) : null}
    </div>
  )
}
