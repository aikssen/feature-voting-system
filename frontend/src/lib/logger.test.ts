import { describe, expect, it, vi } from 'vitest'
import { CORRELATION_ID_HEADER, createLogger, currentLevel, newCorrelationId } from './logger'

describe('logger', () => {
  it('defaults to info when VITE_LOG_LEVEL is unset in the test env', () => {
    expect(currentLevel()).toBe('info')
  })

  it('suppresses lines below the active level', () => {
    const spy = vi.spyOn(console, 'log').mockImplementation(() => {})
    createLogger('test').debug('should not appear')
    expect(spy).not.toHaveBeenCalled()
    spy.mockRestore()
  })

  it('emits lines at or above the active level with the scope', () => {
    const spy = vi.spyOn(console, 'info').mockImplementation(() => {})
    createLogger('vote').info('vote confirmed', { feature_id: 'abc' })

    expect(spy).toHaveBeenCalledTimes(1)
    const [prefix, , message, fields] = spy.mock.calls[0]
    expect(prefix).toContain('[vote]')
    expect(message).toBe('vote confirmed')
    expect(fields).toEqual({ feature_id: 'abc' })
    spy.mockRestore()
  })

  it('stamps bound fields onto every line from a derived logger', () => {
    const spy = vi.spyOn(console, 'warn').mockImplementation(() => {})
    createLogger('api').with({ correlation_id: 'cid-1' }).warn('request rejected', { status: 409 })

    expect(spy.mock.calls[0][3]).toEqual({ correlation_id: 'cid-1', status: 409 })
    spy.mockRestore()
  })

  it('mints a distinct correlation id per call', () => {
    const ids = new Set(Array.from({ length: 50 }, () => newCorrelationId()))
    expect(ids.size).toBe(50)
    for (const id of ids) {
      // Must survive the backend's NormalizeCorrelationID character allow-list,
      // otherwise the server silently replaces it and the trace breaks.
      expect(id).toMatch(/^[A-Za-z0-9_.:-]{1,128}$/)
    }
  })

  it('agrees with the backend on the header name', () => {
    expect(CORRELATION_ID_HEADER).toBe('X-Correlation-ID')
  })
})
