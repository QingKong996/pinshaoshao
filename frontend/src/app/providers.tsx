import type { ReactNode } from "react"
import { QueryClientProvider } from "@tanstack/react-query"
import { UserAuthProvider, MerchantAuthProvider } from "@/features/auth"
import { queryClient } from "./query-client"

function clearSessionCache() {
  queryClient.clear()
}

export function AppProviders({ children }: { children: ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      <UserAuthProvider>
        <MerchantAuthProvider onBeforeSessionChange={clearSessionCache}>
          {children}
        </MerchantAuthProvider>
      </UserAuthProvider>
    </QueryClientProvider>
  )
}
