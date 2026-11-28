import type { ReactElement, ReactNode } from "react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { render } from "@testing-library/react"
import { MemoryRouter, type InitialEntry } from "react-router"

export function createTestQueryClient({
  gcTime = 0,
}: {
  gcTime?: number
} = {}) {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        gcTime,
      },
      mutations: {
        retry: false,
      },
    },
  })
}

export function renderWithProviders(
  ui: ReactElement,
  {
    queryClient = createTestQueryClient(),
    initialEntries = ["/"],
  }: {
    queryClient?: QueryClient
    initialEntries?: InitialEntry[]
  } = {},
) {
  function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={initialEntries}>{children}</MemoryRouter>
      </QueryClientProvider>
    )
  }

  return {
    queryClient,
    ...render(ui, { wrapper: Wrapper }),
  }
}
