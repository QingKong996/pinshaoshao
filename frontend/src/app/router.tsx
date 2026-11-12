import {
  Navigate,
  createBrowserRouter,
} from "react-router"
import { MerchantConsoleLayout } from "./layouts/merchant-console-layout"
import { RootLayout } from "./layouts/root-layout"
import {
  MERCHANT_CONSOLE_BASE_PATH,
  MERCHANT_CONSOLE_HOME_PATH,
  MERCHANT_CONSOLE_SECTIONS,
} from "./merchant-navigation"
import HomePage from "./pages/home"
import NotFoundPage from "./pages/not-found"
import { MerchantDashboardRoute } from "./routes/merchant-dashboard-route"
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
          // Remaining console sections stay empty until their pages land.
          ...MERCHANT_CONSOLE_SECTIONS.map(({ slug }) => {
            if (slug === "dashboard") {
              return { path: slug, element: <MerchantDashboardRoute /> }
            }
            if (slug === "products") {
              return { path: slug, element: <MerchantProductsRoute /> }
            }
            return { path: slug }
          }),
          {
            path: "products/:productID",
            element: <MerchantProductDetailRoute />,
          },
        ],
      },
      { path: "*", element: <NotFoundPage /> },
    ],
  },
])
