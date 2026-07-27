import { createContext, useCallback, useMemo, useState, type ReactNode } from 'react'
import { api } from '../api'
import type { UserView } from '../api/types'
import { createLogger } from '../lib/logger'

const TOKEN_KEY = 'soundflow_token'
const USER_KEY = 'soundflow_user'

const log = createLogger('auth')

export type AuthMode = 'login' | 'signup'

export interface AuthContextValue {
  user: UserView | null
  token: string | null
  isAuthenticated: boolean
  /** Modal state */
  authOpen: boolean
  authMode: AuthMode
  openAuth: (mode?: AuthMode) => void
  closeAuth: () => void
  /** Actions — throw ApiError on failure so callers can surface field issues. */
  login: (email: string, password: string) => Promise<void>
  signup: (name: string, email: string, password: string) => Promise<void>
  logout: () => void
}

export const AuthContext = createContext<AuthContextValue | null>(null)

function readUser(): UserView | null {
  try {
    const raw = window.localStorage.getItem(USER_KEY)
    return raw ? (JSON.parse(raw) as UserView) : null
  } catch (err) {
    // Corrupt storage silently logging the user out is exactly the kind of
    // invisible failure this project had no way to trace.
    log.warn('stored user is unreadable, treating session as anonymous', { error: String(err) })
    return null
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(() => window.localStorage.getItem(TOKEN_KEY))
  const [user, setUser] = useState<UserView | null>(() => readUser())
  const [authOpen, setAuthOpen] = useState(false)
  const [authMode, setAuthMode] = useState<AuthMode>('login')

  const persist = useCallback((t: string, u: UserView) => {
    // t is the JWT — stored, never logged.
    window.localStorage.setItem(TOKEN_KEY, t)
    window.localStorage.setItem(USER_KEY, JSON.stringify(u))
    setToken(t)
    setUser(u)
  }, [])

  const openAuth = useCallback((mode: AuthMode = 'login') => {
    log.debug('auth modal opened', { mode })
    setAuthMode(mode)
    setAuthOpen(true)
  }, [])

  const closeAuth = useCallback(() => setAuthOpen(false), [])

  const login = useCallback(
    async (email: string, password: string) => {
      log.debug('login submitted', { email })
      try {
        const res = await api.login({ email, password })
        log.info('login succeeded', { user_id: res.user.id })
        persist(res.token, res.user)
        setAuthOpen(false)
      } catch (err) {
        // The api client already logged status + correlation id; this line adds
        // the user-facing action that triggered it.
        log.warn('login failed', { email })
        throw err
      }
    },
    [persist],
  )

  const signup = useCallback(
    async (name: string, email: string, password: string) => {
      log.debug('signup submitted', { email })
      try {
        const res = await api.signup({ name, email, password })
        log.info('signup succeeded', { user_id: res.user.id })
        persist(res.token, res.user)
        setAuthOpen(false)
      } catch (err) {
        log.warn('signup failed', { email })
        throw err
      }
    },
    [persist],
  )

  const logout = useCallback(() => {
    log.info('logout')
    window.localStorage.removeItem(TOKEN_KEY)
    window.localStorage.removeItem(USER_KEY)
    setToken(null)
    setUser(null)
  }, [])

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      token,
      isAuthenticated: Boolean(token),
      authOpen,
      authMode,
      openAuth,
      closeAuth,
      login,
      signup,
      logout,
    }),
    [user, token, authOpen, authMode, openAuth, closeAuth, login, signup, logout],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
