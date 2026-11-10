import {
  Outlet,
  RouterProvider,
  ScrollRestoration,
  createBrowserRouter,
} from "react-router"
import { GridOverlay } from "@/components/grid-overlay"
import { MerchantLayout } from "@/components/merchant/merchant-layout"
import { MerchantAuthProvider } from "@/lib/merchant-auth"
import { UserAuthProvider } from "@/lib/user-auth"
import HomePage from "@/pages/home"
import MerchantSignInPage from "@/pages/merchant-sign-in"
import MerchantSignUpPage from "@/pages/merchant-sign-up"
import CategoriesPage from "@/pages/merchant/categories"
import DashboardPage from "@/pages/merchant/dashboard"
import ProductDetailPage from "@/pages/merchant/product-detail"
import ProductsPage from "@/pages/merchant/products"
import ShopProfilePage from "@/pages/merchant/shop"
import NotFoundPage from "@/pages/not-found"
import SignInPage from "@/pages/sign-in"
import SignUpPage from "@/pages/sign-up"

function RootLayout() {
  return (
    <>
      <Outlet />
      <ScrollRestoration />
      {import.meta.env.DEV && <GridOverlay />}
    </>
  )
}

const router = createBrowserRouter([
  {
    element: <RootLayout />,
    children: [
      { path: "/", element: <HomePage /> },
      { path: "/signin", element: <SignInPage /> },
      { path: "/signup", element: <SignUpPage /> },
      {
        path: "/merchant",
        element: <MerchantLayout />,
        children: [
          { index: true, element: <DashboardPage /> },
          { path: "products", element: <ProductsPage /> },
          { path: "products/:productID", element: <ProductDetailPage /> },
          { path: "categories", element: <CategoriesPage /> },
          { path: "shop", element: <ShopProfilePage /> },
        ],
      },
      { path: "/merchant/signin", element: <MerchantSignInPage /> },
      { path: "/merchant/signup", element: <MerchantSignUpPage /> },
      { path: "*", element: <NotFoundPage /> },
    ],
  },
])

function App() {
  return (
    <UserAuthProvider>
      <MerchantAuthProvider>
        <RouterProvider router={router} />
      </MerchantAuthProvider>
    </UserAuthProvider>
  )
}

export default App
