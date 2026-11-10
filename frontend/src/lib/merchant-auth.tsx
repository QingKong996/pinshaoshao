import { useCallback, useMemo, useState, type ReactNode } from "react"
import {
  merchantLogin,
  merchantLogout,
  merchantRegister,
  type LoginRequest,
  type MerchantRegisterRequest,
} from "@/lib/api"
import {
  MerchantAuthContext,
  type MerchantAuthContextValue,
  type MerchantSession,
} from "@/lib/merchant-auth-context"

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

export function MerchantAuthProvider({ children }: { children: ReactNode }) {
  const [session, setSession] = useState<MerchantSession | null>(
    readStoredSession,
  )

  const signIn = useCallback(async (credentials: LoginRequest) => {
    const { token } = await merchantLogin(credentials)
    const next: MerchantSession = { token, username: credentials.username }
    writeStoredSession(next)
    setSession(next)
  }, [])

  const signUp = useCallback(async (details: MerchantRegisterRequest) => {
    const { token } = await merchantRegister(details)
    const next: MerchantSession = { token, username: details.username }
    writeStoredSession(next)
    setSession(next)
  }, [])

  const signOut = useCallback(async () => {
    const token = session?.token
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
  }, [session])

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
