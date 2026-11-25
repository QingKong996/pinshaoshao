import { useTranslation } from "react-i18next"
import { Navigate } from "react-router"
import { MockSignInForm } from "@/mock/auth/sign-in-form"
import { useDocumentTitle } from "@/shared/hooks/use-document-title"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/shared/ui/card"
import { mockPaymentAuth } from "../model/auth"
import { MOCK_PAYMENT_ROUTE_PATHS } from "../model/paths"
import { MockPaymentShell } from "./mock-payment-shell"

const { useMockAuth } = mockPaymentAuth

export function MockPaymentSignInPage() {
  const { t } = useTranslation("mockpayment")
  const { session, signIn } = useMockAuth()
  useDocumentTitle(t("documentTitle.signIn"))

  if (session) {
    return <Navigate to={MOCK_PAYMENT_ROUTE_PATHS.list} replace />
  }

  return (
    <MockPaymentShell>
      <div className="mx-auto w-full max-w-md">
        <Card>
          <CardHeader>
            <CardTitle>{t("auth.title")}</CardTitle>
            <CardDescription>{t("auth.description")}</CardDescription>
          </CardHeader>
          <CardContent>
            <MockSignInForm
              fallbackPath={MOCK_PAYMENT_ROUTE_PATHS.list}
              signIn={signIn}
              labels={{
                username: t("auth.username"),
                password: t("auth.password"),
                submit: t("auth.submit"),
                submitting: t("auth.submitting"),
                invalidCredentials: t("auth.invalidCredentials"),
                unavailable: t("auth.unavailable"),
              }}
            />
          </CardContent>
        </Card>
      </div>
    </MockPaymentShell>
  )
}
