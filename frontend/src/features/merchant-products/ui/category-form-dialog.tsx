import { useEffect } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/shared/ui/field"
import { Input } from "@/shared/ui/input"
import type { CatalogChanged, ProductCategory } from "../types"
import {
  useCreateProductCategory,
  useUpdateProductCategory,
} from "../model/mutations"
import {
  categoryFormSchema,
  categoryToInput,
  categoryToValues,
  INITIAL_VALUES,
  translateValidationMessage,
  type FormValues,
} from "../model/category-form"
import { FormDialog } from "./form-dialog"

export function CategoryFormDialog({
  token,
  onCatalogChanged,
  open,
  onOpenChange,
  category,
}: {
  token: string
  onCatalogChanged: CatalogChanged
  open: boolean
  onOpenChange: (open: boolean) => void
  category?: ProductCategory
}) {
  const { t } = useTranslation("console")
  const createCategoryMutation = useCreateProductCategory(
    token,
    onCatalogChanged,
  )
  const updateCategoryMutation = useUpdateProductCategory(
    token,
    onCatalogChanged,
  )
  const saveError = category
    ? updateCategoryMutation.error
    : createCategoryMutation.error

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<FormValues>({
    resolver: zodResolver(categoryFormSchema),
    defaultValues: INITIAL_VALUES,
  })

  useEffect(() => {
    if (open) {
      reset(category ? categoryToValues(category) : INITIAL_VALUES)
    }
  }, [category, open, reset])

  function handleOpenChange(nextOpen: boolean) {
    if (!nextOpen) {
      createCategoryMutation.reset()
      updateCategoryMutation.reset()
    }
    onOpenChange(nextOpen)
  }

  async function onSubmit(values: FormValues) {
    const input = categoryToInput(values)

    try {
      if (category) {
        await updateCategoryMutation.mutateAsync({
          categoryID: category.id,
          input,
        })
      } else {
        await createCategoryMutation.mutateAsync(input)
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
      title={
        category
          ? t("categoryForm.editTitle")
          : t("categoryForm.createTitle")
      }
      formId="category-form"
      onSubmit={handleSubmit(onSubmit)}
      error={saveError}
      isSubmitting={isSubmitting}
      submitLabel={
        category
          ? t("categoryForm.submit.save")
          : t("categoryForm.submit.create")
      }
      submittingLabel={t("categoryForm.submit.saving")}
      className="sm:max-w-sm"
    >
      <FieldGroup>
        <Field data-invalid={errors.name ? true : undefined}>
          <FieldLabel htmlFor="category-form-name">
            {t("categoryForm.fields.name")}
          </FieldLabel>
          <Input
            id="category-form-name"
            autoComplete="off"
            autoFocus
            placeholder={t("categoryForm.fields.namePlaceholder")}
            aria-invalid={errors.name ? true : undefined}
            {...register("name")}
          />
          {errors.name ? (
            <FieldError>
              {translateValidationMessage(errors.name.message, t)}
            </FieldError>
          ) : null}
        </Field>
      </FieldGroup>
    </FormDialog>
  )
}
