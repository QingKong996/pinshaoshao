import {
  Outlet,
  RouterProvider,
  ScrollRestoration,
  createBrowserRouter,
} from "react-router"
import { GridOverlay } from "@/components/grid-overlay"
import { AuthProvider, MerchantAuthProvider } from "@/lib/auth"
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
      { path: "*", element: <NotFoundPage /> },
    ],
  },
])

function App() {
  return (
    <AuthProvider>
      <MerchantAuthProvider>
        <RouterProvider router={router} />
      </MerchantAuthProvider>
    </AuthProvider>
  )
}

export default App
