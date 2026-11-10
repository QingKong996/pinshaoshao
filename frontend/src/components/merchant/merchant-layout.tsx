import { useEffect, useRef } from "react"
import { useTranslation } from "react-i18next"
import { Navigate, Outlet, useNavigate } from "react-router"
import { LanguageSwitcher } from "@/components/language-switcher"
import { MerchantSidebar } from "@/components/merchant/merchant-sidebar"
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar"
import { onUnauthorized } from "@/lib/api"
import { useMerchantAuth } from "@/lib/merchant-auth-context"

export function MerchantLayout() {
  const { t } = useTranslation("merchant")
  const { session, signOut } = useMerchantAuth()
  const navigate = useNavigate()
  const handlingUnauthorized = useRef(false)

  useEffect(
    () =>
      onUnauthorized(() => {
        if (handlingUnauthorized.current) {
          return
        }
        handlingUnauthorized.current = true
        void signOut()
        navigate("/merchant/signin", { replace: true })
      }),
    [navigate, signOut],
  )

  if (!session) {
    return <Navigate to="/merchant/signin" replace />
  }

  return (
    <SidebarProvider>
      <MerchantSidebar />
      <SidebarInset>
        <header className="sticky top-0 z-20 flex h-14 shrink-0 items-center gap-2 border-b border-border/60 bg-background/80 px-4 backdrop-blur-xl sm:px-6">
          <SidebarTrigger aria-label={t("layout.toggleNav")} />
          <div className="flex min-w-0 flex-1 items-center" />
          <LanguageSwitcher />
        </header>
        <div className="mx-auto w-full max-w-6xl flex-1 px-4 py-6 sm:px-6 lg:py-8">
          <Outlet />
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
