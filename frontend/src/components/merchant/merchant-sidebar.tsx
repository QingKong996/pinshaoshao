import { ChevronsUpDown, LogOut, Store } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Link, useLocation, useNavigate } from "react-router"
import { BrandMark } from "@/components/brand-mark"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/components/ui/sidebar"
import { isMerchantNavItemActive, MERCHANT_NAV } from "@/lib/merchant-nav"
import { useMerchantAuth } from "@/lib/merchant-auth-context"

export function MerchantSidebar() {
  const { t } = useTranslation("merchant")
  const { session, signOut } = useMerchantAuth()
  const { isMobile, setOpenMobile } = useSidebar()
  const location = useLocation()
  const navigate = useNavigate()

  function closeMobileNav() {
    if (isMobile) {
      setOpenMobile(false)
    }
  }

  async function handleSignOut() {
    await signOut()
    navigate("/merchant/signin", { replace: true })
  }

  return (
    <Sidebar
      collapsible="offcanvas"
      mobileNavTitle={t("layout.navLabel")}
      mobileNavDescription={t("nav.group")}
    >
      <SidebarHeader className="border-b border-sidebar-border">
        <div className="flex flex-col gap-1 px-1.5 py-1">
          <BrandMark to="/merchant" />
          <span className="text-xs text-muted-foreground">
            {t("layout.console")}
          </span>
        </div>
      </SidebarHeader>

      <SidebarContent>
        <nav
          aria-label={t("layout.navLabel")}
          className="flex flex-col gap-2"
        >
          <SidebarGroup>
            <SidebarGroupLabel>{t("nav.group")}</SidebarGroupLabel>
            <SidebarGroupContent>
              <SidebarMenu>
                {MERCHANT_NAV.map((item) => {
                  const Icon = item.icon
                  const isActive = isMerchantNavItemActive(
                    item,
                    location.pathname,
                  )
                  return (
                    <SidebarMenuItem key={item.to}>
                      <SidebarMenuButton
                        render={
                          <Link
                            to={item.to}
                            aria-current={isActive ? "page" : undefined}
                          />
                        }
                        isActive={isActive}
                        onClick={closeMobileNav}
                        className="h-9"
                      >
                        <Icon aria-hidden />
                        <span>{t(item.labelKey)}</span>
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                  )
                })}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        </nav>
      </SidebarContent>

      <SidebarFooter className="border-t border-sidebar-border">
        <SidebarMenu>
          <SidebarMenuItem>
            <DropdownMenu>
              <DropdownMenuTrigger
                aria-label={t("account.menu")}
                className="flex w-full items-center gap-2 rounded-md p-1.5 text-left outline-none transition-colors hover:bg-sidebar-accent focus-visible:ring-2 focus-visible:ring-sidebar-ring"
              >
                <Avatar>
                  <AvatarFallback>
                    {session?.username.slice(0, 1).toUpperCase() ?? ""}
                  </AvatarFallback>
                </Avatar>
                <span className="flex min-w-0 flex-1 flex-col">
                  <span className="truncate text-sm font-medium">
                    {session?.username}
                  </span>
                  <span className="truncate text-xs text-muted-foreground">
                    {t("layout.console")}
                  </span>
                </span>
                <ChevronsUpDown
                  className="size-4 shrink-0 text-muted-foreground"
                  aria-hidden
                />
              </DropdownMenuTrigger>
              <DropdownMenuContent side="top" align="start" className="w-56">
                <DropdownMenuItem
                  onClick={() => {
                    closeMobileNav()
                    navigate("/")
                  }}
                >
                  <Store aria-hidden />
                  {t("account.backToStore")}
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem onClick={() => void handleSignOut()}>
                  <LogOut aria-hidden />
                  {t("account.signOut")}
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
    </Sidebar>
  )
}
