import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react"
import { Navigate, useLocation } from "react-router"
import {
  MOCK_SESSION_REJECTED_EVENT,
  createMockApiClient,
  readMockSession,
  writeMockSession,
  type MockSession,
} from "@/shared/api/mock-client"

type MockAuthContextValue = {
  session: MockSession | null
  signIn: (username: string, password: string) => Promise<void>
  signOut: () => Promise<void>
}

type MockAuthConfig = {
  apiBaseURL: string
  signInPath: string
  storageKey: string
}

export function createMockAuth(config: MockAuthConfig) {
  const client = createMockApiClient(config.apiBaseURL, config.storageKey)
  const Context = createContext<MockAuthContextValue | null>(null)

  function Provider({
    children,
    onSessionChange,
  }: {
    children: ReactNode
    onSessionChange?: () => void
  }) {
    const [session, setSession] = useState<MockSession | null>(() =>
      readMockSession(config.storageKey),
    )

    const updateSession = useCallback(
      (next: MockSession | null) => {
        onSessionChange?.()
        writeMockSession(config.storageKey, next)
        setSession(next)
      },
      [onSessionChange],
    )

    useEffect(() => {
      function handleSessionRejected(event: Event) {
        if (
          event instanceof CustomEvent &&
          event.detail === config.storageKey
        ) {
          updateSession(null)
        }
      }

      function handleStorage(event: StorageEvent) {
        try {
          if (
            event.storageArea !== sessionStorage ||
            event.key !== config.storageKey
          ) {
            return
          }
        } catch {
          return
        }
        setSession(readMockSession(config.storageKey))
      }

      window.addEventListener(
        MOCK_SESSION_REJECTED_EVENT,
        handleSessionRejected,
      )
      window.addEventListener("storage", handleStorage)
      return () => {
        window.removeEventListener(
          MOCK_SESSION_REJECTED_EVENT,
          handleSessionRejected,
        )
        window.removeEventListener("storage", handleStorage)
      }
    }, [updateSession])

    const signIn = useCallback(
      async (username: string, password: string) => {
        const { data } = await client.post<{ token: string }>("/auth/login", {
          username,
          password,
        })
        updateSession({
          token: data.token,
          username: username.trim(),
        })
      },
      [updateSession],
    )

    const signOut = useCallback(async () => {
      const token = session?.token
      updateSession(null)
      if (!token) {
        return
      }
      try {
        await client.post("/auth/logout", undefined, {
          headers: { Authorization: `Bearer ${token}` },
        })
      } catch {
        // The local session is already cleared.
      }
    }, [session, updateSession])

    const value = useMemo<MockAuthContextValue>(
      () => ({ session, signIn, signOut }),
      [session, signIn, signOut],
    )

    return <Context.Provider value={value}>{children}</Context.Provider>
  }

  function useMockAuth() {
    const value = useContext(Context)
    if (!value) {
      throw new Error("Mock auth provider is missing")
    }
    return value
  }

  function RequireMockAuth({ children }: { children: ReactNode }) {
    const { session } = useMockAuth()
    const location = useLocation()
    if (!session) {
      return (
        <Navigate
          to={config.signInPath}
          replace
          state={{ from: location.pathname + location.search }}
        />
      )
    }
    return children
  }

  return {
    Provider,
    RequireMockAuth,
    client,
    useMockAuth,
  }
}
