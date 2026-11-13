import {
  LayoutDashboard,
  Package,
  Store,
  type LucideIcon,
} from "lucide-react"

export const MERCHANT_CONSOLE_BASE_PATH = "/merchant"

export const MERCHANT_CONSOLE_PATHS = {
  dashboard: "/merchant/dashboard",
  products: "/merchant/products",
  categories: "/merchant/categories",
  inventory: "/merchant/inventory",
  shop: "/merchant/shop",
} as const

export const MERCHANT_CONSOLE_HOME_PATH = MERCHANT_CONSOLE_PATHS.dashboard

export type MerchantConsoleNavLink = {
  to: string
  labelKey: `console:nav.${keyof typeof MERCHANT_CONSOLE_PATHS}`
}

export const MERCHANT_CONSOLE_NAV: (MerchantConsoleNavLink & {
  icon: LucideIcon
  children?: MerchantConsoleNavLink[]
})[] = [
  {
    to: MERCHANT_CONSOLE_PATHS.dashboard,
    icon: LayoutDashboard,
    labelKey: "console:nav.dashboard",
  },
  {
    to: MERCHANT_CONSOLE_PATHS.products,
    icon: Package,
    labelKey: "console:nav.products",
    children: [
      { to: MERCHANT_CONSOLE_PATHS.categories, labelKey: "console:nav.categories" },
      { to: MERCHANT_CONSOLE_PATHS.inventory, labelKey: "console:nav.inventory" },
    ],
  },
  {
    to: MERCHANT_CONSOLE_PATHS.shop,
    icon: Store,
    labelKey: "console:nav.shop",
  },
]

export function merchantConsoleProductPath(productID: number) {
  return `${MERCHANT_CONSOLE_PATHS.products}/${productID}`
}
