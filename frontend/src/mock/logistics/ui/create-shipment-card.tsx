import { zodResolver } from "@hookform/resolvers/zod"
import { LoaderCircle, Sparkles } from "lucide-react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { useNavigate } from "react-router"
import { Button } from "@/shared/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/shared/ui/card"
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/shared/ui/field"
import { Input } from "@/shared/ui/input"
import { QueryError } from "@/shared/ui/query-error"
import { describeCreateShipmentError } from "../model/errors"
import { useCreateMockShipment } from "../model/mutations"
import { mockLogisticsDetailPath } from "../model/paths"
import {
  createMockShipmentInitialValues,
  mockShipmentFormSchema,
  toCreateMockShipmentInput,
  translateMockShipmentValidationMessage,
  type MockShipmentFormValues,
} from "../model/shipment-form"

export function CreateShipmentCard() {
  const { t } = useTranslation("mocklogistics")
  const navigate = useNavigate()
  const createShipment = useCreateMockShipment()
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<MockShipmentFormValues>({
    resolver: zodResolver(mockShipmentFormSchema),
    defaultValues: createMockShipmentInitialValues(),
  })

  async function onSubmit(values: MockShipmentFormValues) {
    try {
      const shipment = await createShipment.mutateAsync(
        toCreateMockShipmentInput(values),
      )
      navigate(mockLogisticsDetailPath(shipment.shipment_no))
    } catch {
      // mutation state renders the error
    }
  }

  const isPending = isSubmitting || createShipment.isPending

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("console.create.title")}</CardTitle>
        <CardDescription>{t("console.create.description")}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-5">
        {createShipment.error ? (
          <QueryError
            title={t("errors.createTitle")}
            description={describeCreateShipmentError(createShipment.error, t)}
            retryLabel={t("actions.dismiss")}
            onRetry={() => createShipment.reset()}
          />
        ) : null}

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
          <FieldGroup>
            <div className="grid gap-5 sm:grid-cols-[minmax(0,1fr)_12rem]">
              <Field
                data-invalid={errors.merchant_order_no ? true : undefined}
              >
                <FieldLabel htmlFor="mocklogistics-merchant-order-no">
                  {t("form.merchantOrderNo")}
                </FieldLabel>
                <Input
                  id="mocklogistics-merchant-order-no"
                  autoComplete="off"
                  aria-invalid={errors.merchant_order_no ? true : undefined}
                  {...register("merchant_order_no")}
                />
                <FieldDescription>
                  {t("form.merchantOrderNoHint")}
                </FieldDescription>
                {errors.merchant_order_no ? (
                  <FieldError>
                    {translateMockShipmentValidationMessage(
                      errors.merchant_order_no.message,
                      t,
                    )}
                  </FieldError>
                ) : null}
              </Field>

              <Field data-invalid={errors.carrier ? true : undefined}>
                <FieldLabel htmlFor="mocklogistics-carrier">
                  {t("form.carrier")}
                  <span className="font-normal text-muted-foreground">
                    {t("form.optional")}
                  </span>
                </FieldLabel>
                <Input
                  id="mocklogistics-carrier"
                  autoComplete="off"
                  aria-invalid={errors.carrier ? true : undefined}
                  {...register("carrier")}
                />
                {errors.carrier ? (
                  <FieldError>
                    {translateMockShipmentValidationMessage(
                      errors.carrier.message,
                      t,
                    )}
                  </FieldError>
                ) : null}
              </Field>
            </div>

            <Field data-invalid={errors.item_summary ? true : undefined}>
              <FieldLabel htmlFor="mocklogistics-item-summary">
                {t("form.itemSummary")}
                <span className="font-normal text-muted-foreground">
                  {t("form.optional")}
                </span>
              </FieldLabel>
              <Input
                id="mocklogistics-item-summary"
                autoComplete="off"
                aria-invalid={errors.item_summary ? true : undefined}
                {...register("item_summary")}
              />
              <FieldDescription>{t("form.itemSummaryHint")}</FieldDescription>
              {errors.item_summary ? (
                <FieldError>
                  {translateMockShipmentValidationMessage(
                    errors.item_summary.message,
                    t,
                  )}
                </FieldError>
              ) : null}
            </Field>

            <div className="grid gap-5 sm:grid-cols-2">
              <Field data-invalid={errors.origin_city ? true : undefined}>
                <FieldLabel htmlFor="mocklogistics-origin-city">
                  {t("form.originCity")}
                  <span className="font-normal text-muted-foreground">
                    {t("form.optional")}
                  </span>
                </FieldLabel>
                <Input
                  id="mocklogistics-origin-city"
                  autoComplete="off"
                  aria-invalid={errors.origin_city ? true : undefined}
                  {...register("origin_city")}
                />
                <FieldDescription>{t("form.cityHint")}</FieldDescription>
                {errors.origin_city ? (
                  <FieldError>
                    {translateMockShipmentValidationMessage(
                      errors.origin_city.message,
                      t,
                    )}
                  </FieldError>
                ) : null}
              </Field>

              <Field data-invalid={errors.destination_city ? true : undefined}>
                <FieldLabel htmlFor="mocklogistics-destination-city">
                  {t("form.destinationCity")}
                  <span className="font-normal text-muted-foreground">
                    {t("form.optional")}
                  </span>
                </FieldLabel>
                <Input
                  id="mocklogistics-destination-city"
                  autoComplete="off"
                  aria-invalid={errors.destination_city ? true : undefined}
                  {...register("destination_city")}
                />
                {errors.destination_city ? (
                  <FieldError>
                    {translateMockShipmentValidationMessage(
                      errors.destination_city.message,
                      t,
                    )}
                  </FieldError>
                ) : null}
              </Field>
            </div>

            <div className="grid gap-5 sm:grid-cols-2">
              <Field data-invalid={errors.receiver_name ? true : undefined}>
                <FieldLabel htmlFor="mocklogistics-receiver-name">
                  {t("form.receiverName")}
                  <span className="font-normal text-muted-foreground">
                    {t("form.optional")}
                  </span>
                </FieldLabel>
                <Input
                  id="mocklogistics-receiver-name"
                  autoComplete="off"
                  aria-invalid={errors.receiver_name ? true : undefined}
                  {...register("receiver_name")}
                />
                {errors.receiver_name ? (
                  <FieldError>
                    {translateMockShipmentValidationMessage(
                      errors.receiver_name.message,
                      t,
                    )}
                  </FieldError>
                ) : null}
              </Field>

              <Field data-invalid={errors.receiver_phone ? true : undefined}>
                <FieldLabel htmlFor="mocklogistics-receiver-phone">
                  {t("form.receiverPhone")}
                  <span className="font-normal text-muted-foreground">
                    {t("form.optional")}
                  </span>
                </FieldLabel>
                <Input
                  id="mocklogistics-receiver-phone"
                  autoComplete="off"
                  aria-invalid={errors.receiver_phone ? true : undefined}
                  {...register("receiver_phone")}
                />
                {errors.receiver_phone ? (
                  <FieldError>
                    {translateMockShipmentValidationMessage(
                      errors.receiver_phone.message,
                      t,
                    )}
                  </FieldError>
                ) : null}
              </Field>
            </div>

            <Field data-invalid={errors.receiver_address ? true : undefined}>
              <FieldLabel htmlFor="mocklogistics-receiver-address">
                {t("form.receiverAddress")}
                <span className="font-normal text-muted-foreground">
                  {t("form.optional")}
                </span>
              </FieldLabel>
              <Input
                id="mocklogistics-receiver-address"
                autoComplete="off"
                aria-invalid={errors.receiver_address ? true : undefined}
                {...register("receiver_address")}
              />
              <FieldDescription>{t("form.receiverHint")}</FieldDescription>
              {errors.receiver_address ? (
                <FieldError>
                  {translateMockShipmentValidationMessage(
                    errors.receiver_address.message,
                    t,
                  )}
                </FieldError>
              ) : null}
            </Field>

            <Field data-invalid={errors.notify_url ? true : undefined}>
              <FieldLabel htmlFor="mocklogistics-notify-url">
                {t("form.notifyURL")}
              </FieldLabel>
              <Input
                id="mocklogistics-notify-url"
                type="url"
                autoComplete="off"
                className="font-mono text-xs"
                aria-invalid={errors.notify_url ? true : undefined}
                {...register("notify_url")}
              />
              <FieldDescription>{t("form.notifyURLHint")}</FieldDescription>
              {errors.notify_url ? (
                <FieldError>
                  {translateMockShipmentValidationMessage(
                    errors.notify_url.message,
                    t,
                  )}
                </FieldError>
              ) : null}
            </Field>
          </FieldGroup>

          <div className="flex gap-2 rounded-lg border border-dashed p-3 text-xs leading-relaxed text-muted-foreground">
            <Sparkles className="mt-0.5 size-3.5 shrink-0" aria-hidden />
            <p>{t("form.autofillHint")}</p>
          </div>

          <div className="flex flex-col-reverse gap-2 border-t pt-5 sm:flex-row sm:justify-end">
            <Button
              type="button"
              variant="ghost"
              disabled={isPending}
              onClick={() => {
                createShipment.reset()
                reset(createMockShipmentInitialValues())
              }}
            >
              {t("form.reset")}
            </Button>
            <Button type="submit" disabled={isPending}>
              {isPending ? (
                <LoaderCircle className="animate-spin" aria-hidden />
              ) : null}
              {isPending ? t("form.submitting") : t("form.submit")}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  )
}
