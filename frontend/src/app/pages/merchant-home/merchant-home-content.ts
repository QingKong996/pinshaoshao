import {
  LayoutDashboard,
  Package,
  Users,
  Warehouse,
} from "lucide-react"

export const FEATURES = [
  {
    key: "reach",
    icon: Users,
    titleKey: "home:merchant.features.items.reach.title",
    descriptionKey: "home:merchant.features.items.reach.description",
  },
  {
    key: "catalog",
    icon: Package,
    titleKey: "home:merchant.features.items.catalog.title",
    descriptionKey: "home:merchant.features.items.catalog.description",
  },
  {
    key: "inventory",
    icon: Warehouse,
    titleKey: "home:merchant.features.items.inventory.title",
    descriptionKey: "home:merchant.features.items.inventory.description",
  },
  {
    key: "overview",
    icon: LayoutDashboard,
    titleKey: "home:merchant.features.items.overview.title",
    descriptionKey: "home:merchant.features.items.overview.description",
  },
] as const

export const STEPS = [
  {
    key: "register",
    titleKey: "home:merchant.howItWorks.steps.register.title",
    descriptionKey: "home:merchant.howItWorks.steps.register.description",
  },
  {
    key: "publish",
    titleKey: "home:merchant.howItWorks.steps.publish.title",
    descriptionKey: "home:merchant.howItWorks.steps.publish.description",
  },
  {
    key: "manage",
    titleKey: "home:merchant.howItWorks.steps.manage.title",
    descriptionKey: "home:merchant.howItWorks.steps.manage.description",
  },
] as const

export const PREVIEW_PRODUCTS = [
  {
    key: "coffee",
    nameKey: "home:merchant.preview.recent.items.coffee.name",
    metaKey: "home:merchant.preview.recent.items.coffee.meta",
    price: "¥68",
  },
  {
    key: "bag",
    nameKey: "home:merchant.preview.recent.items.bag.name",
    metaKey: "home:merchant.preview.recent.items.bag.meta",
    price: "¥129",
  },
] as const

export const TRUST_POINTS = [
  "home:merchant.hero.trustPoints.free",
  "home:merchant.hero.trustPoints.management",
  "home:merchant.hero.trustPoints.insight",
] as const

export const ASSURANCES = [
  "home:merchant.closingCta.assurances.free",
  "home:merchant.closingCta.assurances.setup",
  "home:merchant.closingCta.assurances.mobile",
] as const
