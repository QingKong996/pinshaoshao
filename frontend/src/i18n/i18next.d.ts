import type { NAMESPACES, resources } from "@/i18n"

declare module "i18next" {
  interface CustomTypeOptions {
    // listing all namespaces lets t() type-check explicitly prefixed keys, e.g. "home:hero.badge"
    defaultNS: typeof NAMESPACES
    resources: (typeof resources)["zh-CN"]
  }
}
