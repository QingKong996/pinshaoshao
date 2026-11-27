import { useEffect, useRef } from "react"
import { Outlet, useNavigate } from "react-router"
import { useAuth } from "@/features/auth"
import { onUnauthorized } from "@/shared/api/client"

export function BuyerSessionLayout() {
  const { session, signOut } = useAuth()
  const navigate = useNavigate()
  const handlingUnauthorized = useRef(false)

  useEffect(() => {
    if (!session) {
      handlingUnauthorized.current = false
      return
    }
    return onUnauthorized(() => {
      if (handlingUnauthorized.current) {
        return
      }
      handlingUnauthorized.current = true
      void signOut()
      navigate("/signin", { replace: true })
    })
  }, [navigate, session, signOut])

  return <Outlet />
}
