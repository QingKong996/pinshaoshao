import { zodResolver } from "@hookform/resolvers/zod"
import { LoaderCircle } from "lucide-react"
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
import { describeCreatePaymentError } from "../model/errors"
import { useCreateMockPayment } from "../model/mutations"
import { mockPaymentDetailPath } from "../model/paths"
import {
  createMockPaymentInitialValues,
  mockPaymentFormSchema,
  toCreateMockPaymentInput,
  translateMockPaymentValidationMessage,
  type MockPaymentFormValues,
} from "../model/payment-form"

export function CreatePaymentCard() {
  const { t } = useTranslation("mockpayment")
  const navigate = useNavigate()
  const createPayment = useCreateMockPayment()
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<MockPaymentFormValues>({
    resolver: zodResolver(mockPaymentFormSchema),
    defaultValues: createMockPaymentInitialValues(),
  })

  async function onSubmit(values: MockPaymentFormValues) {
    try {
      const payment = await createPayment.mutateAsync(
        toCreateMockPaymentInput(values),
      )
      navigate(mockPaymentDetailPath(payment.payment_no))
    } catch {
      // mutation state renders the error
    }
  }

  const isPending = isSubmitting || createPayment.isPending

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("console.create.title")}</CardTitle>
        <CardDescription>{t("console.create.description")}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-5">
        {createPayment.error ? (
          <QueryError
            title={t("errors.createTitle")}
            description={describeCreatePaymentError(createPayment.error, t)}
            retryLabel={t("actions.dismiss")}
            onRetry={() => createPayment.reset()}
          />
        ) : null}

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
          <FieldGroup>
            <div className="grid gap-5 sm:grid-cols-[minmax(0,1fr)_9rem]">
              <Field
                data-invalid={errors.merchant_order_no ? true : undefined}
              >
                <FieldLabel htmlFor="mockpay-merchant-order-no">
                  {t("form.merchantOrderNo")}
                </FieldLabel>
                <Input
                  id="mockpay-merchant-order-no"
                  autoComplete="off"
                  aria-invalid={errors.merchant_order_no ? true : undefined}
                  {...register("merchant_order_no")}
                />
                <FieldDescription>
                  {t("form.merchantOrderNoHint")}
                </FieldDescription>
                {errors.merchant_order_no ? (
                  <FieldError>
                    {translateMockPaymentValidationMessage(
                      errors.merchant_order_no.message,
                      t,
                    )}
                  </FieldError>
                ) : null}
              </Field>

              <Field data-invalid={errors.currency ? true : undefined}>
                <FieldLabel htmlFor="mockpay-currency">
                  {t("form.currency")}
                </FieldLabel>
                <Input
                  id="mockpay-currency"
                  autoComplete="off"
                  maxLength={3}
                  className="font-mono uppercase"
                  aria-invalid={errors.currency ? true : undefined}
                  {...register("currency")}
                />
                {errors.currency ? (
                  <FieldError>
                    {translateMockPaymentValidationMessage(
                      errors.currency.message,
                      t,
                    )}
                  </FieldError>
                ) : null}
              </Field>
            </div>

            <div className="grid gap-5 sm:grid-cols-[minmax(0,1fr)_12rem]">
              <Field data-invalid={errors.subject ? true : undefined}>
                <FieldLabel htmlFor="mockpay-subject">
                  {t("form.subject")}
                </FieldLabel>
                <Input
                  id="mockpay-subject"
                  autoComplete="off"
                  aria-invalid={errors.subject ? true : undefined}
                  {...register("subject")}
                />
                {errors.subject ? (
                  <FieldError>
                    {translateMockPaymentValidationMessage(
                      errors.subject.message,
                      t,
                    )}
                  </FieldError>
                ) : null}
              </Field>

              <Field data-invalid={errors.amount_cents ? true : undefined}>
                <FieldLabel htmlFor="mockpay-amount">
                  {t("form.amount")}
                </FieldLabel>
                <Input
                  id="mockpay-amount"
                  inputMode="numeric"
                  placeholder="3998"
                  aria-invalid={errors.amount_cents ? true : undefined}
                  {...register("amount_cents")}
                />
                <FieldDescription>{t("form.amountHint")}</FieldDescription>
                {errors.amount_cents ? (
                  <FieldError>
                    {translateMockPaymentValidationMessage(
                      errors.amount_cents.message,
                      t,
                    )}
                  </FieldError>
                ) : null}
              </Field>
            </div>

            <Field data-invalid={errors.notify_url ? true : undefined}>
              <FieldLabel htmlFor="mockpay-notify-url">
                {t("form.notifyURL")}
              </FieldLabel>
              <Input
                id="mockpay-notify-url"
                type="url"
                autoComplete="off"
                className="font-mono text-xs"
                aria-invalid={errors.notify_url ? true : undefined}
                {...register("notify_url")}
              />
              <FieldDescription>{t("form.notifyURLHint")}</FieldDescription>
              {errors.notify_url ? (
                <FieldError>
                  {translateMockPaymentValidationMessage(
                    errors.notify_url.message,
                    t,
                  )}
                </FieldError>
              ) : null}
            </Field>

            <div className="grid gap-5 sm:grid-cols-[minmax(0,1fr)_12rem]">
              <Field data-invalid={errors.return_url ? true : undefined}>
                <FieldLabel htmlFor="mockpay-return-url">
                  {t("form.returnURL")}
                  <span className="font-normal text-muted-foreground">
                    {t("form.optional")}
                  </span>
                </FieldLabel>
                <Input
                  id="mockpay-return-url"
                  type="url"
                  autoComplete="off"
                  className="font-mono text-xs"
                  aria-invalid={errors.return_url ? true : undefined}
                  {...register("return_url")}
                />
                {errors.return_url ? (
                  <FieldError>
                    {translateMockPaymentValidationMessage(
                      errors.return_url.message,
                      t,
                    )}
                  </FieldError>
                ) : null}
              </Field>

              <Field
                data-invalid={errors.expires_in_seconds ? true : undefined}
              >
                <FieldLabel htmlFor="mockpay-expires">
                  {t("form.expiresIn")}
                </FieldLabel>
                <Input
                  id="mockpay-expires"
                  inputMode="numeric"
                  aria-invalid={
                    errors.expires_in_seconds ? true : undefined
                  }
                  {...register("expires_in_seconds")}
                />
                <FieldDescription>{t("form.expiresHint")}</FieldDescription>
                {errors.expires_in_seconds ? (
                  <FieldError>
                    {translateMockPaymentValidationMessage(
                      errors.expires_in_seconds.message,
                      t,
                    )}
                  </FieldError>
                ) : null}
              </Field>
            </div>
          </FieldGroup>

          <div className="flex flex-col-reverse gap-2 border-t pt-5 sm:flex-row sm:justify-end">
            <Button
              type="button"
              variant="ghost"
              disabled={isPending}
              onClick={() => {
                createPayment.reset()
                reset(createMockPaymentInitialValues())
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
