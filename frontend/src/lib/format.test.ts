import { describe, expect, it, vi, afterEach } from 'vitest'
import { initialOf, relativeTime, voteLabel } from './format'

describe('initialOf', () => {
  it('returns the uppercased first letter', () => {
    expect(initialOf('ever')).toBe('E')
    expect(initialOf('  mia')).toBe('M')
  })
  it('falls back for empty input', () => {
    expect(initialOf('')).toBe('?')
  })
})

describe('voteLabel', () => {
  it('pluralizes correctly', () => {
    expect(voteLabel(1)).toBe('1 vote')
    expect(voteLabel(0)).toBe('0 votes')
    expect(voteLabel(124)).toBe('124 votes')
  })
})

describe('relativeTime', () => {
  afterEach(() => vi.useRealTimers())

  it('renders past timestamps relative to now', () => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-06-21T12:00:00Z'))
    expect(relativeTime('2026-06-19T12:00:00Z')).toBe('2 days ago')
    expect(relativeTime('2026-06-21T11:00:00Z')).toBe('1 hour ago')
  })

  it('returns "just now" for very recent times', () => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-06-21T12:00:00Z'))
    expect(relativeTime('2026-06-21T11:59:50Z')).toBe('just now')
  })

  it('returns empty string for invalid input', () => {
    expect(relativeTime('not-a-date')).toBe('')
  })
})
