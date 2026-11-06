import {
  Outlet,
  RouterProvider,
  ScrollRestoration,
  createBrowserRouter,
} from "react-router"
import { GridOverlay } from "@/components/grid-overlay"
import { AuthProvider } from "@/lib/auth"
import HomePage from "@/pages/home"
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
      { path: "*", element: <NotFoundPage /> },
    ],
  },
])

function App() {
  return (
    <AuthProvider>
      <RouterProvider router={router} />
    </AuthProvider>
  )
}

export default App
