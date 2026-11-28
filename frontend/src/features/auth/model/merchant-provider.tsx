import { useCallback, useMemo, useRef, useState, type ReactNode } from "react"
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
  try {
    if (session) {
      localStorage.setItem(
        MERCHANT_SESSION_STORAGE_KEY,
        JSON.stringify(session),
      )
      return
    }
    localStorage.removeItem(MERCHANT_SESSION_STORAGE_KEY)
  } catch {
    // Keep the in-memory session when persistent storage is unavailable.
  }
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
  const currentSession = useRef(session)

  const replaceSession = useCallback((next: MerchantSession | null) => {
    currentSession.current = next
    onBeforeSessionChange()
    writeStoredSession(next)
    setSession(next)
  }, [onBeforeSessionChange])

  const signIn = useCallback(async (credentials: LoginRequest) => {
    const { token } = await merchantLogin(credentials)
    replaceSession({ token, username: credentials.username })
  }, [replaceSession])

  const signUp = useCallback(async (details: MerchantRegisterRequest) => {
    const { token } = await merchantRegister(details)
    replaceSession({ token, username: details.username })
  }, [replaceSession])

  const signOut = useCallback(async () => {
    const token = currentSession.current?.token
    replaceSession(null)
    if (!token) {
      return
    }
    try {
      await merchantLogout(token)
    } catch {
      // the local session is already cleared
    }
  }, [replaceSession])

  const expireSession = useCallback((token: string) => {
    if (currentSession.current?.token !== token) {
      return false
    }
    replaceSession(null)
    return true
  }, [replaceSession])

  const value = useMemo<MerchantAuthContextValue>(
    () => ({ session, signIn, signUp, signOut, expireSession }),
    [session, signIn, signUp, signOut, expireSession],
  )

  return (
    <MerchantAuthContext.Provider value={value}>
      {children}
    </MerchantAuthContext.Provider>
  )
}
