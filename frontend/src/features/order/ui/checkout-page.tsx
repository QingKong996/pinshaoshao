import { useRef, useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { LoaderCircle, ShoppingBag } from "lucide-react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { Link, useNavigate } from "react-router"
import { Alert, AlertDescription, AlertTitle } from "@/shared/ui/alert"
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/shared/ui/select"
import { formatPriceCents } from "@/shared/lib/format"
import { useDocumentTitle } from "@/shared/hooks/use-document-title"
import { describeCheckoutError, isCheckoutConflict } from "../model/errors"
import {
  directCheckoutSignature,
  readDirectCheckoutAttempt,
  saveDirectCheckoutAttempt,
} from "../model/direct-checkout-attempt"
import {
  INITIAL_CHECKOUT_VALUES,
  checkoutFormSchema,
  translateCheckoutValidationMessage,
  type CheckoutFormValues,
} from "../model/checkout-form"
import { useCreateDirectOrder, useCreateOrder } from "../model/mutations"
import type { CheckoutLine, CheckoutSavedAddress } from "../types"
import { formatOrderSpecs } from "./format-specs"

type CheckoutAttempt = {
  signature: string
  key: string
}

export type CheckoutMode =
  | { type: "cart" }
  | {
      type: "direct"
      skuID: number
      quantity: number
      attemptStorageKey: string
      canStartNew: boolean
    }

const CART_MODE: CheckoutMode = { type: "cart" }

export function CheckoutPage({
  token,
  items,
  mode = CART_MODE,
  savedAddresses = [],
  backHref,
  cartChangeNeedsConfirmation = false,
  onConfirmCartChanges,
  onCartConflict,
  onOrderCreated,
}: {
  token: string
  items: CheckoutLine[]
  mode?: CheckoutMode
  savedAddresses?: CheckoutSavedAddress[]
  backHref?: string
  cartChangeNeedsConfirmation?: boolean
  onConfirmCartChanges?: () => void
  onCartConflict?: () => Promise<void>
  onOrderCreated?: () => void
}) {
  const { t } = useTranslation("order")
  useDocumentTitle(t("checkout.documentTitle"))

  const navigate = useNavigate()
  const createOrder = useCreateOrder(token, onOrderCreated)
  const createDirectOrder = useCreateDirectOrder(token, onOrderCreated)
  const directMode = mode.type === "direct" ? mode : null
  const activeOrder = directMode ? createDirectOrder : createOrder
  const resolvedBackHref = backHref ?? (directMode ? "/" : "/cart")
  const cartAttempt = useRef<CheckoutAttempt | null>(null)
  const [directAttempt, setDirectAttempt] = useState(() =>
    directMode ? readDirectCheckoutAttempt(directMode.attemptStorageKey) : null,
  )
  const [directAttemptMismatch, setDirectAttemptMismatch] = useState(false)
  const submittingDirectOrder = useRef(false)
  const initialAddress =
    savedAddresses.find((candidate) => candidate.is_default) ??
    savedAddresses[0]
  const [selectedAddressID, setSelectedAddressID] = useState<number | null>(
    () => initialAddress?.id ?? null,
  )
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<CheckoutFormValues>({
    resolver: zodResolver(checkoutFormSchema),
    defaultValues: initialAddress
      ? savedAddressFormValues(initialAddress)
      : INITIAL_CHECKOUT_VALUES,
  })

  function selectSavedAddress(value: string | null) {
    if (value === null) {
      return
    }
    const address = savedAddresses.find(
      (candidate) => candidate.id === Number(value),
    )
    if (!address) {
      return
    }
    setSelectedAddressID(address.id)
    reset(savedAddressFormValues(address))
  }

  const totalCents = items.reduce(
    (total, item) => total + item.line_total_cents,
    0,
  )

  async function onSubmit(values: CheckoutFormValues) {
    if (directMode) {
      await submitDirectOrder(values)
      return
    }
    if (cartChangeNeedsConfirmation || items.length === 0) {
      return
    }
    const payload = {
      items: items.map((item) => ({
        cart_item_id: item.id,
        quantity: item.quantity,
      })),
      ...values,
    }
    const signature = JSON.stringify(payload)
    const checkoutKey =
      cartAttempt.current?.signature === signature
        ? cartAttempt.current.key
        : crypto.randomUUID()
    cartAttempt.current = { signature, key: checkoutKey }

    try {
      const order = await createOrder.mutateAsync({
        checkout_key: checkoutKey,
        ...payload,
      })
      navigate(`/orders/${order.id}`, { replace: true })
    } catch (error) {
      if (isCheckoutConflict(error)) {
        await onCartConflict?.()
      }
    }
  }

  async function submitDirectOrder(
    values: CheckoutFormValues,
    startNew = false,
  ) {
    if (!directMode || submittingDirectOrder.current) {
      return
    }
    const { attemptStorageKey, canStartNew, quantity, skuID } = directMode
    submittingDirectOrder.current = true
    try {
      const payload = { sku_id: skuID, quantity, ...values }
      const signature = await directCheckoutSignature(payload)
      const pendingAttempt =
        readDirectCheckoutAttempt(attemptStorageKey) ?? directAttempt
      if (!canStartNew && (startNew || !pendingAttempt)) {
        return
      }
      if (!startNew && pendingAttempt && pendingAttempt.signature !== signature) {
        setDirectAttempt(pendingAttempt)
        setDirectAttemptMismatch(true)
        return
      }

      const checkoutKey =
        (!startNew && pendingAttempt?.key) || crypto.randomUUID()
      const nextAttempt = { signature, key: checkoutKey }
      setDirectAttemptMismatch(false)
      saveDirectCheckoutAttempt(attemptStorageKey, nextAttempt)
      const order = await createDirectOrder
        .mutateAsync({
          checkout_key: checkoutKey,
          ...payload,
        })
        .catch(() => null)
      if (!order) {
        setDirectAttempt(nextAttempt)
        return
      }
      if (readDirectCheckoutAttempt(attemptStorageKey)?.key === checkoutKey) {
        saveDirectCheckoutAttempt(attemptStorageKey, null)
      }
      navigate(`/orders/${order.id}`, { replace: true })
    } finally {
      submittingDirectOrder.current = false
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("checkout.title")}
        description={t("checkout.description")}
        actions={
          <Link
            to={resolvedBackHref}
            className={buttonVariants({ variant: "outline", size: "sm" })}
          >
            {directMode
              ? t("checkout.backToProduct")
              : t("checkout.backToCart")}
          </Link>
        }
      />

      {activeOrder.error ? (
        <QueryError
          title={t("errors.checkoutTitle")}
          description={describeCheckoutError(activeOrder.error, t)}
          retryLabel={t("errors.dismiss")}
          onRetry={() => activeOrder.reset()}
        />
      ) : null}

      {cartChangeNeedsConfirmation ? (
        <Alert>
          <AlertTitle>{t("checkout.cartChanged.title")}</AlertTitle>
          <AlertDescription className="flex flex-col items-start gap-3">
            <span>{t("checkout.cartChanged.description")}</span>
            {items.length > 0 ? (
              <Button type="button" onClick={onConfirmCartChanges}>
                {t("checkout.cartChanged.confirm")}
              </Button>
            ) : null}
          </AlertDescription>
        </Alert>
      ) : null}

      {directMode && directAttempt ? (
        <Alert>
          <AlertTitle>{t("checkout.directAttempt.title")}</AlertTitle>
          <AlertDescription className="flex flex-col items-start gap-3">
            <span>
              {t(
                directAttemptMismatch
                  ? "checkout.directAttempt.mismatch"
                  : "checkout.directAttempt.description",
              )}
            </span>
            <Link to="/orders" className="underline">
              {t("checkout.directAttempt.viewOrders")}
            </Link>
          </AlertDescription>
        </Alert>
      ) : null}

      {items.length === 0 && !directAttempt ? (
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
              {directMode
                ? t("checkout.direct.invalidDescription")
                : t("checkout.empty.description")}
            </p>
          </div>
          <Link to={resolvedBackHref} className={buttonVariants()}>
            {directMode
              ? t("checkout.backToProduct")
              : t("checkout.empty.action")}
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
                onSubmit={(event) => handleSubmit(onSubmit)(event)}
                className="flex flex-col gap-5"
              >
                <FieldGroup>
                  {savedAddresses.length > 0 ? (
                    <Field>
                      <FieldLabel htmlFor="checkout-saved-address">
                        {t("checkout.form.savedAddress.label")}
                      </FieldLabel>
                      <Select
                        items={savedAddresses.map((address) => ({
                          value: String(address.id),
                          label: `${address.receiver_name} · ${address.shipping_address}`,
                        }))}
                        value={
                          selectedAddressID === null
                            ? null
                            : String(selectedAddressID)
                        }
                        onValueChange={selectSavedAddress}
                      >
                        <SelectTrigger
                          id="checkout-saved-address"
                          className="w-full"
                          aria-label={t("checkout.form.savedAddress.label")}
                        >
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {savedAddresses.map((address) => (
                            <SelectItem
                              key={address.id}
                              value={String(address.id)}
                            >
                              {address.receiver_name} ·{" "}
                              {address.shipping_address}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </Field>
                  ) : null}

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

                <div className="flex flex-wrap justify-end gap-2 border-t pt-4">
                  {directMode?.canStartNew && directAttempt ? (
                    <Button
                      type="button"
                      variant="outline"
                      onClick={(event) =>
                        handleSubmit((values) => submitDirectOrder(values, true))(
                          event,
                        )
                      }
                      disabled={isSubmitting || activeOrder.isPending}
                    >
                      {t("checkout.directAttempt.startNew")}
                    </Button>
                  ) : null}
                  <Button
                    type="submit"
                    disabled={
                      cartChangeNeedsConfirmation ||
                      isSubmitting ||
                      activeOrder.isPending
                    }
                  >
                    {isSubmitting || activeOrder.isPending ? (
                      <LoaderCircle className="animate-spin" aria-hidden />
                    ) : null}
                    {directAttempt
                      ? t("checkout.directAttempt.retry")
                      : t("checkout.form.submit")}
                  </Button>
                </div>
              </form>
            </CardContent>
          </Card>

          {items.length > 0 ? (
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
                  {directMode
                    ? t("checkout.summary.directNotice")
                    : t("checkout.summary.notice")}
                </p>
              </CardContent>
            </Card>
          ) : null}
        </div>
      )}
    </div>
  )
}

function savedAddressFormValues(
  address: CheckoutSavedAddress,
): CheckoutFormValues {
  return {
    receiver_name: address.receiver_name,
    receiver_phone: address.receiver_phone,
    shipping_address: address.shipping_address,
  }
}
