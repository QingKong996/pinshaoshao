import { useEffect } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { LoaderCircle } from "lucide-react"
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
import { Textarea } from "@/shared/ui/textarea"
import { describeAddressError } from "../model/errors"
import {
  addressToFormValues,
  addressToInput,
  addressFormSchema,
  INITIAL_ADDRESS_VALUES,
  translateAddressValidationMessage,
  type AddressFormValues,
} from "../model/form"
import { useCreateAddress, useUpdateAddress } from "../model/mutations"
import type { Address } from "../types"

export function AddressFormDialog({
  token,
  address,
  open,
  onOpenChange,
}: {
  token: string
  address: Address | null
  open: boolean
  onOpenChange: (open: boolean) => void
}) {
  const { t } = useTranslation("address")
  const createAddress = useCreateAddress(token)
  const updateAddress = useUpdateAddress(token)
  const saveError = address ? updateAddress.error : createAddress.error
  const isMutating = createAddress.isPending || updateAddress.isPending
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<AddressFormValues>({
    resolver: zodResolver(addressFormSchema),
    defaultValues: INITIAL_ADDRESS_VALUES,
  })

  useEffect(() => {
    if (open) {
      reset(address ? addressToFormValues(address) : INITIAL_ADDRESS_VALUES)
    }
  }, [address, open, reset])

  function handleOpenChange(nextOpen: boolean) {
    if (!nextOpen) {
      createAddress.reset()
      updateAddress.reset()
    }
    onOpenChange(nextOpen)
  }

  async function onSubmit(values: AddressFormValues) {
    const input = addressToInput(values)
    try {
      if (address) {
        await updateAddress.mutateAsync({
          addressID: address.id,
          input,
        })
      } else {
        await createAddress.mutateAsync(input)
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
            {address ? t("form.editTitle") : t("form.createTitle")}
          </DialogTitle>
          <DialogDescription>{t("form.description")}</DialogDescription>
        </DialogHeader>

        {saveError ? (
          <Alert variant="destructive">
            <AlertTitle>{t("errors.saveTitle")}</AlertTitle>
            <AlertDescription>
              {describeAddressError(saveError, t)}
            </AlertDescription>
          </Alert>
        ) : null}

        <form
          id="address-form"
          onSubmit={handleSubmit(onSubmit)}
          className="flex flex-col gap-5"
        >
          <FieldGroup>
            <Field data-invalid={errors.receiver_name ? true : undefined}>
              <FieldLabel htmlFor="address-receiver-name">
                {t("form.receiverName")}
              </FieldLabel>
              <Input
                id="address-receiver-name"
                autoComplete="name"
                aria-invalid={errors.receiver_name ? true : undefined}
                {...register("receiver_name")}
              />
              {errors.receiver_name ? (
                <FieldError>
                  {translateAddressValidationMessage(
                    errors.receiver_name.message,
                    t,
                  )}
                </FieldError>
              ) : null}
            </Field>

            <Field data-invalid={errors.receiver_phone ? true : undefined}>
              <FieldLabel htmlFor="address-receiver-phone">
                {t("form.receiverPhone")}
              </FieldLabel>
              <Input
                id="address-receiver-phone"
                autoComplete="tel"
                aria-invalid={errors.receiver_phone ? true : undefined}
                {...register("receiver_phone")}
              />
              {errors.receiver_phone ? (
                <FieldError>
                  {translateAddressValidationMessage(
                    errors.receiver_phone.message,
                    t,
                  )}
                </FieldError>
              ) : null}
            </Field>

            <Field
              data-invalid={errors.shipping_address ? true : undefined}
            >
              <FieldLabel htmlFor="address-shipping-address">
                {t("form.shippingAddress")}
              </FieldLabel>
              <Textarea
                id="address-shipping-address"
                autoComplete="street-address"
                rows={3}
                aria-invalid={errors.shipping_address ? true : undefined}
                {...register("shipping_address")}
              />
              {errors.shipping_address ? (
                <FieldError>
                  {translateAddressValidationMessage(
                    errors.shipping_address.message,
                    t,
                  )}
                </FieldError>
              ) : null}
            </Field>

            <Field orientation="horizontal">
              <input
                id="address-is-default"
                type="checkbox"
                className="size-4 accent-primary"
                {...register("is_default")}
              />
              <div className="flex flex-col gap-1">
                <FieldLabel htmlFor="address-is-default">
                  {t("form.default")}
                </FieldLabel>
                <FieldDescription>
                  {t("form.defaultHint")}
                </FieldDescription>
              </div>
            </Field>
          </FieldGroup>
        </form>

        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            disabled={isMutating}
            onClick={() => handleOpenChange(false)}
          >
            {t("form.cancel")}
          </Button>
          <Button
            type="submit"
            form="address-form"
            disabled={isSubmitting || isMutating}
          >
            {isSubmitting || isMutating ? (
              <LoaderCircle className="animate-spin" aria-hidden />
            ) : null}
            {address ? t("form.save") : t("form.create")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
