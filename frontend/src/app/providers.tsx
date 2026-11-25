import type { ReactNode } from "react"
import { QueryClientProvider } from "@tanstack/react-query"
import { UserAuthProvider, MerchantAuthProvider } from "@/features/auth"
import { mockLogisticsAuth } from "@/mock/logistics"
import { mockPaymentAuth } from "@/mock/payment"
import { queryClient } from "./query-client"

const { Provider: MockPaymentAuthProvider } = mockPaymentAuth
const { Provider: MockLogisticsAuthProvider } = mockLogisticsAuth

function clearSessionCache() {
  queryClient.clear()
}

export function AppProviders({ children }: { children: ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      <UserAuthProvider>
        <MerchantAuthProvider onBeforeSessionChange={clearSessionCache}>
          <MockPaymentAuthProvider onSessionChange={clearSessionCache}>
            <MockLogisticsAuthProvider onSessionChange={clearSessionCache}>
              {children}
            </MockLogisticsAuthProvider>
          </MockPaymentAuthProvider>
        </MerchantAuthProvider>
      </UserAuthProvider>
    </QueryClientProvider>
  )
}
