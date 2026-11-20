import { useMemo, useState } from "react"
import { ShoppingBag, Trash2 } from "lucide-react"
import type { TFunction } from "i18next"
import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { Badge } from "@/shared/ui/badge"
import { Button, buttonVariants } from "@/shared/ui/button"
import { Card } from "@/shared/ui/card"
import { Input } from "@/shared/ui/input"
import { PageHeader } from "@/shared/ui/page-header"
import { QueryError } from "@/shared/ui/query-error"
import { Skeleton } from "@/shared/ui/skeleton"
import { formatPriceCents } from "@/shared/lib/format"
import { useDocumentTitle } from "@/shared/hooks/use-document-title"
import type { CartItem } from "../types"
import { formatCartSpecs, summarizeSelection } from "../model/cart"
import { describeCartError } from "../model/errors"
import {
  useRemoveCartItem,
  useUpdateCartItemQuantity,
} from "../model/mutations"
import { useCart } from "../model/queries"
import { CartRemoveDialog } from "./cart-remove-dialog"

const EMPTY_CART: CartItem[] = []

export function CartPage({ token }: { token: string }) {
  const { t } = useTranslation("cart")
  useDocumentTitle(t("documentTitle"))

  const cartQuery = useCart(token)
  const updateQuantity = useUpdateCartItemQuantity(token)
  const removeItem = useRemoveCartItem(token)
  const [selection, setSelection] = useState<Set<number> | null>(null)
  const [removeTarget, setRemoveTarget] = useState<CartItem | null>(null)
  const [actionError, setActionError] = useState<unknown>(null)

  const items = cartQuery.data ?? EMPTY_CART
  const availableIDs = useMemo(
    () =>
      items
        .filter((item) => item.availability === "available")
        .map((item) => item.id),
    [items],
  )
  const selectedIDs = useMemo(
    () => selection ?? new Set(availableIDs),
    [availableIDs, selection],
  )
  const summary = useMemo(
    () => summarizeSelection(items, selectedIDs),
    [items, selectedIDs],
  )

  function toggleItem(item: CartItem) {
    const next = new Set(selectedIDs)
    if (next.has(item.id)) {
      next.delete(item.id)
    } else {
      next.add(item.id)
    }
    setSelection(next)
  }

  async function changeQuantity(item: CartItem, quantity: number) {
    if (quantity === item.quantity || quantity < 1) {
      return
    }
    setActionError(null)
    try {
      await updateQuantity.mutateAsync({ itemID: item.id, quantity })
    } catch (error) {
      setActionError(error)
    }
  }

  async function confirmRemove() {
    if (!removeTarget) {
      return
    }
    setActionError(null)
    try {
      await removeItem.mutateAsync(removeTarget.id)
      setSelection((current) => {
        if (!current) {
          return current
        }
        const next = new Set(current)
        next.delete(removeTarget.id)
        return next
      })
      setRemoveTarget(null)
    } catch (error) {
      setActionError(error)
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title={t("title")} description={t("description")} />

      {actionError ? (
        <QueryError
          title={t("errors.title")}
          description={describeCartError(actionError, t)}
          retryLabel={t("errors.dismiss")}
          onRetry={() => setActionError(null)}
        />
      ) : null}

      {cartQuery.error && !cartQuery.data ? (
        <QueryError
          title={t("errors.loadTitle")}
          description={describeCartError(cartQuery.error, t)}
          retryLabel={t("errors.retry")}
          onRetry={cartQuery.refetch}
        />
      ) : cartQuery.isPending ? (
        <CartSkeleton />
      ) : items.length === 0 ? (
        <Card className="items-center gap-3 px-6 py-14 text-center">
          <ShoppingBag
            className="size-10 text-muted-foreground"
            aria-hidden
          />
          <div className="space-y-1">
            <h2 className="font-heading text-lg font-medium">
              {t("empty.title")}
            </h2>
            <p className="text-muted-foreground">
              {t("empty.description")}
            </p>
          </div>
        </Card>
      ) : (
        <div className="grid items-start gap-6 lg:grid-cols-[minmax(0,1fr)_20rem]">
          <Card className="gap-0 overflow-hidden p-0">
            <ul className="divide-y">
              {items.map((item) => (
                <CartItemRow
                  key={item.id}
                  item={item}
                  selected={selectedIDs.has(item.id)}
                  isPending={
                    updateQuantity.isPending || removeItem.isPending
                  }
                  onToggle={() => toggleItem(item)}
                  onQuantityChange={(quantity) =>
                    void changeQuantity(item, quantity)
                  }
                  onRemove={() => setRemoveTarget(item)}
                />
              ))}
            </ul>
          </Card>

          <Card className="gap-4 lg:sticky lg:top-24">
            <h2 className="font-heading text-base font-medium">
              {t("summary.title")}
            </h2>
            <dl className="space-y-3">
              <div className="flex items-center justify-between gap-4">
                <dt className="text-muted-foreground">
                  {t("summary.selected")}
                </dt>
                <dd>{t("summary.itemCount", { count: summary.count })}</dd>
              </div>
              <div className="flex items-center justify-between gap-4 border-t pt-3">
                <dt className="font-medium">{t("summary.subtotal")}</dt>
                <dd className="font-heading text-lg font-semibold tabular-nums">
                  {formatPriceCents(summary.subtotalCents)}
                </dd>
              </div>
            </dl>
            <p className="text-xs leading-relaxed text-muted-foreground">
              {t("summary.notice")}
            </p>
          </Card>
        </div>
      )}

      <CartRemoveDialog
        item={removeTarget}
        isPending={removeItem.isPending}
        onClose={() => setRemoveTarget(null)}
        onConfirm={() => void confirmRemove()}
      />
    </div>
  )
}

export function CartSignedOut() {
  const { t } = useTranslation("cart")
  useDocumentTitle(t("documentTitle"))

  return (
    <Card className="items-center gap-3 px-6 py-16 text-center">
      <ShoppingBag className="size-10 text-muted-foreground" aria-hidden />
      <div className="space-y-1">
        <h1 className="font-heading text-2xl font-semibold">
          {t("signedOut.title")}
        </h1>
        <p className="max-w-md text-muted-foreground">
          {t("signedOut.description")}
        </p>
      </div>
      <Link to="/signin" className={buttonVariants()}>
        {t("signedOut.action")}
      </Link>
    </Card>
  )
}

function CartItemRow({
  item,
  selected,
  isPending,
  onToggle,
  onQuantityChange,
  onRemove,
}: {
  item: CartItem
  selected: boolean
  isPending: boolean
  onToggle: () => void
  onQuantityChange: (quantity: number) => void
  onRemove: () => void
}) {
  const { t } = useTranslation("cart")
  const available = item.availability === "available"

  return (
    <li className="flex gap-3 p-4 sm:gap-4 sm:p-5">
      <input
        type="checkbox"
        checked={available && selected}
        disabled={!available}
        aria-label={t("item.select", { name: item.product_name })}
        onChange={onToggle}
        className="mt-1 size-4 shrink-0 accent-primary disabled:cursor-not-allowed"
      />

      <div className="flex min-w-0 flex-1 flex-col gap-4 sm:flex-row sm:justify-between">
        <div className="min-w-0 space-y-1.5">
          <div className="flex flex-wrap items-center gap-2">
            <h2 className="font-medium">{item.product_name}</h2>
            {available ? null : (
              <Badge variant="destructive">
                {availabilityLabel(item, t)}
              </Badge>
            )}
          </div>
          <p className="text-sm text-muted-foreground">
            {formatCartSpecs(item.sku_specs, t("item.defaultSpecs"))}
          </p>
          <p className="text-sm">
            {t("item.snapshotPrice", {
              price: formatPriceCents(item.unit_price_cents),
            })}
          </p>
          <p className="text-xs text-muted-foreground">
            {item.available_stock === null
              ? t("item.stockUnknown")
              : t("item.availableStock", {
                  count: item.available_stock,
                })}
          </p>
        </div>

        <div className="flex shrink-0 flex-wrap items-end gap-4 sm:flex-col sm:items-end">
          <QuantityControl
            key={item.quantity}
            quantity={item.quantity}
            disabled={isPending}
            onChange={onQuantityChange}
          />
          <div className="text-right">
            <p className="text-xs text-muted-foreground">
              {t("item.lineTotal")}
            </p>
            <p className="font-medium tabular-nums">
              {formatPriceCents(item.line_total_cents)}
            </p>
          </div>
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label={t("item.remove", { name: item.product_name })}
            disabled={isPending}
            onClick={onRemove}
          >
            <Trash2 aria-hidden />
          </Button>
        </div>
      </div>
    </li>
  )
}

function QuantityControl({
  quantity,
  disabled,
  onChange,
}: {
  quantity: number
  disabled: boolean
  onChange: (quantity: number) => void
}) {
  const { t } = useTranslation("cart")
  const [value, setValue] = useState(String(quantity))

  function commit() {
    const next = Number(value)
    if (!Number.isInteger(next) || next < 1) {
      setValue(String(quantity))
      return
    }
    if (next !== quantity) {
      onChange(next)
    }
  }

  return (
    <div className="flex items-center gap-1">
      <Button
        variant="outline"
        size="icon-sm"
        aria-label={t("item.decrease")}
        disabled={disabled || quantity <= 1}
        onClick={() => onChange(quantity - 1)}
      >
        −
      </Button>
      <Input
        value={value}
        inputMode="numeric"
        aria-label={t("item.quantity")}
        disabled={disabled}
        className="h-8 w-12 px-1 text-center tabular-nums"
        onChange={(event) => setValue(event.target.value)}
        onBlur={commit}
        onKeyDown={(event) => {
          if (event.key === "Enter") {
            event.currentTarget.blur()
          }
          if (event.key === "Escape") {
            setValue(String(quantity))
            event.currentTarget.blur()
          }
        }}
      />
      <Button
        variant="outline"
        size="icon-sm"
        aria-label={t("item.increase")}
        disabled={disabled}
        onClick={() => onChange(quantity + 1)}
      >
        +
      </Button>
    </div>
  )
}

function availabilityLabel(
  item: CartItem,
  t: TFunction<"cart">,
) {
  switch (item.availability) {
    case "product_deleted":
      return t("availability.productDeleted")
    case "product_off_shelf":
      return t("availability.productOffShelf")
    case "sku_deleted":
      return t("availability.skuDeleted")
    case "sku_disabled":
      return t("availability.skuDisabled")
    case "out_of_stock":
      return t("availability.outOfStock")
    case "insufficient_stock":
      return t("availability.insufficientStock", {
        count: item.available_stock ?? 0,
      })
    default:
      return t("availability.available")
  }
}

function CartSkeleton() {
  return (
    <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_20rem]">
      <Card className="gap-0 divide-y overflow-hidden p-0">
        {Array.from({ length: 3 }, (_, index) => (
          <div key={index} className="flex gap-4 p-5">
            <Skeleton className="size-4" />
            <div className="flex-1 space-y-2">
              <Skeleton className="h-4 w-40" />
              <Skeleton className="h-3 w-56" />
              <Skeleton className="h-3 w-24" />
            </div>
          </div>
        ))}
      </Card>
      <Skeleton className="h-52 rounded-xl" />
    </div>
  )
}
