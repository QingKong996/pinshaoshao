import { useCallback, useMemo, useState, type ReactNode } from "react"
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
  if (session) {
    localStorage.setItem(USER_SESSION_STORAGE_KEY, JSON.stringify(session))
    return
  }
  localStorage.removeItem(USER_SESSION_STORAGE_KEY)
}

export function UserAuthProvider({ children }: { children: ReactNode }) {
  const [session, setSession] = useState<UserSession | null>(readStoredSession)

  const signIn = useCallback(async (credentials: LoginRequest) => {
    const { token } = await userLogin(credentials)
    const next: UserSession = { token, username: credentials.username }
    writeStoredSession(next)
    setSession(next)
  }, [])

  const signUp = useCallback(async (details: RegisterRequest) => {
    const { token } = await userRegister(details)
    const next: UserSession = { token, username: details.username }
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
      await userLogout(token)
    } catch {
      // the local session is already cleared
    }
  }, [session])

  const value = useMemo<UserAuthContextValue>(
    () => ({ session, signIn, signUp, signOut }),
    [session, signIn, signUp, signOut],
  )

  return (
    <UserAuthContext.Provider value={value}>
      {children}
    </UserAuthContext.Provider>
  )
}
