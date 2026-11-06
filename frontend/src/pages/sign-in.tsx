import { useTranslation } from "react-i18next"
import { Link } from "react-router"
import { AuthLayout } from "@/components/auth-layout"
import { SignInForm } from "@/components/auth/sign-in-form"
import { useDocumentTitle } from "@/lib/use-document-title"

export default function SignInPage() {
  const { t } = useTranslation()
  useDocumentTitle(t("auth:documentTitle.signIn"))

  return (
    <AuthLayout
      title={t("auth:signIn.title")}
      description={t("auth:signIn.description")}
      footer={
        <>
          {t("auth:signIn.switchPrompt")}
          <Link
            to="/signup"
            className="font-medium text-foreground underline-offset-4 transition-colors hover:underline"
          >
            {t("auth:signIn.switchLink")}
          </Link>
        </>
      }
    >
      <SignInForm />
    </AuthLayout>
  )
}
