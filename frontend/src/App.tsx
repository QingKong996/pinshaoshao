import {
  Navigate,
  Outlet,
  RouterProvider,
  ScrollRestoration,
  createBrowserRouter,
} from "react-router"
import { GridOverlay } from "@/components/grid-overlay"
import { MerchantConsoleLayout } from "@/components/merchant-console-layout"
import { MerchantAuthProvider } from "@/lib/merchant-auth"
import {
  MERCHANT_CONSOLE_BASE_PATH,
  MERCHANT_CONSOLE_HOME_PATH,
  MERCHANT_CONSOLE_SECTIONS,
} from "@/lib/merchant-console-nav"
import { UserAuthProvider } from "@/lib/user-auth"
import HomePage from "@/pages/home"
import MerchantSignInPage from "@/pages/merchant-sign-in"
import MerchantSignUpPage from "@/pages/merchant-sign-up"
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
      { path: "/merchant/signin", element: <MerchantSignInPage /> },
      { path: "/merchant/signup", element: <MerchantSignUpPage /> },
      {
        path: MERCHANT_CONSOLE_BASE_PATH,
        element: <MerchantConsoleLayout />,
        children: [
          {
            index: true,
            element: <Navigate to={MERCHANT_CONSOLE_HOME_PATH} replace />,
          },
          // Section pages are intentionally empty until their content lands.
          ...MERCHANT_CONSOLE_SECTIONS.map(({ slug }) => ({ path: slug })),
        ],
      },
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
