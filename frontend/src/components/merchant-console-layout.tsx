import { useEffect, useRef } from "react"
import { useTranslation } from "react-i18next"
import {
  Link,
  Navigate,
  Outlet,
  useLocation,
  useNavigate,
} from "react-router"
import { BrandMark } from "@/components/brand-mark"
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb"
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarRail,
  SidebarTrigger,
} from "@/components/ui/sidebar"
import { onUnauthorized } from "@/lib/api"
import { useMerchantAuth } from "@/lib/merchant-auth-context"
import {
  findMerchantConsoleSection,
  MERCHANT_CONSOLE_HOME_PATH,
  MERCHANT_CONSOLE_SECTIONS,
  merchantConsoleSectionPath,
} from "@/lib/merchant-console-nav"
import { cn } from "@/lib/utils"

export function MerchantConsoleLayout() {
  const { session, signOut } = useMerchantAuth()
  const navigate = useNavigate()
  const handlingUnauthorized = useRef(false)

  useEffect(() => {
    if (!session) {
      return
    }
    return onUnauthorized(() => {
      if (handlingUnauthorized.current) {
        return
      }
      handlingUnauthorized.current = true
      void signOut()
      navigate("/merchant/signin", { replace: true })
    })
  }, [navigate, session, signOut])

  if (!session) {
    return <Navigate to="/merchant/signin" replace />
  }

  return (
    <SidebarProvider>
      <MerchantConsoleSidebar />
      <SidebarInset>
        <header className="sticky top-0 z-10 flex h-14 shrink-0 items-center gap-2 border-b border-border/60 bg-background/80 px-4 backdrop-blur-xl">
          <MerchantConsoleBreadcrumb className="ml-auto" />
        </header>
        <div className="flex flex-1 flex-col gap-4 p-4 md:p-6">
          <Outlet />
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}

function MerchantConsoleSidebar() {
  const { t } = useTranslation()
  const { pathname } = useLocation()

  return (
    <Sidebar
      collapsible="icon"
      mobileNavTitle={t("console:title")}
      mobileNavDescription={t("console:sidebarDescription")}
    >
      <SidebarHeader className="px-4 pt-4 group-data-[collapsible=icon]:px-2">
        <div className="flex items-center gap-1 group-data-[collapsible=icon]:justify-center">
          <BrandMark className="group-data-[collapsible=icon]:hidden" />
          <SidebarTrigger className="ml-auto group-data-[collapsible=icon]:ml-0" />
        </div>
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              {MERCHANT_CONSOLE_SECTIONS.map(({ slug, icon: Icon, labelKey }) => {
                const to = merchantConsoleSectionPath(slug)
                const label = t(labelKey)
                const isActive =
                  pathname === to || pathname.startsWith(`${to}/`)

                return (
                  <SidebarMenuItem key={slug}>
                    <SidebarMenuButton
                      isActive={isActive}
                      tooltip={label}
                      className="[&_svg]:size-3.5"
                      render={<Link to={to} />}
                    >
                      <Icon />
                      <span>{label}</span>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                )
              })}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
      <SidebarRail />
    </Sidebar>
  )
}

function MerchantConsoleBreadcrumb({ className }: { className?: string }) {
  const { t } = useTranslation()
  const { pathname } = useLocation()
  const section = findMerchantConsoleSection(pathname)

  return (
    <Breadcrumb className={cn(className)}>
      <BreadcrumbList>
        <BreadcrumbItem>
          <BreadcrumbLink render={<Link to={MERCHANT_CONSOLE_HOME_PATH} />}>
            {t("console:title")}
          </BreadcrumbLink>
        </BreadcrumbItem>
        {section ? (
          <>
            <BreadcrumbSeparator />
            <BreadcrumbItem>
              <BreadcrumbPage>{t(section.labelKey)}</BreadcrumbPage>
            </BreadcrumbItem>
          </>
        ) : null}
      </BreadcrumbList>
    </Breadcrumb>
  )
}
