import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { AuthLayout } from "@/components/auth-layout"
import { SignUpForm } from "@/components/auth/sign-up-form"
import { useDocumentTitle } from "@/hooks/use-document-title"

export default function SignUpPage() {
  const { t } = useTranslation()
  useDocumentTitle(t("auth:documentTitle.signUp"))

  return (
    <AuthLayout
      title={t("auth:signUp.title")}
      description={t("auth:signUp.description")}
      footer={
        <>
          {t("auth:signUp.switchPrompt")}
          <Link
            to="/signin"
            className="font-medium text-foreground underline-offset-4 transition-colors hover:underline"
          >
            {t("auth:signUp.switchLink")}
          </Link>
        </>
      }
    >
      <SignUpForm />
    </AuthLayout>
  )
}
