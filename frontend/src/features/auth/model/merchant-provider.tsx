import { useCallback, useMemo, useState, type ReactNode } from "react"
import {
  merchantLogin,
  merchantLogout,
  merchantRegister,
} from "../api/merchant"
import { MerchantAuthContext } from "./merchant-context"
import type {
  LoginRequest,
  MerchantRegisterRequest,
  MerchantAuthContextValue,
  MerchantSession,
} from "../types"

const MERCHANT_SESSION_STORAGE_KEY = "pinshaoshao.merchant-session"

function readStoredSession(): MerchantSession | null {
  try {
    const raw = localStorage.getItem(MERCHANT_SESSION_STORAGE_KEY)
    if (!raw) {
      return null
    }
    const parsed = JSON.parse(raw) as Partial<MerchantSession>
    if (
      typeof parsed.token !== "string" ||
      !parsed.token ||
      typeof parsed.username !== "string" ||
      !parsed.username
    ) {
      return null
    }
    return { token: parsed.token, username: parsed.username }
  } catch {
    return null
  }
}

function writeStoredSession(session: MerchantSession | null) {
  if (session) {
    localStorage.setItem(
      MERCHANT_SESSION_STORAGE_KEY,
      JSON.stringify(session),
    )
    return
  }
  localStorage.removeItem(MERCHANT_SESSION_STORAGE_KEY)
}

export function MerchantAuthProvider({
  children,
  onBeforeSessionChange,
}: {
  children: ReactNode
  onBeforeSessionChange: () => void
}) {
  const [session, setSession] = useState<MerchantSession | null>(
    readStoredSession,
  )

  const signIn = useCallback(async (credentials: LoginRequest) => {
    const { token } = await merchantLogin(credentials)
    const next: MerchantSession = { token, username: credentials.username }
    onBeforeSessionChange()
    writeStoredSession(next)
    setSession(next)
  }, [onBeforeSessionChange])

  const signUp = useCallback(async (details: MerchantRegisterRequest) => {
    const { token } = await merchantRegister(details)
    const next: MerchantSession = { token, username: details.username }
    onBeforeSessionChange()
    writeStoredSession(next)
    setSession(next)
  }, [onBeforeSessionChange])

  const signOut = useCallback(async () => {
    const token = session?.token
    onBeforeSessionChange()
    writeStoredSession(null)
    setSession(null)
    if (!token) {
      return
    }
    try {
      await merchantLogout(token)
    } catch {
      // the local session is already cleared
    }
  }, [session, onBeforeSessionChange])

  const value = useMemo<MerchantAuthContextValue>(
    () => ({ session, signIn, signUp, signOut }),
    [session, signIn, signUp, signOut],
  )

  return (
    <MerchantAuthContext.Provider value={value}>
      {children}
    </MerchantAuthContext.Provider>
  )
}
