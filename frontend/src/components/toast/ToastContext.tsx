import { createContext, useCallback, useRef, useState, type ReactNode } from 'react'

export type ToastKind = 'success' | 'error'

export interface Toast {
  id: number
  kind: ToastKind
  message: string
}

export interface ToastContextValue {
  notify: (kind: ToastKind, message: string) => void
  success: (message: string) => void
  error: (message: string) => void
}

export const ToastContext = createContext<ToastContextValue | null>(null)

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([])
  const idRef = useRef(0)

  const dismiss = useCallback((id: number) => {
    setToasts((prev) => prev.filter((t) => t.id !== id))
  }, [])

  const notify = useCallback(
    (kind: ToastKind, message: string) => {
      const id = ++idRef.current
      setToasts((prev) => [...prev, { id, kind, message }])
      setTimeout(() => dismiss(id), 3800)
    },
    [dismiss],
  )

  const success = useCallback((m: string) => notify('success', m), [notify])
  const error = useCallback((m: string) => notify('error', m), [notify])

  return (
    <ToastContext.Provider value={{ notify, success, error }}>
      {children}
      <div
        className="pointer-events-none fixed inset-x-0 bottom-0 z-[60] flex flex-col items-center gap-2 p-4"
        aria-live="polite"
        aria-atomic="false"
      >
        {toasts.map((t) => (
          <div
            key={t.id}
            role={t.kind === 'error' ? 'alert' : 'status'}
            className={`animate-rise glass pointer-events-auto flex max-w-sm items-center gap-2.5 rounded-xl px-4 py-3 text-sm shadow-xl ${
              t.kind === 'success' ? 'border-success/40' : 'border-danger/40'
            }`}
          >
            <span
              aria-hidden="true"
              className={`h-2 w-2 shrink-0 rounded-full ${t.kind === 'success' ? 'bg-success' : 'bg-danger'}`}
            />
            <span className="text-text">{t.message}</span>
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  )
}
