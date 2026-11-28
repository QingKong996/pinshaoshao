import { useEffect } from "react"
import { Outlet, useNavigate } from "react-router"
import { useAuth } from "@/features/auth"
import { onUnauthorized } from "@/shared/api/client"

export function BuyerSessionLayout() {
  const { expireSession } = useAuth()
  const navigate = useNavigate()

  useEffect(() => {
    return onUnauthorized((token) => {
      if (expireSession(token)) {
        navigate("/signin", { replace: true })
      }
    })
  }, [expireSession, navigate])

  return <Outlet />
}
