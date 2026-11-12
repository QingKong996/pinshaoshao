import { useEffect, useRef } from "react"
import { useTranslation } from "react-i18next"
import {
  Link,
  Navigate,
  Outlet,
  useLocation,
  useNavigate,
} from "react-router"
import { BrandMark } from "@/shared/ui/brand-mark"
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/shared/ui/breadcrumb"
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
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
  SidebarProvider,
  SidebarRail,
  SidebarTrigger,
} from "@/shared/ui/sidebar"
import { onUnauthorized } from "@/shared/api/client"
import { useMerchantAuth } from "@/features/auth"
import {
  findMerchantConsoleSection,
  findMerchantConsoleSectionBySlug,
  isMerchantConsoleSectionActive,
  MERCHANT_CONSOLE_HOME_PATH,
  MERCHANT_CONSOLE_SECTIONS,
  merchantConsoleChildSections,
  merchantConsoleSectionPath,
} from "../merchant-navigation"
import { cn } from "@/shared/lib/utils"

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
        <header className="sticky top-0 z-10 flex h-14 shrink-0 items-center gap-2 border-b border-border/60 bg-background/80 px-4 backdrop-blur-xl md:px-6">
          <MerchantConsoleBreadcrumb />
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
              {MERCHANT_CONSOLE_SECTIONS.filter((section) => !section.parentSlug).map(
                ({ slug, icon: Icon, labelKey }) => {
                  const label = t(labelKey)
                  const children = merchantConsoleChildSections(slug)

                  return (
                    <SidebarMenuItem key={slug}>
                      <SidebarMenuButton
                        isActive={isMerchantConsoleSectionActive(pathname, slug)}
                        tooltip={label}
                        className="h-10 text-lg [&_svg]:size-4"
                        render={<Link to={merchantConsoleSectionPath(slug)} />}
                      >
                        <Icon />
                        <span>{label}</span>
                      </SidebarMenuButton>
                      {children.length > 0 ? (
                        <SidebarMenuSub>
                          {children.map((child) => (
                            <SidebarMenuSubItem key={child.slug}>
                              <SidebarMenuSubButton
                                isActive={isMerchantConsoleSectionActive(
                                  pathname,
                                  child.slug,
                                )}
                                className="h-9 text-lg"
                                render={
                                  <Link
                                    to={merchantConsoleSectionPath(child.slug)}
                                  />
                                }
                              >
                                <span>{t(child.labelKey)}</span>
                              </SidebarMenuSubButton>
                            </SidebarMenuSubItem>
                          ))}
                        </SidebarMenuSub>
                      ) : null}
                    </SidebarMenuItem>
                  )
                },
              )}
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
  const parent = section?.parentSlug
    ? findMerchantConsoleSectionBySlug(section.parentSlug)
    : undefined

  return (
    <Breadcrumb className={cn(className)}>
      <BreadcrumbList>
        <BreadcrumbItem>
          <BreadcrumbLink render={<Link to={MERCHANT_CONSOLE_HOME_PATH} />}>
            {t("console:title")}
          </BreadcrumbLink>
        </BreadcrumbItem>
        {parent ? (
          <>
            <BreadcrumbSeparator />
            <BreadcrumbItem>
              <BreadcrumbLink
                render={
                  <Link to={merchantConsoleSectionPath(parent.slug)} />
                }
              >
                {t(parent.labelKey)}
              </BreadcrumbLink>
            </BreadcrumbItem>
          </>
        ) : null}
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
