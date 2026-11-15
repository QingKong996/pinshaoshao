import { useTranslation } from "react-i18next"
import { AuthLayout } from "../auth-layout"
import { AuthSwitchPrompt } from "../auth-switch-prompt"
import { SignUpForm } from "../sign-up-form"
import { useDocumentTitle } from "@/shared/hooks/use-document-title"

export default function SignUpPage() {
  const { t } = useTranslation()
  useDocumentTitle(t("auth:documentTitle.signUp"))

  return (
    <AuthLayout
      title={t("auth:signUp.title")}
      description={t("auth:signUp.description")}
      footer={
        <AuthSwitchPrompt
          prompt={t("auth:signUp.switchPrompt")}
          linkLabel={t("auth:signUp.switchLink")}
          to="/signin"
        />
      }
    >
      <SignUpForm />
    </AuthLayout>
  )
}
