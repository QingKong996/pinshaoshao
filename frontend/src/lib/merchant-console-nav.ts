import {
  Boxes,
  FolderTree,
  LayoutDashboard,
  Package,
  Store,
  type LucideIcon,
} from "lucide-react"

export const MERCHANT_CONSOLE_BASE_PATH = "/merchant"

export type MerchantConsoleLabelKey =
  | "console:nav.dashboard"
  | "console:nav.products"
  | "console:nav.categories"
  | "console:nav.inventory"
  | "console:nav.shop"

export type MerchantConsoleSection = {
  slug: string
  icon: LucideIcon
  labelKey: MerchantConsoleLabelKey
}

export const MERCHANT_CONSOLE_SECTIONS: MerchantConsoleSection[] = [
  {
    slug: "dashboard",
    icon: LayoutDashboard,
    labelKey: "console:nav.dashboard",
  },
  { slug: "products", icon: Package, labelKey: "console:nav.products" },
  { slug: "categories", icon: FolderTree, labelKey: "console:nav.categories" },
  { slug: "inventory", icon: Boxes, labelKey: "console:nav.inventory" },
  { slug: "shop", icon: Store, labelKey: "console:nav.shop" },
]

export function merchantConsoleSectionPath(slug: string) {
  return `${MERCHANT_CONSOLE_BASE_PATH}/${slug}`
}

export const MERCHANT_CONSOLE_HOME_PATH = merchantConsoleSectionPath(
  MERCHANT_CONSOLE_SECTIONS[0].slug,
)

export function findMerchantConsoleSection(pathname: string) {
  return MERCHANT_CONSOLE_SECTIONS.find(({ slug }) => {
    const path = merchantConsoleSectionPath(slug)
    return pathname === path || pathname.startsWith(`${path}/`)
  })
}
