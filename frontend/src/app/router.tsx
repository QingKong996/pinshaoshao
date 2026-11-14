import {
  Navigate,
  createBrowserRouter,
} from "react-router"
import {
  MerchantConsoleLayout,
  type MerchantConsoleRouteHandle,
} from "./layouts/merchant-console-layout"
import { RootLayout } from "./layouts/root-layout"
import {
  MERCHANT_CONSOLE_BASE_PATH,
  MERCHANT_CONSOLE_HOME_PATH,
  MERCHANT_CONSOLE_PATHS,
} from "./merchant-navigation"
import HomePage from "./pages/home"
import NotFoundPage from "./pages/not-found"
import { MerchantCategoriesRoute } from "./routes/merchant-categories-route"
import { MerchantDashboardRoute } from "./routes/merchant-dashboard-route"
import { MerchantInventoryRoute } from "./routes/merchant-inventory-route"
import { MerchantProductDetailRoute } from "./routes/merchant-product-detail-route"
import { MerchantProductsRoute } from "./routes/merchant-products-route"
import {
  SignInPage,
  SignUpPage,
  MerchantSignInPage,
  MerchantSignUpPage,
} from "@/features/auth"

export const router = createBrowserRouter([
  {
    element: <RootLayout />,
    children: [
      { path: "/", element: <HomePage /> },
      { path: "/signin", element: <SignInPage /> },
      { path: "/signup", element: <SignUpPage /> },
      {
        path: "/merchant/signin",
        element: <MerchantSignInPage redirectTo={MERCHANT_CONSOLE_HOME_PATH} />,
      },
      {
        path: "/merchant/signup",
        element: <MerchantSignUpPage redirectTo={MERCHANT_CONSOLE_HOME_PATH} />,
      },
      {
        path: MERCHANT_CONSOLE_BASE_PATH,
        element: <MerchantConsoleLayout />,
        children: [
          {
            index: true,
            element: <Navigate to={MERCHANT_CONSOLE_HOME_PATH} replace />,
          },
          {
            path: MERCHANT_CONSOLE_PATHS.dashboard,
            element: <MerchantDashboardRoute />,
            handle: {
              breadcrumbs: [
                {
                  to: MERCHANT_CONSOLE_PATHS.dashboard,
                  labelKey: "console:nav.dashboard",
                },
              ],
            } satisfies MerchantConsoleRouteHandle,
          },
          {
            path: MERCHANT_CONSOLE_PATHS.products,
            handle: {
              breadcrumbs: [
                {
                  to: MERCHANT_CONSOLE_PATHS.products,
                  labelKey: "console:nav.productInfo",
                },
              ],
            } satisfies MerchantConsoleRouteHandle,
            children: [
              { index: true, element: <MerchantProductsRoute /> },
              { path: ":productID", element: <MerchantProductDetailRoute /> },
            ],
          },
          {
            path: MERCHANT_CONSOLE_PATHS.categories,
            element: <MerchantCategoriesRoute />,
            handle: {
              breadcrumbs: [
                {
                  to: MERCHANT_CONSOLE_PATHS.products,
                  labelKey: "console:nav.products",
                },
                {
                  to: MERCHANT_CONSOLE_PATHS.categories,
                  labelKey: "console:nav.categories",
                },
              ],
            } satisfies MerchantConsoleRouteHandle,
          },
          {
            path: MERCHANT_CONSOLE_PATHS.inventory,
            element: <MerchantInventoryRoute />,
            handle: {
              breadcrumbs: [
                {
                  to: MERCHANT_CONSOLE_PATHS.products,
                  labelKey: "console:nav.products",
                },
                {
                  to: MERCHANT_CONSOLE_PATHS.inventory,
                  labelKey: "console:nav.inventory",
                },
              ],
            } satisfies MerchantConsoleRouteHandle,
          },
          {
            path: MERCHANT_CONSOLE_PATHS.shop,
            handle: {
              breadcrumbs: [
                {
                  to: MERCHANT_CONSOLE_PATHS.shop,
                  labelKey: "console:nav.shop",
                },
              ],
            } satisfies MerchantConsoleRouteHandle,
          },
        ],
      },
      { path: "*", element: <NotFoundPage /> },
    ],
  },
])
