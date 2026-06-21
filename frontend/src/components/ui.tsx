import { forwardRef, type ButtonHTMLAttributes, type InputHTMLAttributes, type TextareaHTMLAttributes } from 'react'

type Variant = 'primary' | 'success' | 'ghost' | 'subtle'

const VARIANTS: Record<Variant, string> = {
  primary:
    'bg-accent text-bg font-semibold shadow-[0_0_0_1px_rgba(6,182,212,0.5),0_8px_30px_-12px_rgba(6,182,212,0.6)] hover:brightness-110 active:brightness-95',
  success:
    'bg-success text-bg font-semibold hover:brightness-110 active:brightness-95',
  ghost:
    'border border-border bg-surface-2/60 text-text hover:border-border-strong hover:bg-surface-2',
  subtle: 'text-muted hover:text-text',
}

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant
  loading?: boolean
}

export function Button({ variant = 'primary', loading, className = '', children, disabled, ...rest }: ButtonProps) {
  return (
    <button
      {...rest}
      disabled={disabled || loading}
      className={`inline-flex items-center justify-center gap-2 rounded-xl px-4 py-2.5 text-sm transition-all duration-150 disabled:cursor-not-allowed disabled:opacity-60 ${VARIANTS[variant]} ${className}`}
    >
      {loading && <Spinner className="h-3.5 w-3.5" />}
      {children}
    </button>
  )
}

interface FieldProps {
  label: string
  error?: string
  hint?: string
  id: string
}

export const TextField = forwardRef<HTMLInputElement, FieldProps & InputHTMLAttributes<HTMLInputElement>>(
  function TextField({ label, error, hint, id, className = '', ...rest }, ref) {
    return (
      <div className="flex flex-col gap-1.5">
        <label htmlFor={id} className="text-xs font-medium text-muted">
          {label}
        </label>
        <input
          ref={ref}
          id={id}
          aria-invalid={error ? true : undefined}
          aria-describedby={error ? `${id}-error` : hint ? `${id}-hint` : undefined}
          className={`w-full rounded-xl border bg-surface-2/60 px-3.5 py-2.5 text-sm text-text placeholder:text-faint transition-colors focus:border-accent focus:outline-none ${
            error ? 'border-danger' : 'border-border'
          } ${className}`}
          {...rest}
        />
        {error ? (
          <p id={`${id}-error`} className="text-xs text-danger">
            {error}
          </p>
        ) : hint ? (
          <p id={`${id}-hint`} className="text-xs text-faint">
            {hint}
          </p>
        ) : null}
      </div>
    )
  },
)

export const TextArea = forwardRef<HTMLTextAreaElement, FieldProps & TextareaHTMLAttributes<HTMLTextAreaElement>>(
  function TextArea({ label, error, hint, id, className = '', ...rest }, ref) {
    return (
      <div className="flex flex-col gap-1.5">
        <label htmlFor={id} className="text-xs font-medium text-muted">
          {label}
        </label>
        <textarea
          ref={ref}
          id={id}
          aria-invalid={error ? true : undefined}
          aria-describedby={error ? `${id}-error` : hint ? `${id}-hint` : undefined}
          className={`w-full resize-none rounded-xl border bg-surface-2/60 px-3.5 py-2.5 text-sm text-text placeholder:text-faint transition-colors focus:border-accent focus:outline-none ${
            error ? 'border-danger' : 'border-border'
          } ${className}`}
          {...rest}
        />
        {error ? (
          <p id={`${id}-error`} className="text-xs text-danger">
            {error}
          </p>
        ) : hint ? (
          <p id={`${id}-hint`} className="text-xs text-faint">
            {hint}
          </p>
        ) : null}
      </div>
    )
  },
)

export function Spinner({ className = '' }: { className?: string }) {
  return (
    <svg className={`animate-spin ${className}`} viewBox="0 0 24 24" fill="none" aria-hidden="true">
      <circle className="opacity-20" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="3" />
      <path className="opacity-90" d="M22 12a10 10 0 0 1-10 10" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
    </svg>
  )
}
