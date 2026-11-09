import { useCallback, useMemo, useState, type ReactNode } from "react"
import {
  login as userLogin,
  logout as userLogout,
  merchantLogin,
  merchantLogout,
  merchantRegister,
  register as userRegister,
  type LoginRequest,
  type MerchantRegisterRequest,
  type RegisterRequest,
} from "@/lib/api"
import {
  AuthContext,
  MerchantAuthContext,
  type AuthContextValue,
  type AuthSession,
} from "@/lib/auth-context"

const USER_SESSION_STORAGE_KEY = "pinshaoshao.session"
const MERCHANT_SESSION_STORAGE_KEY = "pinshaoshao.merchant-session"

type SessionAuthClient<TRegister extends RegisterRequest> = {
  login: (credentials: LoginRequest) => Promise<{ token: string }>
  register: (details: TRegister) => Promise<{ token: string }>
  logout: (token: string) => Promise<void>
}

const USER_AUTH_CLIENT: SessionAuthClient<RegisterRequest> = {
  login: userLogin,
  register: userRegister,
  logout: userLogout,
}

const MERCHANT_AUTH_CLIENT: SessionAuthClient<MerchantRegisterRequest> = {
  login: merchantLogin,
  register: merchantRegister,
  logout: merchantLogout,
}

function readSession(storageKey: string): AuthSession | null {
  try {
    const raw = localStorage.getItem(storageKey)
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

function writeSession(storageKey: string, session: AuthSession | null) {
  if (session) {
    localStorage.setItem(storageKey, JSON.stringify(session))
    return
  }
  localStorage.removeItem(storageKey)
}

function useSessionAuth<TRegister extends RegisterRequest>(
  storageKey: string,
  client: SessionAuthClient<TRegister>,
): AuthContextValue<TRegister> {
  const [session, setSession] = useState<AuthSession | null>(() =>
    readSession(storageKey),
  )

  const signIn = useCallback(
    async (credentials: LoginRequest) => {
      const { token } = await client.login(credentials)
      const next: AuthSession = { token, username: credentials.username }
      writeSession(storageKey, next)
      setSession(next)
    },
    [client, storageKey],
  )

  const signUp = useCallback(
    async (details: TRegister) => {
      const { token } = await client.register(details)
      const next: AuthSession = { token, username: details.username }
      writeSession(storageKey, next)
      setSession(next)
    },
    [client, storageKey],
  )

  const signOut = useCallback(async () => {
    const token = session?.token
    writeSession(storageKey, null)
    setSession(null)
    if (!token) {
      return
    }
    try {
      await client.logout(token)
    } catch {
      // the local session is already cleared
    }
  }, [client, session, storageKey])

  return useMemo(
    () => ({ session, signIn, signUp, signOut }),
    [session, signIn, signUp, signOut],
  )
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const value = useSessionAuth(USER_SESSION_STORAGE_KEY, USER_AUTH_CLIENT)

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function MerchantAuthProvider({ children }: { children: ReactNode }) {
  const value = useSessionAuth(MERCHANT_SESSION_STORAGE_KEY, MERCHANT_AUTH_CLIENT)

  return (
    <MerchantAuthContext.Provider value={value}>
      {children}
    </MerchantAuthContext.Provider>
  )
}
