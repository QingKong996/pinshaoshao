import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { AuthLayout } from "@/components/auth-layout"
import { MerchantSignInForm } from "@/components/auth/merchant-sign-in-form"
import { useDocumentTitle } from "@/lib/use-document-title"

export default function MerchantSignInPage() {
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
      <MerchantSignInForm />
    </AuthLayout>
  )
}
