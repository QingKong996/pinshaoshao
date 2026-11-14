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

type MerchantConsoleNavKey =
  | keyof typeof MERCHANT_CONSOLE_PATHS
  | "productInfo"

export type MerchantConsoleNavLink = {
  to: string
  labelKey: `console:nav.${MerchantConsoleNavKey}`
}

export type MerchantConsoleNavItem = {
  icon: LucideIcon
  labelKey: `console:nav.${MerchantConsoleNavKey}`
} & (
  | {
      to: string
      children?: never
    }
  | {
      to?: never
      children: MerchantConsoleNavLink[]
    }
)

export const MERCHANT_CONSOLE_NAV: MerchantConsoleNavItem[] = [
  {
    to: MERCHANT_CONSOLE_PATHS.dashboard,
    icon: LayoutDashboard,
    labelKey: "console:nav.dashboard",
  },
  {
    icon: Package,
    labelKey: "console:nav.products",
    children: [
      {
        to: MERCHANT_CONSOLE_PATHS.products,
        labelKey: "console:nav.productInfo",
      },
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
