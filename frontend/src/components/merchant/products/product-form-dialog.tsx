import { useEffect, useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { AlertCircle, LoaderCircle } from "lucide-react"
import { Controller, useForm } from "react-hook-form"
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
import type { Category, Product, ProductInput } from "@/lib/merchant-api"
import { createProduct, updateProduct } from "@/lib/merchant-api"
import { describeMerchantError } from "@/lib/merchant-errors"
import {
  merchantProductSchema,
  translateMerchantValidationMessage,
  type MerchantProductValues,
} from "@/lib/merchant-schemas"

type ProductSavedKind = "created" | "updated"

type ProductFormDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  token: string
  categories: Category[]
  product: Product | null
  onSaved: (kind: ProductSavedKind) => void
}

type ProductFormProps = {
  token: string
  categories: Category[]
  product: Product | null
  onSaved: (kind: ProductSavedKind) => void
  onPendingChange: (pending: boolean) => void
  onCancel: () => void
}

function toDefaultValues(product: Product | null): MerchantProductValues {
  if (!product) {
    return {
      name: "",
      description: "",
      categoryId: "",
      status: "off_shelf",
    }
  }
  return {
    name: product.name,
    description: product.description,
    categoryId:
      product.category_id === null ? "" : String(product.category_id),
    status: product.status,
  }
}

export function ProductFormDialog({
  open,
  onOpenChange,
  token,
  categories,
  product,
  onSaved,
}: ProductFormDialogProps) {
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
      <DialogContent className="sm:max-w-md" closeLabel={t("common.close")}>
        <ProductForm
          key={product?.id ?? "create"}
          token={token}
          categories={categories}
          product={product}
          onSaved={onSaved}
          onPendingChange={setPending}
          onCancel={() => onOpenChange(false)}
        />
      </DialogContent>
    </Dialog>
  )
}

function ProductForm({
  token,
  categories,
  product,
  onSaved,
  onPendingChange,
  onCancel,
}: ProductFormProps) {
  const { t } = useTranslation("merchant")
  const [formError, setFormError] = useState<unknown>(null)
  const {
    register,
    handleSubmit,
    control,
    formState: { errors, isSubmitting },
  } = useForm<MerchantProductValues>({
    resolver: zodResolver(merchantProductSchema),
    defaultValues: toDefaultValues(product),
  })

  useEffect(() => {
    onPendingChange(isSubmitting)
  }, [isSubmitting, onPendingChange])

  function clearFormError() {
    setFormError(null)
  }

  async function onSubmit(values: MerchantProductValues) {
    setFormError(null)
    const input: ProductInput = {
      name: values.name,
      description: values.description,
      category_id: values.categoryId === "" ? null : Number(values.categoryId),
      status: values.status,
    }
    try {
      if (product) {
        await updateProduct(token, product.id, input)
        onSaved("updated")
      } else {
        await createProduct(token, input)
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
          {product
            ? t("products.form.titleEdit")
            : t("products.form.titleCreate")}
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

        <FieldGroup>
          <Field data-invalid={errors.name ? true : undefined}>
            <FieldLabel htmlFor="product-name">
              {t("products.form.nameLabel")}
            </FieldLabel>
            <Input
              id="product-name"
              placeholder={t("products.form.namePlaceholder")}
              required
              aria-invalid={errors.name ? true : undefined}
              aria-describedby={errors.name ? "product-name-error" : undefined}
              className="h-9"
              {...register("name", { onChange: clearFormError })}
            />
            {errors.name ? (
              <FieldError id="product-name-error">
                {translateMerchantValidationMessage(errors.name.message, t)}
              </FieldError>
            ) : null}
          </Field>

          <Field data-invalid={errors.description ? true : undefined}>
            <FieldLabel htmlFor="product-description">
              {t("products.form.descriptionLabel")}
            </FieldLabel>
            <Textarea
              id="product-description"
              rows={3}
              placeholder={t("products.form.descriptionPlaceholder")}
              aria-invalid={errors.description ? true : undefined}
              aria-describedby={
                errors.description ? "product-description-error" : undefined
              }
              {...register("description", { onChange: clearFormError })}
            />
            {errors.description ? (
              <FieldError id="product-description-error">
                {translateMerchantValidationMessage(
                  errors.description.message,
                  t,
                )}
              </FieldError>
            ) : null}
          </Field>

          <Field>
            <FieldLabel>{t("products.form.categoryLabel")}</FieldLabel>
            <Controller
              control={control}
              name="categoryId"
              render={({ field }) => (
                <Select
                  value={field.value}
                  onValueChange={(value) => {
                    field.onChange(value ?? "")
                    clearFormError()
                  }}
                >
                  <SelectTrigger
                    className="w-full"
                    aria-label={t("products.form.categoryLabel")}
                  >
                    <SelectValue>
                      {field.value === ""
                        ? t("products.form.uncategorized")
                        : (categories.find(
                            (item) => String(item.id) === field.value,
                          )?.name ??
                          product?.category_name ??
                          t("products.form.uncategorized"))}
                    </SelectValue>
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="">
                      {t("products.form.uncategorized")}
                    </SelectItem>
                    {categories.map((category) => (
                      <SelectItem key={category.id} value={String(category.id)}>
                        {category.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            />
          </Field>

          <Field>
            <FieldLabel>{t("products.form.statusLabel")}</FieldLabel>
            <Controller
              control={control}
              name="status"
              render={({ field }) => (
                <Select
                  value={field.value}
                  onValueChange={(value) => {
                    field.onChange(value ?? "off_shelf")
                    clearFormError()
                  }}
                >
                  <SelectTrigger
                    className="w-full"
                    aria-label={t("products.form.statusLabel")}
                  >
                    <SelectValue>
                      {field.value === "on_shelf"
                        ? t("status.onShelf")
                        : t("status.offShelf")}
                    </SelectValue>
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="on_shelf">
                      {t("status.onShelf")}
                    </SelectItem>
                    <SelectItem value="off_shelf">
                      {t("status.offShelf")}
                    </SelectItem>
                  </SelectContent>
                </Select>
              )}
            />
          </Field>
        </FieldGroup>

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
