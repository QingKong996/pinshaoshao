import { useTranslation } from "react-i18next"
import { AuthLayout } from "../auth-layout"
import { AuthSwitchPrompt } from "../auth-switch-prompt"
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
        <AuthSwitchPrompt
          prompt={t("auth:merchant.signIn.switchPrompt")}
          linkLabel={t("auth:merchant.signIn.switchLink")}
          to="/merchant/signup"
        />
      }
    >
      <MerchantSignInForm redirectTo={redirectTo} />
    </AuthLayout>
  )
}
