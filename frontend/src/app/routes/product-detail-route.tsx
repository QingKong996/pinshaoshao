import { useParams } from "react-router"
import { useAuth } from "@/features/auth"
import { useAddCartItem } from "@/features/cart"
import { ProductDetailPage, ProductSignedOut } from "@/features/feed"
import { SiteFooter } from "../layouts/site-footer"
import { SiteHeader } from "../layouts/site-header"

export function ProductDetailRoute() {
  const { session } = useAuth()
  const { productID } = useParams()
  const parsedProductID = Number(productID)

  return (
    <div className="flex min-h-svh flex-col">
      <SiteHeader />
      <main className="mx-auto w-full max-w-6xl flex-1 px-4 py-8 sm:px-6">
        {session ? (
          <SignedInProductDetail
            token={session.token}
            productID={parsedProductID}
          />
        ) : (
          <ProductSignedOut />
        )}
      </main>
      <SiteFooter />
    </div>
  )
}

function SignedInProductDetail({
  token,
  productID,
}: {
  token: string
  productID: number
}) {
  const addToCart = useAddCartItem(token)

  return (
    <ProductDetailPage
      token={token}
      productID={productID}
      onAddToCart={async (skuID, quantity) => {
        await addToCart.mutateAsync({ sku_id: skuID, quantity })
      }}
    />
  )
}
