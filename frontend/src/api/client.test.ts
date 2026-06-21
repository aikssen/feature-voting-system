import { afterEach, describe, expect, it, vi } from 'vitest'
import { ApiError, request } from './client'

function mockFetch(status: number, body: unknown) {
  return vi.fn().mockResolvedValue({
    ok: status >= 200 && status < 300,
    status,
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
})
