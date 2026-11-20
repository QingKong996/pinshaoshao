import { useAuth } from "@/features/auth"
import { CartPage, CartSignedOut } from "@/features/cart"
import { SiteFooter } from "../layouts/site-footer"
import { SiteHeader } from "../layouts/site-header"

export function CartRoute() {
  const { session } = useAuth()

  return (
    <div className="flex min-h-svh flex-col">
      <SiteHeader />
      <main className="mx-auto w-full max-w-6xl flex-1 px-4 py-8 sm:px-6">
        {session ? (
          <CartPage token={session.token} />
        ) : (
          <CartSignedOut />
        )}
      </main>
      <SiteFooter />
    </div>
  )
}
