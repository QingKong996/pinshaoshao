import { useEffect } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import type { ParseKeys, TFunction } from "i18next"
import { AlertCircle, LoaderCircle } from "lucide-react"
import { Controller, useForm } from "react-hook-form"
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Textarea } from "@/components/ui/textarea"
import {
  type Product,
  type ProductCategory,
  type ProductInput,
} from "@/api/merchant-console"
import {
  useCreateProduct,
  useUpdateProduct,
} from "@/merchant-console/mutations"
import { describeConsoleError } from "@/merchant-console/errors"
import { useMerchantSession } from "@/hooks/use-merchant-session"

const PRODUCT_NAME_MAX = 100
const UNCATEGORIZED = "none"

const VALIDATION_MESSAGES = {
  nameRequired: "productForm.errors.nameRequired",
  nameLength: "productForm.errors.nameLength",
} as const

const productFormSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, VALIDATION_MESSAGES.nameRequired)
    .max(PRODUCT_NAME_MAX, VALIDATION_MESSAGES.nameLength),
  description: z.string(),
  categoryId: z.string(),
  status: z.enum(["on_shelf", "off_shelf"]),
})

type FormValues = z.infer<typeof productFormSchema>

const INITIAL_VALUES: FormValues = {
  name: "",
  description: "",
  categoryId: UNCATEGORIZED,
  status: "off_shelf",
}

function productToValues(product: Product): FormValues {
  return {
    name: product.name,
    description: product.description,
    categoryId:
      product.category_id === null ? UNCATEGORIZED : String(product.category_id),
    status: product.status,
  }
}

function translateValidationMessage(
  message: string | undefined,
  t: TFunction<"console">,
) {
  if (!message) {
    return undefined
  }
  return message === VALIDATION_MESSAGES.nameLength
    ? t(message as ParseKeys<"console">, { max: PRODUCT_NAME_MAX })
    : t(message as ParseKeys<"console">)
}

export type ProductFormDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  product?: Product
  categories: ProductCategory[]
}

export function ProductFormDialog({
  open,
  onOpenChange,
  product,
  categories,
}: ProductFormDialogProps) {
  const { t } = useTranslation("console")
  const { token } = useMerchantSession()
  const createProductMutation = useCreateProduct(token)
  const updateProductMutation = useUpdateProduct(token)
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
    const input: ProductInput = {
      name: values.name,
      description: values.description,
      category_id:
        values.categoryId === UNCATEGORIZED ? null : Number(values.categoryId),
      status: values.status,
    }

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
