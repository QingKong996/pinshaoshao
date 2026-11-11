import { useCallback, useEffect, useRef, useState } from "react"

export type RequestState<T> = {
  data: T | null
  error: unknown
  loading: boolean
  reload: () => void
}

type RequestResult<T> = Omit<RequestState<T>, "reload">

export function useRequest<T>(
  loader: () => Promise<T>,
  key: string,
): RequestState<T> {
  const [result, setResult] = useState<RequestResult<T>>({
    data: null,
    error: null,
    loading: true,
  })
  const [version, setVersion] = useState(0)
  const loaderRef = useRef(loader)

  useEffect(() => {
    loaderRef.current = loader
  })

  useEffect(() => {
    let active = true

    loaderRef
      .current()
      .then((data) => {
        if (active) {
          setResult({ data, error: null, loading: false })
        }
      })
      .catch((error: unknown) => {
        if (active) {
          setResult((current) => ({ ...current, error, loading: false }))
        }
      })

    return () => {
      active = false
    }
  }, [version, key])

  const reload = useCallback(() => {
    setResult((current) => ({ ...current, error: null, loading: true }))
    setVersion((current) => current + 1)
  }, [])

  return { ...result, reload }
}
