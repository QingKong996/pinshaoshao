import { useTranslation } from "react-i18next"
import { AuthLayout } from "../auth-layout"
import { AuthSwitchPrompt } from "../auth-switch-prompt"
import { MerchantSignUpForm } from "../merchant-sign-up-form"
import { useDocumentTitle } from "@/shared/hooks/use-document-title"

export default function MerchantSignUpPage({ redirectTo }: { redirectTo: string }) {
  const { t } = useTranslation()
  useDocumentTitle(t("auth:documentTitle.merchantSignUp"))

  return (
    <AuthLayout
      variant="merchant"
      title={t("auth:merchant.signUp.title")}
      description={t("auth:merchant.signUp.description")}
      footer={
        <AuthSwitchPrompt
          prompt={t("auth:merchant.signUp.switchPrompt")}
          linkLabel={t("auth:merchant.signUp.switchLink")}
          to="/merchant/signin"
        />
      }
    >
      <MerchantSignUpForm redirectTo={redirectTo} />
    </AuthLayout>
  )
}
