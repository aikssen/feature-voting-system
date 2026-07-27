import { CORRELATION_ID_HEADER, createLogger, newCorrelationId } from '../lib/logger'
import type { ApiErrorEnvelope, FieldError } from './types'

const BASE_URL = (import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:3000/api/v1').replace(/\/$/, '')

const log = createLogger('api')

/** ApiError carries the machine code clients branch on (never the message). */
export class ApiError extends Error {
  readonly code: string
  readonly status: number
  readonly details: FieldError[]
  /** Trace id of the failed request — quote it to find the server-side lines. */
  readonly correlationId: string

  constructor(status: number, code: string, message: string, details: FieldError[] = [], correlationId = '') {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.details = details
    this.correlationId = correlationId
  }
}

/** A network/parse failure that never reached a structured API response. */
export const NETWORK_ERROR_CODE = 'NETWORK_ERROR'

interface RequestOptions {
  method?: string
  body?: unknown
  token?: string | null
  signal?: AbortSignal
}

export async function request<T>(path: string, opts: RequestOptions = {}): Promise<T> {
  const method = opts.method ?? 'GET'

  // One id per request, minted here and sent to the backend, which stamps it on
  // every line it logs. The same id labels the console lines below, so a single
  // search spans browser and server.
  const correlationId = newCorrelationId()
  const reqLog = log.with({ correlation_id: correlationId, method, path })

  const headers: Record<string, string> = { [CORRELATION_ID_HEADER]: correlationId }
  if (opts.body !== undefined) headers['Content-Type'] = 'application/json'
  // Presence of the token is worth logging; the token itself never is.
  if (opts.token) headers['Authorization'] = `Bearer ${opts.token}`

  reqLog.debug('request started', { authenticated: Boolean(opts.token) })
  const startedAt = performance.now()
  const elapsed = () => Math.round(performance.now() - startedAt)

  let res: Response
  try {
    res = await fetch(`${BASE_URL}${path}`, {
      method,
      headers,
      body: opts.body !== undefined ? JSON.stringify(opts.body) : undefined,
      signal: opts.signal,
    })
  } catch (err) {
    if (err instanceof DOMException && err.name === 'AbortError') {
      reqLog.debug('request aborted', { duration_ms: elapsed() })
      throw err
    }
    reqLog.error('request failed: server unreachable', { duration_ms: elapsed(), error: String(err) })
    throw new ApiError(
      0,
      NETWORK_ERROR_CODE,
      'Cannot reach the server. Check your connection and try again.',
      [],
      correlationId,
    )
  }

  // The server echoes the id back (and mints one if ours was unusable), so trust
  // its value over ours when they differ.
  const serverCorrelationId = res.headers.get(CORRELATION_ID_HEADER) || correlationId
  const resLog = log.with({ correlation_id: serverCorrelationId, method, path })

  if (res.status === 204) {
    resLog.debug('request completed', { status: 204, duration_ms: elapsed() })
    return undefined as T
  }

  const payload = await res.json().catch(() => null)

  if (!res.ok) {
    const env = payload as ApiErrorEnvelope | null
    const code = env?.error?.code ?? 'INTERNAL'
    // 5xx is ours to fix; 4xx is usually the user being told something valid.
    const level = res.status >= 500 ? 'error' : 'warn'
    resLog[level]('request rejected', { status: res.status, code, duration_ms: elapsed() })

    if (env?.error) {
      throw new ApiError(res.status, env.error.code, env.error.message, env.error.details ?? [], serverCorrelationId)
    }
    throw new ApiError(res.status, 'INTERNAL', 'Something went wrong. Please try again.', [], serverCorrelationId)
  }

  resLog.debug('request completed', { status: res.status, duration_ms: elapsed() })
  return payload as T
}
