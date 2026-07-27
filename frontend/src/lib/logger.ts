/**
 * Client-side structured logging (DECISIONS.md D-LOG).
 *
 * Mirrors the backend's model so one browser action can be followed end to end:
 * every API call carries an `X-Correlation-ID`, the server stamps it on all the
 * lines it emits, and the same id appears in the console lines here.
 *
 * Verbosity comes from `VITE_LOG_LEVEL` (Compose mirrors the root `LOG_LEVEL`
 * into it, since Vite only exposes `VITE_`-prefixed vars to the browser).
 *
 * Never log tokens or passwords.
 */

export const LOG_LEVELS = ['debug', 'info', 'warn', 'error', 'silent'] as const
export type LogLevel = (typeof LOG_LEVELS)[number]

const SEVERITY: Record<LogLevel, number> = { debug: 10, info: 20, warn: 30, error: 40, silent: 100 }

/** Wire name of the trace id — must match the backend's constant. */
export const CORRELATION_ID_HEADER = 'X-Correlation-ID'

function resolveLevel(raw: string | undefined): LogLevel {
  const candidate = (raw ?? '').trim().toLowerCase()
  if ((LOG_LEVELS as readonly string[]).includes(candidate)) return candidate as LogLevel
  // 'warning' is the only alias the backend accepts; keep the two in step.
  if (candidate === 'warning') return 'warn'
  return 'info'
}

const activeLevel = resolveLevel(import.meta.env.VITE_LOG_LEVEL)
const threshold = SEVERITY[activeLevel]

/** The configured level, exposed for diagnostics and tests. */
export function currentLevel(): LogLevel {
  return activeLevel
}

/**
 * Generates a correlation id. `crypto.randomUUID` needs a secure context, which
 * plain-http LAN dev servers are not, so fall back to a random hex string —
 * uniqueness within one session is all a trace id needs.
 */
export function newCorrelationId(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }
  const bytes = new Uint8Array(16)
  if (typeof crypto !== 'undefined' && typeof crypto.getRandomValues === 'function') {
    crypto.getRandomValues(bytes)
  } else {
    for (let i = 0; i < bytes.length; i++) bytes[i] = Math.floor(Math.random() * 256)
  }
  return Array.from(bytes, (b) => b.toString(16).padStart(2, '0')).join('')
}

type Fields = Record<string, unknown>

function emit(level: Exclude<LogLevel, 'silent'>, scope: string, message: string, fields?: Fields) {
  if (SEVERITY[level] < threshold) return

  // console.debug is hidden behind the browser's "Verbose" filter by default,
  // which would silently swallow the level the user explicitly asked for.
  const sink = level === 'debug' ? console.log : console[level]
  const prefix = `%c[${scope}]`
  const style = level === 'error' ? 'color:#f87171' : level === 'warn' ? 'color:#fbbf24' : 'color:#22d3ee'

  if (fields && Object.keys(fields).length > 0) {
    sink(prefix, style, message, fields)
  } else {
    sink(prefix, style, message)
  }
}

export interface Logger {
  debug: (message: string, fields?: Fields) => void
  info: (message: string, fields?: Fields) => void
  warn: (message: string, fields?: Fields) => void
  error: (message: string, fields?: Fields) => void
  /** Returns a logger that stamps `fields` onto every line — used to pin a correlation id. */
  with: (fields: Fields) => Logger
}

function build(scope: string, bound: Fields): Logger {
  const merge = (fields?: Fields): Fields => ({ ...bound, ...fields })
  return {
    debug: (message, fields) => emit('debug', scope, message, merge(fields)),
    info: (message, fields) => emit('info', scope, message, merge(fields)),
    warn: (message, fields) => emit('warn', scope, message, merge(fields)),
    error: (message, fields) => emit('error', scope, message, merge(fields)),
    with: (fields) => build(scope, merge(fields)),
  }
}

/**
 * Creates a logger for one module. The scope is the first thing shown, so
 * filtering the console by `[api]` or `[vote]` isolates a subsystem.
 */
export function createLogger(scope: string): Logger {
  return build(scope, {})
}
