import { useState } from "react"
import { MapPin, Pencil, Plus, Star, Trash2 } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { Alert, AlertDescription, AlertTitle } from "@/shared/ui/alert"
import { Badge } from "@/shared/ui/badge"
import { Button, buttonVariants } from "@/shared/ui/button"
import { Card } from "@/shared/ui/card"
import { PageHeader } from "@/shared/ui/page-header"
import { QueryError } from "@/shared/ui/query-error"
import { Skeleton } from "@/shared/ui/skeleton"
import { useDocumentTitle } from "@/shared/hooks/use-document-title"
import { describeAddressError } from "../model/errors"
import {
  useDeleteAddress,
  useSetDefaultAddress,
} from "../model/mutations"
import { useAddresses } from "../model/queries"
import type { Address } from "../types"
import { AddressDeleteDialog } from "./address-delete-dialog"
import { AddressFormDialog } from "./address-form-dialog"

const EMPTY_ADDRESSES: Address[] = []

export function AddressBookPage({ token }: { token: string }) {
  const { t } = useTranslation("address")
  useDocumentTitle(t("documentTitle"))

  const addressesQuery = useAddresses(token)
  const setDefaultAddress = useSetDefaultAddress(token)
  const deleteAddress = useDeleteAddress(token)
  const [formOpen, setFormOpen] = useState(false)
  const [editTarget, setEditTarget] = useState<Address | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<Address | null>(null)
  const addresses = addressesQuery.data ?? EMPTY_ADDRESSES

  function openCreate() {
    setEditTarget(null)
    setFormOpen(true)
  }

  function openEdit(address: Address) {
    setEditTarget(address)
    setFormOpen(true)
  }

  function closeForm(open: boolean) {
    setFormOpen(open)
    if (!open) {
      setEditTarget(null)
    }
  }

  async function confirmDelete() {
    if (!deleteTarget) {
      return
    }
    try {
      await deleteAddress.mutateAsync(deleteTarget.id)
      setDeleteTarget(null)
      deleteAddress.reset()
    } catch {
      // the delete dialog renders the error
    }
  }

  function closeDeleteDialog() {
    setDeleteTarget(null)
    deleteAddress.reset()
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("title")}
        description={t("description")}
        actions={
          <Button onClick={openCreate}>
            <Plus aria-hidden />
            {t("new")}
          </Button>
        }
      />

      {setDefaultAddress.error ? (
        <Alert variant="destructive">
          <AlertTitle>{t("errors.defaultTitle")}</AlertTitle>
          <AlertDescription>
            {describeAddressError(setDefaultAddress.error, t)}
          </AlertDescription>
        </Alert>
      ) : null}

      {addressesQuery.error && !addressesQuery.data ? (
        <QueryError
          title={t("errors.loadTitle")}
          description={describeAddressError(addressesQuery.error, t)}
          retryLabel={t("errors.retry")}
          onRetry={addressesQuery.refetch}
        />
      ) : addressesQuery.isPending ? (
        <AddressSkeleton />
      ) : addresses.length === 0 ? (
        <Card className="items-center gap-3 px-6 py-14 text-center">
          <MapPin className="size-10 text-muted-foreground" aria-hidden />
          <div className="space-y-1">
            <h2 className="font-heading text-lg font-medium">
              {t("empty.title")}
            </h2>
            <p className="text-muted-foreground">
              {t("empty.description")}
            </p>
          </div>
          <Button onClick={openCreate}>
            <Plus aria-hidden />
            {t("new")}
          </Button>
        </Card>
      ) : (
        <Card className="gap-0 overflow-hidden p-0">
          <ul className="divide-y">
            {addresses.map((address) => (
              <AddressRow
                key={address.id}
                address={address}
                isSettingDefault={setDefaultAddress.isPending}
                isDeleting={deleteAddress.isPending}
                onSetDefault={() =>
                  setDefaultAddress.mutate(address.id)
                }
                onEdit={() => openEdit(address)}
                onDelete={() => setDeleteTarget(address)}
              />
            ))}
          </ul>
        </Card>
      )}

      <AddressFormDialog
        token={token}
        address={editTarget}
        open={formOpen}
        onOpenChange={closeForm}
      />

      <AddressDeleteDialog
        address={deleteTarget}
        error={deleteAddress.error}
        isPending={deleteAddress.isPending}
        onClose={closeDeleteDialog}
        onConfirm={() => void confirmDelete()}
      />
    </div>
  )
}

export function AddressSignedOut() {
  const { t } = useTranslation("address")
  useDocumentTitle(t("documentTitle"))

  return (
    <Card className="items-center gap-3 px-6 py-16 text-center">
      <MapPin className="size-10 text-muted-foreground" aria-hidden />
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

function AddressRow({
  address,
  isSettingDefault,
  isDeleting,
  onSetDefault,
  onEdit,
  onDelete,
}: {
  address: Address
  isSettingDefault: boolean
  isDeleting: boolean
  onSetDefault: () => void
  onEdit: () => void
  onDelete: () => void
}) {
  const { t } = useTranslation("address")
  const isPending = isSettingDefault || isDeleting

  return (
    <li className="flex flex-col gap-4 p-5 sm:flex-row sm:items-start">
      <div className="min-w-0 flex-1 space-y-2">
        <div className="flex flex-wrap items-center gap-2">
          <h2 className="font-medium">{address.receiver_name}</h2>
          {address.is_default ? (
            <Badge>{t("defaultBadge")}</Badge>
          ) : null}
          <span className="text-sm text-muted-foreground">
            {address.receiver_phone}
          </span>
        </div>
        <p className="whitespace-pre-line text-sm leading-relaxed text-muted-foreground">
          {address.shipping_address}
        </p>
      </div>

      <div className="flex shrink-0 flex-wrap items-center gap-1">
        {address.is_default ? null : (
          <Button
            variant="ghost"
            size="sm"
            disabled={isPending}
            onClick={onSetDefault}
          >
            <Star aria-hidden />
            {t("actions.setDefault")}
          </Button>
        )}
        <Button
          variant="ghost"
          size="icon-sm"
          aria-label={t("actions.edit", { name: address.receiver_name })}
          disabled={isPending}
          onClick={onEdit}
        >
          <Pencil aria-hidden />
        </Button>
        <Button
          variant="ghost"
          size="icon-sm"
          aria-label={t("actions.delete", { name: address.receiver_name })}
          disabled={isPending}
          onClick={onDelete}
        >
          <Trash2 aria-hidden />
        </Button>
      </div>
    </li>
  )
}

function AddressSkeleton() {
  return (
    <Card className="gap-0 divide-y overflow-hidden p-0">
      {Array.from({ length: 3 }, (_, index) => (
        <div key={index} className="space-y-2 p-5">
          <Skeleton className="h-4 w-40" />
          <Skeleton className="h-3 w-64" />
          <Skeleton className="h-3 w-52" />
        </div>
      ))}
    </Card>
  )
}
