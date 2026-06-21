import '@testing-library/jest-dom/vitest'
import { afterEach, expect } from 'vitest'
import { cleanup } from '@testing-library/react'
import { toHaveNoViolations } from 'jest-axe'

expect.extend(toHaveNoViolations)

// jsdom under this Node build does not expose a Web Storage implementation,
// so provide a minimal in-memory one for tests.
class MemoryStorage implements Storage {
  private store = new Map<string, string>()
  get length() {
    return this.store.size
  }
  clear() {
    this.store.clear()
  }
  getItem(key: string) {
    return this.store.has(key) ? this.store.get(key)! : null
  }
  setItem(key: string, value: string) {
    this.store.set(key, String(value))
  }
  removeItem(key: string) {
    this.store.delete(key)
  }
  key(index: number) {
    return Array.from(this.store.keys())[index] ?? null
  }
}

if (!('localStorage' in globalThis) || globalThis.localStorage == null) {
  const storage = new MemoryStorage()
  Object.defineProperty(globalThis, 'localStorage', { value: storage, writable: true, configurable: true })
  if (typeof window !== 'undefined') {
    Object.defineProperty(window, 'localStorage', { value: storage, writable: true, configurable: true })
  }
}

// jsdom doesn't implement scrollIntoView; stub it so handlers that call it
// (e.g. the hero CTA scrolling to the board) don't throw during tests.
if (!Element.prototype.scrollIntoView) {
  Element.prototype.scrollIntoView = () => {}
}

afterEach(() => {
  cleanup()
  window.localStorage.clear()
})
