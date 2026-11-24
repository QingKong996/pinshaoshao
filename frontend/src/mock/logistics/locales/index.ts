import en from "./en.json"
import zhCN from "./zh-CN.json"

export const mockLogisticsResources = {
  "zh-CN": { mocklogistics: zhCN },
  en: { mocklogistics: en },
} as const
