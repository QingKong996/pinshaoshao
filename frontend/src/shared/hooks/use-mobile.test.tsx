import { act, renderHook } from "@testing-library/react"
import { describe, expect, it, vi } from "vitest"
import { useIsMobile } from "@/shared/hooks/use-mobile"

describe("useIsMobile", () => {
  it("tracks the viewport breakpoint", () => {
    const listeners = new Set<() => void>()
    let matches = false
    vi.mocked(window.matchMedia).mockReturnValue({
      get matches() {
        return matches
      },
      media: "(max-width: 767px)",
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: (
        _type: string,
        listener: EventListenerOrEventListenerObject,
      ) => {
        listeners.add(listener as () => void)
      },
      removeEventListener: (
        _type: string,
        listener: EventListenerOrEventListenerObject,
      ) => {
        listeners.delete(listener as () => void)
      },
      dispatchEvent: vi.fn(),
    })
    Object.defineProperty(window, "innerWidth", {
      configurable: true,
      writable: true,
      value: 1024,
    })

    const { result } = renderHook(() => useIsMobile())
    expect(result.current).toBe(false)

    act(() => {
      matches = true
      window.innerWidth = 500
      listeners.forEach((listener) => listener())
    })

    expect(result.current).toBe(true)
  })
})
