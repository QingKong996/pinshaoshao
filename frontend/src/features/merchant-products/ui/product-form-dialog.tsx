import { useEffect } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { AlertCircle, LoaderCircle } from "lucide-react"
import { Controller, useForm } from "react-hook-form"
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/shared/ui/select"
import { Textarea } from "@/shared/ui/textarea"
import {
  type Product,
  type ProductCategory,
  type CatalogChanged,
} from "../types"
import {
  useCreateProduct,
  useUpdateProduct,
} from "../model/mutations"
import { describeConsoleError } from "../model/errors"
import {
  productFormSchema,
  INITIAL_VALUES,
  UNCATEGORIZED,
  productToValues,
  productToInput,
  translateValidationMessage,
  type FormValues,
} from "../model/product-form"

export type ProductFormDialogProps = {
  token: string
  onCatalogChanged: CatalogChanged
  open: boolean
  onOpenChange: (open: boolean) => void
  product?: Product
  categories: ProductCategory[]
}

export function ProductFormDialog({
  token,
  onCatalogChanged,
  open,
  onOpenChange,
  product,
  categories,
}: ProductFormDialogProps) {
  const { t } = useTranslation("console")
  const createProductMutation = useCreateProduct(token, onCatalogChanged)
  const updateProductMutation = useUpdateProduct(token, onCatalogChanged)
  const saveError = product
    ? updateProductMutation.error
    : createProductMutation.error

  const {
    register,
    control,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<FormValues>({
    resolver: zodResolver(productFormSchema),
    defaultValues: INITIAL_VALUES,
  })

  useEffect(() => {
    if (open) {
      reset(product ? productToValues(product) : INITIAL_VALUES)
    }
  }, [open, product, reset])

  function handleOpenChange(nextOpen: boolean) {
    if (!nextOpen) {
      createProductMutation.reset()
      updateProductMutation.reset()
    }
    onOpenChange(nextOpen)
  }

  const categoryItems = [
    { value: UNCATEGORIZED, label: t("products.filters.uncategorized") },
    ...categories.map((category) => ({
      value: String(category.id),
      label: category.name,
    })),
  ]
  const statusItems = [
    { value: "on_shelf", label: t("productStatus.on_shelf") },
    { value: "off_shelf", label: t("productStatus.off_shelf") },
  ]

  async function onSubmit(values: FormValues) {
    const input = productToInput(values)

    try {
      if (product) {
        await updateProductMutation.mutateAsync({
          productID: product.id,
          input,
        })
      } else {
        await createProductMutation.mutateAsync(input)
      }
      handleOpenChange(false)
    } catch {
      // the mutation state renders the error
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>
            {product
              ? t("productForm.editTitle")
              : t("productForm.createTitle")}
          </DialogTitle>
        </DialogHeader>

        <form
          id="product-form"
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
              <FieldLabel htmlFor="product-form-name">
                {t("productForm.fields.name")}
              </FieldLabel>
              <Input
                id="product-form-name"
                autoComplete="off"
                placeholder={t("productForm.fields.namePlaceholder")}
                aria-invalid={errors.name ? true : undefined}
                {...register("name")}
              />
              {errors.name ? (
                <FieldError>
                  {translateValidationMessage(errors.name.message, t)}
                </FieldError>
              ) : null}
            </Field>

            <div className="grid gap-4 sm:grid-cols-2">
              <Field>
                <FieldLabel htmlFor="product-form-category">
                  {t("productForm.fields.category")}
                </FieldLabel>
                <Controller
                  control={control}
                  name="categoryId"
                  render={({ field }) => (
                    <Select
                      value={field.value}
                      onValueChange={field.onChange}
                      items={categoryItems}
                    >
                      <SelectTrigger
                        id="product-form-category"
                        className="w-full"
                      >
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {categoryItems.map((item) => (
                          <SelectItem key={item.value} value={item.value}>
                            {item.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  )}
                />
              </Field>

              <Field>
                <FieldLabel htmlFor="product-form-status">
                  {t("productForm.fields.status")}
                </FieldLabel>
                <Controller
                  control={control}
                  name="status"
                  render={({ field }) => (
                    <Select
                      value={field.value}
                      onValueChange={field.onChange}
                      items={statusItems}
                    >
                      <SelectTrigger
                        id="product-form-status"
                        className="w-full"
                      >
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {statusItems.map((item) => (
                          <SelectItem key={item.value} value={item.value}>
                            {item.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  )}
                />
              </Field>
            </div>

            <Field>
              <FieldLabel htmlFor="product-form-description">
                {t("productForm.fields.description")}
              </FieldLabel>
              <Textarea
                id="product-form-description"
                rows={3}
                placeholder={t("productForm.fields.descriptionPlaceholder")}
                {...register("description")}
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
          <Button type="submit" form="product-form" disabled={isSubmitting}>
            {isSubmitting ? (
              <>
                <LoaderCircle className="animate-spin" aria-hidden />
                {t("productForm.submit.saving")}
              </>
            ) : product ? (
              t("productForm.submit.save")
            ) : (
              t("productForm.submit.create")
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
