import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { AuthLayout } from "@/components/auth-layout"
import { MerchantSignUpForm } from "@/components/auth/merchant-sign-up-form"
import { useDocumentTitle } from "@/lib/use-document-title"

export default function MerchantSignUpPage() {
  const { t } = useTranslation()
  useDocumentTitle(t("auth:documentTitle.merchantSignUp"))

  return (
    <AuthLayout
      variant="merchant"
      title={t("auth:merchant.signUp.title")}
      description={t("auth:merchant.signUp.description")}
      footer={
        <>
          {t("auth:merchant.signUp.switchPrompt")}
          <Link
            to="/merchant/signin"
            className="font-medium text-foreground underline-offset-4 transition-colors hover:underline"
          >
            {t("auth:merchant.signUp.switchLink")}
          </Link>
        </>
      }
    >
      <MerchantSignUpForm />
    </AuthLayout>
  )
}
