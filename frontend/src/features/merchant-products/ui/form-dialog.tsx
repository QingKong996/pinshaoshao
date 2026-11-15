import type { FormEventHandler, ReactNode } from "react"
import { AlertCircle, LoaderCircle } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Alert, AlertDescription, AlertTitle } from "@/shared/ui/alert"
import { Button } from "@/shared/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/shared/ui/dialog"
import { describeConsoleError } from "../model/errors"

type FormDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  description?: ReactNode
  formId: string
  onSubmit: FormEventHandler<HTMLFormElement>
  error?: unknown
  isSubmitting: boolean
  submitLabel: string
  submittingLabel: string
  className?: string
  children: ReactNode
}

export function FormDialog({
  open,
  onOpenChange,
  title,
  description,
  formId,
  onSubmit,
  error,
  isSubmitting,
  submitLabel,
  submittingLabel,
  className,
  children,
}: FormDialogProps) {
  const { t } = useTranslation("console")

  function handleOpenChange(nextOpen: boolean) {
    if (nextOpen || !isSubmitting) {
      onOpenChange(nextOpen)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent
        className={className}
        closeLabel={t("common.close")}
        showCloseButton={!isSubmitting}
      >
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          {description ? (
            <DialogDescription>{description}</DialogDescription>
          ) : null}
        </DialogHeader>

        <form
          id={formId}
          noValidate
          onSubmit={onSubmit}
          className="flex flex-col gap-4"
        >
          {error ? (
            <Alert variant="destructive">
              <AlertCircle aria-hidden />
              <AlertTitle>{t("common.saveError")}</AlertTitle>
              <AlertDescription>
                {describeConsoleError(error, t)}
              </AlertDescription>
            </Alert>
          ) : null}
          {children}
        </form>

        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            disabled={isSubmitting}
            onClick={() => handleOpenChange(false)}
          >
            {t("common.cancel")}
          </Button>
          <Button type="submit" form={formId} disabled={isSubmitting}>
            {isSubmitting ? (
              <>
                <LoaderCircle className="animate-spin" aria-hidden />
                {submittingLabel}
              </>
            ) : (
              submitLabel
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
