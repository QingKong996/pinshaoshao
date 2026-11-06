import { useCallback, useMemo, useState, type ReactNode } from "react"
import {
  login as loginRequest,
  logout as logoutRequest,
  register as registerRequest,
} from "@/lib/api"
import { AuthContext, type AuthSession } from "@/lib/auth-context"

const SESSION_STORAGE_KEY = "pinshaoshao.session"

function readSession(): AuthSession | null {
  try {
    const raw = localStorage.getItem(SESSION_STORAGE_KEY)
    if (!raw) {
      return null
    }
    const parsed = JSON.parse(raw) as Partial<AuthSession>
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

function writeSession(session: AuthSession | null) {
  if (session) {
    localStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify(session))
    return
  }
  localStorage.removeItem(SESSION_STORAGE_KEY)
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [session, setSession] = useState<AuthSession | null>(readSession)

  const signIn = useCallback(async (credentials: Parameters<typeof loginRequest>[0]) => {
    const { token } = await loginRequest(credentials)
    const next: AuthSession = { token, username: credentials.username }
    writeSession(next)
    setSession(next)
  }, [])

  const signUp = useCallback(async (details: Parameters<typeof registerRequest>[0]) => {
    const { token } = await registerRequest(details)
    const next: AuthSession = { token, username: details.username }
    writeSession(next)
    setSession(next)
  }, [])

  const signOut = useCallback(async () => {
    const token = session?.token
    writeSession(null)
    setSession(null)
    if (!token) {
      return
    }
    try {
      await logoutRequest(token)
    } catch {
      // the local session is already cleared
    }
  }, [session])

  const value = useMemo(
    () => ({ session, signIn, signUp, signOut }),
    [session, signIn, signUp, signOut],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
