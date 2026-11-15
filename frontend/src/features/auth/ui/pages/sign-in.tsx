import { useTranslation } from "react-i18next"
import { AuthLayout } from "../auth-layout"
import { AuthSwitchPrompt } from "../auth-switch-prompt"
import { SignInForm } from "../sign-in-form"
import { useDocumentTitle } from "@/shared/hooks/use-document-title"

export default function SignInPage() {
  const { t } = useTranslation()
  useDocumentTitle(t("auth:documentTitle.signIn"))

  return (
    <AuthLayout
      title={t("auth:signIn.title")}
      description={t("auth:signIn.description")}
      footer={
        <AuthSwitchPrompt
          prompt={t("auth:signIn.switchPrompt")}
          linkLabel={t("auth:signIn.switchLink")}
          to="/signup"
        />
      }
    >
      <SignInForm />
    </AuthLayout>
  )
}
