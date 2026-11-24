import { useState, type FormEvent } from "react"
import { Search } from "lucide-react"
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
import { Field, FieldError, FieldLabel } from "@/shared/ui/field"
import { Input } from "@/shared/ui/input"
import { mockPaymentDetailPath } from "../model/paths"

export function PaymentLookupCard() {
  const { t } = useTranslation("mockpayment")
  const navigate = useNavigate()
  const [paymentNo, setPaymentNo] = useState("")
  const [error, setError] = useState(false)

  function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const value = paymentNo.trim()
    if (!value) {
      setError(true)
      return
    }
    navigate(mockPaymentDetailPath(value))
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("console.lookup.title")}</CardTitle>
        <CardDescription>
          {t("console.lookup.description")}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={onSubmit} className="space-y-4">
          <Field data-invalid={error ? true : undefined}>
            <FieldLabel htmlFor="mockpay-payment-no">
              {t("console.lookup.label")}
            </FieldLabel>
            <Input
              id="mockpay-payment-no"
              autoComplete="off"
              className="font-mono"
              placeholder={t("console.lookup.placeholder")}
              value={paymentNo}
              aria-invalid={error ? true : undefined}
              onChange={(event) => {
                setPaymentNo(event.target.value)
                setError(false)
              }}
            />
            {error ? (
              <FieldError>{t("console.lookup.required")}</FieldError>
            ) : null}
          </Field>
          <Button type="submit" variant="outline" className="w-full">
            <Search aria-hidden />
            {t("console.lookup.submit")}
          </Button>
        </form>
      </CardContent>
    </Card>
  )
}
