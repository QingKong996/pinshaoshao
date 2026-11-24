import { createBrowserRouter } from "react-router"
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
import MerchantHomePage from "./pages/merchant-home/merchant-home-page"
import NotFoundPage from "./pages/not-found"
import { CartRoute } from "./routes/cart-route"
import { CheckoutRoute } from "./routes/checkout-route"
import { MerchantCategoriesRoute } from "./routes/merchant-categories-route"
import { MerchantDashboardRoute } from "./routes/merchant-dashboard-route"
import { MerchantInventoryRoute } from "./routes/merchant-inventory-route"
import { MerchantOrdersRoute } from "./routes/merchant-orders-route"
import { MerchantProductDetailRoute } from "./routes/merchant-product-detail-route"
import { MerchantProductsRoute } from "./routes/merchant-products-route"
import { MerchantShipmentDetailRoute } from "./routes/merchant-shipment-detail-route"
import { MerchantShipmentsRoute } from "./routes/merchant-shipments-route"
import { OrderDetailRoute } from "./routes/order-detail-route"
import { OrdersRoute } from "./routes/orders-route"
import { ProductDetailRoute } from "./routes/product-detail-route"
import { ShipmentDetailRoute } from "./routes/shipment-detail-route"
import {
  SignInPage,
  SignUpPage,
  MerchantSignInPage,
  MerchantSignUpPage,
} from "@/features/auth"
import {
  MOCK_PAYMENT_ROUTE_PATHS,
  MockPaymentDetailPage,
  MockPaymentPage,
} from "@/mock/payment"
import {
  MOCK_LOGISTICS_ROUTE_PATHS,
  MockLogisticsDetailPage,
  MockLogisticsPage,
} from "@/mock/logistics"

export const router = createBrowserRouter([
  {
    element: <RootLayout />,
    children: [
      { path: "/", element: <HomePage /> },
      { path: "/signin", element: <SignInPage /> },
      { path: "/signup", element: <SignUpPage /> },
      { path: "/cart", element: <CartRoute /> },
      { path: "/checkout", element: <CheckoutRoute /> },
      { path: "/products/:productID", element: <ProductDetailRoute /> },
      { path: "/orders", element: <OrdersRoute /> },
      { path: "/orders/:orderID", element: <OrderDetailRoute /> },
      { path: "/shipments/:shipmentNo", element: <ShipmentDetailRoute /> },
      {
        path: MOCK_PAYMENT_ROUTE_PATHS.list,
        element: <MockPaymentPage />,
      },
      {
        path: MOCK_PAYMENT_ROUTE_PATHS.detail,
        element: <MockPaymentDetailPage />,
      },
      {
        path: MOCK_LOGISTICS_ROUTE_PATHS.list,
        element: <MockLogisticsPage />,
      },
      {
        path: MOCK_LOGISTICS_ROUTE_PATHS.detail,
        element: <MockLogisticsDetailPage />,
      },
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
        element: <MerchantHomePage />,
      },
      {
        element: <MerchantConsoleLayout />,
        children: [
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
            path: MERCHANT_CONSOLE_PATHS.orders,
            element: <MerchantOrdersRoute />,
            handle: {
              breadcrumbs: [
                {
                  to: MERCHANT_CONSOLE_PATHS.orders,
                  labelKey: "console:nav.orders",
                },
              ],
            } satisfies MerchantConsoleRouteHandle,
          },
          {
            path: MERCHANT_CONSOLE_PATHS.logistics,
            element: <MerchantShipmentsRoute />,
            handle: {
              breadcrumbs: [
                {
                  to: MERCHANT_CONSOLE_PATHS.logistics,
                  labelKey: "console:nav.logistics",
                },
              ],
            } satisfies MerchantConsoleRouteHandle,
          },
          {
            path: `${MERCHANT_CONSOLE_PATHS.logistics}/:shipmentNo`,
            element: <MerchantShipmentDetailRoute />,
            handle: {
              breadcrumbs: [
                {
                  to: MERCHANT_CONSOLE_PATHS.logistics,
                  labelKey: "console:nav.logistics",
                },
                {
                  to: MERCHANT_CONSOLE_PATHS.logistics,
                  labelKey: "console:nav.shipmentDetail",
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
