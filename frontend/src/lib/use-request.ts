import { useCallback, useEffect, useRef, useState, type DependencyList } from "react"

export type RequestState<T> = {
  data: T | null
  error: unknown
  loading: boolean
  reload: () => void
}

export function useRequest<T>(
  loader: () => Promise<T>,
  deps: DependencyList,
): RequestState<T> {
  const [data, setData] = useState<T | null>(null)
  const [error, setError] = useState<unknown>(null)
  const [loading, setLoading] = useState(true)
  const [version, setVersion] = useState(0)
  const loaderRef = useRef(loader)

  useEffect(() => {
    loaderRef.current = loader
    let active = true

    setLoading(true)
    setError(null)
    loaderRef
      .current()
      .then((result) => {
        if (!active) {
          return
        }
        setData(result)
        setError(null)
      })
      .catch((cause: unknown) => {
        if (active) {
          setError(cause)
        }
      })
      .finally(() => {
        if (active) {
          setLoading(false)
        }
      })

    return () => {
      active = false
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [version, ...deps])

  const reload = useCallback(() => {
    setVersion((current) => current + 1)
  }, [])

  return { data, error, loading, reload }
}
