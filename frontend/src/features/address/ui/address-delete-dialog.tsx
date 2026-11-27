import { useTranslation } from "react-i18next"
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/shared/ui/alert-dialog"
import { Alert, AlertDescription } from "@/shared/ui/alert"
import { Button } from "@/shared/ui/button"
import { describeAddressError } from "../model/errors"
import type { Address } from "../types"

export function AddressDeleteDialog({
  address,
  error,
  isPending,
  onClose,
  onConfirm,
}: {
  address: Address | null
  error: unknown
  isPending: boolean
  onClose: () => void
  onConfirm: () => void
}) {
  const { t } = useTranslation("address")

  return (
    <AlertDialog
      open={address !== null}
      onOpenChange={(open) => {
        if (!open && !isPending) {
          onClose()
        }
      }}
    >
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{t("deleteDialog.title")}</AlertDialogTitle>
          <AlertDialogDescription>
            {t("deleteDialog.description", {
              name: address?.receiver_name ?? "",
            })}
          </AlertDialogDescription>
        </AlertDialogHeader>

        {error ? (
          <Alert variant="destructive">
            <AlertDescription>
              {describeAddressError(error, t)}
            </AlertDescription>
          </Alert>
        ) : null}

        <AlertDialogFooter>
          <AlertDialogCancel disabled={isPending}>
            {t("deleteDialog.cancel")}
          </AlertDialogCancel>
          <Button
            variant="destructive"
            disabled={isPending}
            onClick={onConfirm}
          >
            {t("deleteDialog.confirm")}
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
