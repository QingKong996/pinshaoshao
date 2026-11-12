import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { AuthLayout } from "../auth-layout"
import { MerchantSignInForm } from "../merchant-sign-in-form"
import { useDocumentTitle } from "@/shared/hooks/use-document-title"

export default function MerchantSignInPage({ redirectTo }: { redirectTo: string }) {
  const { t } = useTranslation()
  useDocumentTitle(t("auth:documentTitle.merchantSignIn"))

  return (
    <AuthLayout
      variant="merchant"
      title={t("auth:merchant.signIn.title")}
      description={t("auth:merchant.signIn.description")}
      footer={
        <>
          {t("auth:merchant.signIn.switchPrompt")}
          <Link
            to="/merchant/signup"
            className="font-medium text-foreground underline-offset-4 transition-colors hover:underline"
          >
            {t("auth:merchant.signIn.switchLink")}
          </Link>
        </>
      }
    >
      <MerchantSignInForm redirectTo={redirectTo} />
    </AuthLayout>
  )
}
