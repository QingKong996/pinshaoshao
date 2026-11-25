import { defineConfig, mergeConfig } from "vitest/config"
import viteConfig from "./vite.config.ts"

const baseConfig = viteConfig({ command: "serve", mode: "test" })

export default mergeConfig(
  baseConfig,
  defineConfig({
    test: {
      environment: "jsdom",
      environmentOptions: {
        jsdom: {
          url: "http://localhost/",
        },
      },
      setupFiles: ["./src/test/setup.ts"],
      include: ["src/**/*.test.{ts,tsx}"],
      restoreMocks: true,
      clearMocks: true,
      coverage: {
        provider: "v8",
        reporter: ["text", "lcov"],
        reportsDirectory: "./coverage",
        include: [
          "src/shared/api/**/*.ts",
          "src/shared/i18n/config.ts",
          "src/shared/lib/format.ts",
          "src/app/routes/checkout-route.tsx",
          "src/app/routes/product-detail-route.tsx",
          "src/features/auth/model/**/*.{ts,tsx}",
          "src/features/cart/model/**/*.{ts,tsx}",
          "src/features/feed/model/**/*.{ts,tsx}",
          "src/features/merchant-dashboard/model/**/*.ts",
          "src/features/merchant-orders/model/**/*.ts",
          "src/features/merchant-products/model/**/*.ts",
          "src/features/order/model/**/*.{ts,tsx}",
        ],
        exclude: ["**/*.test.{ts,tsx}"],
        thresholds: {
          lines: 85,
          functions: 85,
          branches: 80,
          statements: 85,
        },
      },
    },
  }),
)
