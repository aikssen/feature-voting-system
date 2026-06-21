import type { ApiErrorEnvelope, FieldError } from './types'

const BASE_URL = (import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:3000/api/v1').replace(/\/$/, '')

/** ApiError carries the machine code clients branch on (never the message). */
export class ApiError extends Error {
  readonly code: string
  readonly status: number
  readonly details: FieldError[]

  constructor(status: number, code: string, message: string, details: FieldError[] = []) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.details = details
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
  const headers: Record<string, string> = {}
  if (opts.body !== undefined) headers['Content-Type'] = 'application/json'
  if (opts.token) headers['Authorization'] = `Bearer ${opts.token}`

  let res: Response
  try {
    res = await fetch(`${BASE_URL}${path}`, {
      method: opts.method ?? 'GET',
      headers,
      body: opts.body !== undefined ? JSON.stringify(opts.body) : undefined,
      signal: opts.signal,
    })
  } catch (err) {
    if (err instanceof DOMException && err.name === 'AbortError') throw err
    throw new ApiError(0, NETWORK_ERROR_CODE, 'Cannot reach the server. Check your connection and try again.')
  }

  if (res.status === 204) return undefined as T

  const payload = await res.json().catch(() => null)

  if (!res.ok) {
    const env = payload as ApiErrorEnvelope | null
    if (env?.error) {
      throw new ApiError(res.status, env.error.code, env.error.message, env.error.details ?? [])
    }
    throw new ApiError(res.status, 'INTERNAL', 'Something went wrong. Please try again.')
  }

  return payload as T
}
