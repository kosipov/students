import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState, type ReactNode } from 'react'

const TOAST_DURATION = 2200

type ShowToast = (message: string) => void

const ToastContext = createContext<ShowToast | null>(null)

export function ToastProvider({ children }: { children: ReactNode }) {
  const [message, setMessage] = useState('')
  // Bumped on every toast, so the same message shown twice restarts the animation.
  const [key, setKey] = useState(0)
  const timer = useRef<number | undefined>(undefined)

  const show = useCallback<ShowToast>((text) => {
    setMessage(text)
    setKey((k) => k + 1)
    window.clearTimeout(timer.current)
    timer.current = window.setTimeout(() => setMessage(''), TOAST_DURATION)
  }, [])

  useEffect(() => () => window.clearTimeout(timer.current), [])

  const value = useMemo(() => show, [show])

  return (
    <ToastContext.Provider value={value}>
      {children}
      <div role="status" aria-live="polite">
        {message && (
          <div key={key} className="toast">
            {message}
          </div>
        )}
      </div>
    </ToastContext.Provider>
  )
}

export function useToast(): ShowToast {
  const show = useContext(ToastContext)
  if (!show) {
    throw new Error('useToast must be used inside ToastProvider')
  }
  return show
}
