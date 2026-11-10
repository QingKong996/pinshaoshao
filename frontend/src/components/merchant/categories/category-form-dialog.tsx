import { useEffect, useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { AlertCircle, LoaderCircle } from "lucide-react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Field, FieldError, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import type { Category } from "@/lib/merchant-api"
import { createCategory, updateCategory } from "@/lib/merchant-api"
import { describeMerchantError } from "@/lib/merchant-errors"
import {
  merchantCategorySchema,
  translateMerchantValidationMessage,
  type MerchantCategoryValues,
} from "@/lib/merchant-schemas"

type CategorySavedKind = "created" | "updated"

type CategoryFormDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  token: string
  category: Category | null
  onSaved: (kind: CategorySavedKind) => void
}

type CategoryFormProps = {
  token: string
  category: Category | null
  onSaved: (kind: CategorySavedKind) => void
  onPendingChange: (pending: boolean) => void
  onCancel: () => void
}

export function CategoryFormDialog({
  open,
  onOpenChange,
  token,
  category,
  onSaved,
}: CategoryFormDialogProps) {
  const { t } = useTranslation("merchant")
  const [pending, setPending] = useState(false)

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
        <CategoryForm
          key={category?.id ?? "create"}
          token={token}
          category={category}
          onSaved={onSaved}
          onPendingChange={setPending}
          onCancel={() => onOpenChange(false)}
        />
      </DialogContent>
    </Dialog>
  )
}

function CategoryForm({
  token,
  category,
  onSaved,
  onPendingChange,
  onCancel,
}: CategoryFormProps) {
  const { t } = useTranslation("merchant")
  const [formError, setFormError] = useState<unknown>(null)
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<MerchantCategoryValues>({
    resolver: zodResolver(merchantCategorySchema),
    defaultValues: { name: category?.name ?? "" },
  })

  useEffect(() => {
    onPendingChange(isSubmitting)
  }, [isSubmitting, onPendingChange])

  async function onSubmit(values: MerchantCategoryValues) {
    setFormError(null)
    try {
      if (category) {
        await updateCategory(token, category.id, values)
        onSaved("updated")
      } else {
        await createCategory(token, values)
        onSaved("created")
      }
    } catch (error) {
      setFormError(error)
    }
  }

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          {category
            ? t("categories.form.titleEdit")
            : t("categories.form.titleCreate")}
        </DialogTitle>
      </DialogHeader>
      <form
        noValidate
        aria-busy={isSubmitting}
        onSubmit={handleSubmit(onSubmit)}
        className="flex flex-col gap-4"
      >
        {formError ? (
          <Alert variant="destructive">
            <AlertCircle aria-hidden />
            <AlertDescription>
              {describeMerchantError(formError, t)}
            </AlertDescription>
          </Alert>
        ) : null}

        <Field data-invalid={errors.name ? true : undefined}>
          <FieldLabel htmlFor="category-name">
            {t("categories.form.nameLabel")}
          </FieldLabel>
          <Input
            id="category-name"
            placeholder={t("categories.form.namePlaceholder")}
            required
            aria-invalid={errors.name ? true : undefined}
            aria-describedby={errors.name ? "category-name-error" : undefined}
            className="h-9"
            {...register("name", { onChange: () => setFormError(null) })}
          />
          {errors.name ? (
            <FieldError id="category-name-error">
              {translateMerchantValidationMessage(errors.name.message, t)}
            </FieldError>
          ) : null}
        </Field>

        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            disabled={isSubmitting}
            onClick={onCancel}
          >
            {t("common.cancel")}
          </Button>
          <Button type="submit" disabled={isSubmitting} className="gap-2">
            {isSubmitting ? (
              <LoaderCircle className="animate-spin" aria-hidden />
            ) : null}
            {isSubmitting ? t("common.saving") : t("common.save")}
          </Button>
        </DialogFooter>
      </form>
    </>
  )
}
