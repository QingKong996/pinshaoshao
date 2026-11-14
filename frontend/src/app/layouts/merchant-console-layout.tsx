import { Fragment, useEffect, useRef } from "react"
import { useTranslation } from "react-i18next"
import {
  Link,
  NavLink,
  Navigate,
  Outlet,
  useMatches,
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
  MERCHANT_CONSOLE_HOME_PATH,
  MERCHANT_CONSOLE_NAV,
  type MerchantConsoleNavLink,
} from "../merchant-navigation"
import { cn } from "@/shared/lib/utils"

export type MerchantConsoleRouteHandle = {
  breadcrumbs: MerchantConsoleNavLink[]
}

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
              {MERCHANT_CONSOLE_NAV.map((item) => {
                const { icon: Icon, labelKey, children } = item
                const label = t(labelKey)

                return (
                  <SidebarMenuItem key={item.to ?? labelKey}>
                    {children ? (
                      <SidebarMenuButton
                        tooltip={label}
                        className="h-10 cursor-default text-base hover:bg-transparent hover:text-sidebar-foreground group-has-[[aria-current=page]]/menu-item:bg-sidebar-accent group-has-[[aria-current=page]]/menu-item:font-medium group-has-[[aria-current=page]]/menu-item:text-sidebar-accent-foreground [&_svg]:size-4"
                        render={<div />}
                      >
                        <Icon />
                        <span>{label}</span>
                      </SidebarMenuButton>
                    ) : (
                      <SidebarMenuButton
                        tooltip={label}
                        className="h-10 text-base group-has-[[aria-current=page]]/menu-item:bg-sidebar-accent group-has-[[aria-current=page]]/menu-item:font-medium group-has-[[aria-current=page]]/menu-item:text-sidebar-accent-foreground [&_svg]:size-4"
                        render={<NavLink to={item.to} />}
                      >
                        <Icon />
                        <span>{label}</span>
                      </SidebarMenuButton>
                    )}
                    {children ? (
                      <SidebarMenuSub>
                        {children.map((child) => (
                          <SidebarMenuSubItem key={child.to}>
                            <SidebarMenuSubButton
                              className="h-9 text-lg aria-[current=page]:bg-sidebar-accent aria-[current=page]:text-sidebar-accent-foreground"
                              render={<NavLink to={child.to} />}
                            >
                              <span>{t(child.labelKey)}</span>
                            </SidebarMenuSubButton>
                          </SidebarMenuSubItem>
                        ))}
                      </SidebarMenuSub>
                    ) : null}
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
  const breadcrumbs = useMatches().flatMap(
    ({ handle }) =>
      (handle as MerchantConsoleRouteHandle | undefined)?.breadcrumbs ?? [],
  )

  return (
    <Breadcrumb className={cn(className)}>
      <BreadcrumbList>
        <BreadcrumbItem>
          <BreadcrumbLink render={<Link to={MERCHANT_CONSOLE_HOME_PATH} />}>
            {t("console:title")}
          </BreadcrumbLink>
        </BreadcrumbItem>
        {breadcrumbs.map((item, index) => (
          <Fragment key={item.to}>
            <BreadcrumbSeparator />
            <BreadcrumbItem>
              {index === breadcrumbs.length - 1 ? (
                <BreadcrumbPage>{t(item.labelKey)}</BreadcrumbPage>
              ) : (
                <BreadcrumbLink render={<Link to={item.to} />}>
                  {t(item.labelKey)}
                </BreadcrumbLink>
              )}
            </BreadcrumbItem>
          </Fragment>
        ))}
      </BreadcrumbList>
    </Breadcrumb>
  )
}
