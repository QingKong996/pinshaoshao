import { defineConfig, mergeConfig } from "vitest/config"
import viteConfig from "./vite.config.ts"

export default mergeConfig(
  viteConfig,
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
          "src/features/auth/model/**/*.{ts,tsx}",
          "src/features/merchant-dashboard/model/**/*.ts",
          "src/features/merchant-products/model/**/*.ts",
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
