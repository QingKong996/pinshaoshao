import { useEffect, useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { AlertCircle, LoaderCircle, Truck } from "lucide-react"
import { useForm } from "react-hook-form"
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
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/shared/ui/field"
import { Input } from "@/shared/ui/input"
import {
  dispatchFormSchema,
  translateDispatchValidation,
  type DispatchFormValues,
} from "../model/dispatch-form"
import { describeShipmentError } from "../model/errors"
import { useCreateShipment } from "../model/mutations"
import type { Shipment } from "../types"

export function ShipmentDispatchDialog({
  token,
  open,
  onOpenChange,
  subOrderID,
  receiverName,
  onCreated,
}: {
  token: string
  open: boolean
  onOpenChange: (open: boolean) => void
  subOrderID?: number
  receiverName?: string
  onCreated?: (shipment: Shipment) => void
}) {
  const { t } = useTranslation("logistics")
  const {
    mutateAsync: createShipment,
    reset: resetMutation,
    error: mutationError,
  } = useCreateShipment(token)
  const [localError, setLocalError] = useState<unknown>(null)
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<DispatchFormValues>({
    resolver: zodResolver(dispatchFormSchema),
    defaultValues: {
      subOrderID: "",
      carrier: "",
    },
  })
  const isFixedSubOrder = subOrderID !== undefined

  useEffect(() => {
    if (!open) {
      return
    }
    resetMutation()
    reset({
      subOrderID: subOrderID ? String(subOrderID) : "",
      carrier: "",
    })
  }, [open, reset, resetMutation, subOrderID])

  function closeDialog() {
    if (isSubmitting) {
      return
    }
    setLocalError(null)
    resetMutation()
    onOpenChange(false)
  }

  async function onSubmit(values: DispatchFormValues) {
    setLocalError(null)
    try {
      const shipment = await createShipment({
        sub_order_id: subOrderID ?? Number(values.subOrderID),
        carrier: values.carrier.trim() || undefined,
      })
      onCreated?.(shipment)
      onOpenChange(false)
    } catch (error) {
      setLocalError(error)
    }
  }

  const error = localError ?? mutationError

  return (
    <Dialog
      open={open}
      onOpenChange={(nextOpen) => {
        if (nextOpen) {
          onOpenChange(true)
        } else {
          closeDialog()
        }
      }}
    >
      <DialogContent
        closeLabel={t("actions.close")}
        showCloseButton={!isSubmitting}
      >
        <DialogHeader>
          <DialogTitle>{t("dispatchForm.title")}</DialogTitle>
          <DialogDescription>
            {isFixedSubOrder
              ? t("dispatchForm.fixedDescription", {
                  id: subOrderID,
                  receiver:
                    receiverName ?? t("dispatchForm.unknownReceiver"),
                })
              : t("dispatchForm.description")}
          </DialogDescription>
        </DialogHeader>

        <form
          id="shipment-dispatch-form"
          noValidate
          className="flex flex-col gap-4"
          onSubmit={handleSubmit(onSubmit)}
        >
          {error ? (
            <Alert variant="destructive">
              <AlertCircle aria-hidden />
              <AlertTitle>{t("errors.dispatchTitle")}</AlertTitle>
              <AlertDescription>
                {describeShipmentError(error, t)}
              </AlertDescription>
            </Alert>
          ) : null}

          <FieldGroup>
            {isFixedSubOrder ? null : (
              <Field data-invalid={errors.subOrderID ? true : undefined}>
                <FieldLabel htmlFor="shipment-sub-order-id">
                  {t("dispatchForm.fields.subOrderID")}
                </FieldLabel>
                <Input
                  id="shipment-sub-order-id"
                  type="number"
                  inputMode="numeric"
                  min={1}
                  step={1}
                  autoFocus
                  aria-invalid={errors.subOrderID ? true : undefined}
                  {...register("subOrderID")}
                />
                <FieldDescription>
                  {t("dispatchForm.subOrderHint")}
                </FieldDescription>
                {errors.subOrderID ? (
                  <FieldError>
                    {translateDispatchValidation(
                      errors.subOrderID.message,
                      t,
                    )}
                  </FieldError>
                ) : null}
              </Field>
            )}

            <Field data-invalid={errors.carrier ? true : undefined}>
              <FieldLabel htmlFor="shipment-carrier">
                {t("dispatchForm.fields.carrier")}
              </FieldLabel>
              <Input
                id="shipment-carrier"
                autoFocus={isFixedSubOrder}
                placeholder={t("dispatchForm.carrierPlaceholder")}
                aria-invalid={errors.carrier ? true : undefined}
                {...register("carrier")}
              />
              <FieldDescription>
                {t("dispatchForm.carrierHint")}
              </FieldDescription>
              {errors.carrier ? (
                <FieldError>
                  {translateDispatchValidation(errors.carrier.message, t)}
                </FieldError>
              ) : null}
            </Field>
          </FieldGroup>
        </form>

        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            disabled={isSubmitting}
            onClick={closeDialog}
          >
            {t("actions.cancel")}
          </Button>
          <Button
            type="submit"
            form="shipment-dispatch-form"
            disabled={isSubmitting}
          >
            {isSubmitting ? (
              <>
                <LoaderCircle className="animate-spin" aria-hidden />
                {t("dispatchForm.submitting")}
              </>
            ) : (
              <>
                <Truck aria-hidden />
                {t("dispatchForm.submit")}
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
