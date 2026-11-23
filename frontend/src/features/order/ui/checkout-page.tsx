import { useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { LoaderCircle, ShoppingBag } from "lucide-react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { Link, useNavigate } from "react-router"
import { Button, buttonVariants } from "@/shared/ui/button"
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/shared/ui/card"
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/shared/ui/field"
import { Input } from "@/shared/ui/input"
import { PageHeader } from "@/shared/ui/page-header"
import { QueryError } from "@/shared/ui/query-error"
import { Textarea } from "@/shared/ui/textarea"
import { formatPriceCents } from "@/shared/lib/format"
import { useDocumentTitle } from "@/shared/hooks/use-document-title"
import { describeCheckoutError } from "../model/errors"
import {
  INITIAL_CHECKOUT_VALUES,
  checkoutFormSchema,
  translateCheckoutValidationMessage,
  type CheckoutFormValues,
} from "../model/checkout-form"
import { useCreateOrder } from "../model/mutations"
import type { CheckoutLine } from "../types"
import { formatOrderSpecs } from "./format-specs"

type CheckoutAttempt = {
  signature: string
  key: string
}

export function CheckoutPage({
  token,
  items,
  onOrderCreated,
}: {
  token: string
  items: CheckoutLine[]
  onOrderCreated?: () => void
}) {
  const { t } = useTranslation("order")
  useDocumentTitle(t("checkout.documentTitle"))

  const navigate = useNavigate()
  const createOrder = useCreateOrder(token, onOrderCreated)
  const [attempt, setAttempt] = useState<CheckoutAttempt>({
    signature: "",
    key: "",
  })
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<CheckoutFormValues>({
    resolver: zodResolver(checkoutFormSchema),
    defaultValues: INITIAL_CHECKOUT_VALUES,
  })

  const totalCents = items.reduce(
    (total, item) => total + item.line_total_cents,
    0,
  )

  async function onSubmit(values: CheckoutFormValues) {
    const payload = {
      items: items.map((item) => ({
        cart_item_id: item.id,
        quantity: item.quantity,
      })),
      ...values,
    }
    const signature = JSON.stringify(payload)
    const checkoutKey =
      attempt.signature === signature ? attempt.key : crypto.randomUUID()
    if (checkoutKey !== attempt.key) {
      setAttempt({ signature, key: checkoutKey })
    }

    try {
      const order = await createOrder.mutateAsync({
        checkout_key: checkoutKey,
        ...payload,
      })
      navigate(`/orders/${order.id}`, { replace: true })
    } catch {
      // the mutation state renders the error
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("checkout.title")}
        description={t("checkout.description")}
        actions={
          <Link
            to="/cart"
            className={buttonVariants({ variant: "outline", size: "sm" })}
          >
            {t("checkout.backToCart")}
          </Link>
        }
      />

      {createOrder.error ? (
        <QueryError
          title={t("errors.checkoutTitle")}
          description={describeCheckoutError(createOrder.error, t)}
          retryLabel={t("errors.dismiss")}
          onRetry={() => createOrder.reset()}
        />
      ) : null}

      {items.length === 0 ? (
        <Card className="items-center gap-3 px-6 py-14 text-center">
          <ShoppingBag
            className="size-10 text-muted-foreground"
            aria-hidden
          />
          <div className="space-y-1">
            <h2 className="font-heading text-lg font-medium">
              {t("checkout.empty.title")}
            </h2>
            <p className="text-muted-foreground">
              {t("checkout.empty.description")}
            </p>
          </div>
          <Link to="/cart" className={buttonVariants()}>
            {t("checkout.empty.action")}
          </Link>
        </Card>
      ) : (
        <div className="grid items-start gap-6 lg:grid-cols-[minmax(0,1fr)_22rem]">
          <Card>
            <CardHeader>
              <CardTitle>{t("checkout.form.title")}</CardTitle>
            </CardHeader>
            <CardContent>
              <form
                id="checkout-form"
                onSubmit={handleSubmit(onSubmit)}
                className="flex flex-col gap-5"
              >
                <FieldGroup>
                  <Field data-invalid={errors.receiver_name ? true : undefined}>
                    <FieldLabel htmlFor="checkout-receiver-name">
                      {t("checkout.form.receiverName")}
                    </FieldLabel>
                    <Input
                      id="checkout-receiver-name"
                      autoComplete="name"
                      aria-invalid={
                        errors.receiver_name ? true : undefined
                      }
                      {...register("receiver_name")}
                    />
                    {errors.receiver_name ? (
                      <FieldError>
                        {translateCheckoutValidationMessage(
                          errors.receiver_name.message,
                          t,
                        )}
                      </FieldError>
                    ) : null}
                  </Field>

                  <Field data-invalid={errors.receiver_phone ? true : undefined}>
                    <FieldLabel htmlFor="checkout-receiver-phone">
                      {t("checkout.form.receiverPhone")}
                    </FieldLabel>
                    <Input
                      id="checkout-receiver-phone"
                      autoComplete="tel"
                      aria-invalid={
                        errors.receiver_phone ? true : undefined
                      }
                      {...register("receiver_phone")}
                    />
                    {errors.receiver_phone ? (
                      <FieldError>
                        {translateCheckoutValidationMessage(
                          errors.receiver_phone.message,
                          t,
                        )}
                      </FieldError>
                    ) : null}
                  </Field>

                  <Field
                    data-invalid={
                      errors.shipping_address ? true : undefined
                    }
                  >
                    <FieldLabel htmlFor="checkout-shipping-address">
                      {t("checkout.form.shippingAddress")}
                    </FieldLabel>
                    <Textarea
                      id="checkout-shipping-address"
                      autoComplete="street-address"
                      rows={3}
                      aria-invalid={
                        errors.shipping_address ? true : undefined
                      }
                      {...register("shipping_address")}
                    />
                    {errors.shipping_address ? (
                      <FieldError>
                        {translateCheckoutValidationMessage(
                          errors.shipping_address.message,
                          t,
                        )}
                      </FieldError>
                    ) : null}
                  </Field>
                </FieldGroup>

                <div className="flex justify-end border-t pt-4">
                  <Button
                    type="submit"
                    disabled={isSubmitting || createOrder.isPending}
                  >
                    {isSubmitting || createOrder.isPending ? (
                      <LoaderCircle className="animate-spin" aria-hidden />
                    ) : null}
                    {t("checkout.form.submit")}
                  </Button>
                </div>
              </form>
            </CardContent>
          </Card>

          <Card className="gap-4 lg:sticky lg:top-24">
            <CardHeader>
              <CardTitle>{t("checkout.summary.title")}</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <ul className="divide-y">
                {items.map((item) => (
                  <li
                    key={item.id}
                    className="flex items-start justify-between gap-4 py-3 first:pt-0 last:pb-0"
                  >
                    <div className="min-w-0">
                      <p className="font-medium">{item.product_name}</p>
                      <p className="text-sm text-muted-foreground">
                        {formatOrderSpecs(
                          item.sku_specs,
                          t("checkout.summary.noSpecs"),
                        )}
                      </p>
                      <p className="text-sm text-muted-foreground">
                        {t("checkout.summary.quantity", {
                          count: item.quantity,
                        })}
                      </p>
                    </div>
                    <p className="shrink-0 font-medium tabular-nums">
                      {formatPriceCents(item.line_total_cents)}
                    </p>
                  </li>
                ))}
              </ul>
              <div className="flex items-center justify-between gap-4 border-t pt-4">
                <span className="font-medium">
                  {t("checkout.summary.total")}
                </span>
                <span className="font-heading text-lg font-semibold tabular-nums">
                  {formatPriceCents(totalCents)}
                </span>
              </div>
              <p className="text-xs leading-relaxed text-muted-foreground">
                {t("checkout.summary.notice")}
              </p>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  )
}
