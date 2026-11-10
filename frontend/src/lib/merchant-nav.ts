import type { LucideIcon } from "lucide-react"
import { Boxes, LayoutDashboard, Package, Store, Tags } from "lucide-react"
import type { ParseKeys } from "i18next"

export type MerchantNavItem = {
  to: string
  labelKey: ParseKeys<"merchant">
  icon: LucideIcon
  end?: boolean
}

export const MERCHANT_NAV = [
  {
    to: "/merchant",
    labelKey: "nav.dashboard",
    icon: LayoutDashboard,
    end: true,
  },
  {
    to: "/merchant/products",
    labelKey: "nav.products",
    icon: Package,
  },
  {
    to: "/merchant/categories",
    labelKey: "nav.categories",
    icon: Tags,
  },
  {
    to: "/merchant/inventory",
    labelKey: "nav.inventory",
    icon: Boxes,
  },
  {
    to: "/merchant/shop",
    labelKey: "nav.shop",
    icon: Store,
  },
] as const satisfies readonly MerchantNavItem[]

export function isMerchantNavItemActive(
  item: MerchantNavItem,
  pathname: string,
): boolean {
  if (item.end) {
    return pathname === item.to
  }
  return pathname === item.to || pathname.startsWith(`${item.to}/`)
}
