import { useTranslation } from "react-i18next"
import { useAuth } from "@/features/auth"
import { FeedPage, FeedSignedOut } from "@/features/feed"
import { SiteFooter } from "../layouts/site-footer"
import { SiteHeader } from "../layouts/site-header"
import { useDocumentTitle } from "@/shared/hooks/use-document-title"

export default function HomePage() {
  const { t } = useTranslation()
  const { session } = useAuth()
  useDocumentTitle(t("home:documentTitle"))

  return (
    <div className="flex min-h-svh flex-col">
      <SiteHeader />
      <main className="mx-auto w-full max-w-6xl flex-1 px-4 py-8 sm:px-6">
        {session ? (
          <FeedPage token={session.token} />
        ) : (
          <FeedSignedOut />
        )}
      </main>
      <SiteFooter />
    </div>
  )
}
