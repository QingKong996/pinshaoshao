import { useCallback, useMemo, useRef, useState, type ReactNode } from "react"
import {
  login as userLogin,
  logout as userLogout,
  register as userRegister,
} from "../api/user"
import { UserAuthContext } from "./user-context"
import type {
  LoginRequest,
  RegisterRequest,
  UserAuthContextValue,
  UserSession,
} from "../types"

const USER_SESSION_STORAGE_KEY = "pinshaoshao.session"

function readStoredSession(): UserSession | null {
  try {
    const raw = localStorage.getItem(USER_SESSION_STORAGE_KEY)
    if (!raw) {
      return null
    }
    const parsed = JSON.parse(raw) as Partial<UserSession>
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

function writeStoredSession(session: UserSession | null) {
  try {
    if (session) {
      localStorage.setItem(USER_SESSION_STORAGE_KEY, JSON.stringify(session))
      return
    }
    localStorage.removeItem(USER_SESSION_STORAGE_KEY)
  } catch {
    // Keep the in-memory session when persistent storage is unavailable.
  }
}

export function UserAuthProvider({
  children,
  onBeforeSessionChange,
}: {
  children: ReactNode
  onBeforeSessionChange: () => void
}) {
  const [session, setSession] = useState<UserSession | null>(readStoredSession)
  const currentSession = useRef(session)

  const replaceSession = useCallback((next: UserSession | null) => {
    currentSession.current = next
    onBeforeSessionChange()
    writeStoredSession(next)
    setSession(next)
  }, [onBeforeSessionChange])

  const signIn = useCallback(async (credentials: LoginRequest) => {
    const { token } = await userLogin(credentials)
    replaceSession({ token, username: credentials.username })
  }, [replaceSession])

  const signUp = useCallback(async (details: RegisterRequest) => {
    const { token } = await userRegister(details)
    replaceSession({ token, username: details.username })
  }, [replaceSession])

  const signOut = useCallback(async () => {
    const token = currentSession.current?.token
    replaceSession(null)
    if (!token) {
      return
    }
    try {
      await userLogout(token)
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

  const value = useMemo<UserAuthContextValue>(
    () => ({ session, signIn, signUp, signOut, expireSession }),
    [session, signIn, signUp, signOut, expireSession],
  )

  return (
    <UserAuthContext.Provider value={value}>
      {children}
    </UserAuthContext.Provider>
  )
}
