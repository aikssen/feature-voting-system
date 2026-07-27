import { afterEach, describe, expect, it, vi } from 'vitest'
import { CORRELATION_ID_HEADER } from '../lib/logger'
import { ApiError, request } from './client'

const SERVER_CORRELATION_ID = 'server-cid-1'

// A real Response always carries headers; the fake must too, since the client
// reads the correlation id the backend echoes back off them.
function mockFetch(status: number, body: unknown) {
  return vi.fn().mockResolvedValue({
    ok: status >= 200 && status < 300,
    status,
    headers: new Headers({ [CORRELATION_ID_HEADER]: SERVER_CORRELATION_ID }),
    json: () => Promise.resolve(body),
  } as Response)
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('request', () => {
  it('returns the parsed payload on success', async () => {
    vi.stubGlobal('fetch', mockFetch(200, { token: 'abc' }))
    await expect(request('/auth/login', { method: 'POST', body: {} })).resolves.toEqual({ token: 'abc' })
  })

  it('maps the error envelope to ApiError with code + status', async () => {
    vi.stubGlobal('fetch', mockFetch(409, { error: { code: 'ALREADY_VOTED', message: 'Already voted.' } }))
    await expect(request('/features/1/vote', { method: 'POST' })).rejects.toMatchObject({
      code: 'ALREADY_VOTED',
      status: 409,
    })
  })

  it('carries validation details', async () => {
    vi.stubGlobal(
      'fetch',
      mockFetch(400, { error: { code: 'VALIDATION_ERROR', message: 'Invalid.', details: [{ field: 'title', issue: 'too short' }] } }),
    )
    try {
      await request('/features', { method: 'POST', body: {} })
      throw new Error('should have thrown')
    } catch (err) {
      expect(err).toBeInstanceOf(ApiError)
      expect((err as ApiError).details).toEqual([{ field: 'title', issue: 'too short' }])
    }
  })

  it('wraps connection failures as a NETWORK_ERROR ApiError', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')))
    await expect(request('/features')).rejects.toMatchObject({ code: 'NETWORK_ERROR', status: 0 })
  })

  it('sends a distinct correlation id on every request', async () => {
    const fetchMock = mockFetch(200, {})
    vi.stubGlobal('fetch', fetchMock)

    await request('/features')
    await request('/features')

    const sent = fetchMock.mock.calls.map((call) => (call[1].headers as Record<string, string>)[CORRELATION_ID_HEADER])
    expect(sent[0]).toBeTruthy()
    expect(sent[0]).not.toBe(sent[1])
  })

  it('surfaces the correlation id the server echoed back on failures', async () => {
    vi.stubGlobal('fetch', mockFetch(500, { error: { code: 'INTERNAL', message: 'Something went wrong.' } }))
    await expect(request('/features')).rejects.toMatchObject({ correlationId: SERVER_CORRELATION_ID })
  })

  it('falls back to the client-generated id when the server echoes none', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: false,
        status: 500,
        headers: new Headers(),
        json: () => Promise.resolve({ error: { code: 'INTERNAL', message: 'Something went wrong.' } }),
      } as Response),
    )
    await expect(request('/features')).rejects.toSatisfy(
      (err: ApiError) => err.correlationId.length > 0 && err.correlationId !== SERVER_CORRELATION_ID,
    )
  })
})
