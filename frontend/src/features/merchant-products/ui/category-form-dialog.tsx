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
import type { CatalogChanged, ProductCategory } from "../types"
import {
  useCreateProductCategory,
  useUpdateProductCategory,
} from "../model/mutations"
import { describeConsoleError } from "../model/errors"
import {
  categoryFormSchema,
  categoryToInput,
  categoryToValues,
  INITIAL_VALUES,
  translateValidationMessage,
  type FormValues,
} from "../model/category-form"

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
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-sm">
        <DialogHeader>
          <DialogTitle>
            {category
              ? t("categoryForm.editTitle")
              : t("categoryForm.createTitle")}
          </DialogTitle>
        </DialogHeader>

        <form
          id="category-form"
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
        </form>

        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            onClick={() => handleOpenChange(false)}
          >
            {t("common.cancel")}
          </Button>
          <Button type="submit" form="category-form" disabled={isSubmitting}>
            {isSubmitting ? (
              <>
                <LoaderCircle className="animate-spin" aria-hidden />
                {t("categoryForm.submit.saving")}
              </>
            ) : category ? (
              t("categoryForm.submit.save")
            ) : (
              t("categoryForm.submit.create")
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
