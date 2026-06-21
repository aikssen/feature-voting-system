import { createContext, useCallback, useMemo, useState, type ReactNode } from 'react'
import { api } from '../api'
import type { UserView } from '../api/types'

const TOKEN_KEY = 'soundflow_token'
const USER_KEY = 'soundflow_user'

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
  } catch {
    return null
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(() => window.localStorage.getItem(TOKEN_KEY))
  const [user, setUser] = useState<UserView | null>(() => readUser())
  const [authOpen, setAuthOpen] = useState(false)
  const [authMode, setAuthMode] = useState<AuthMode>('login')

  const persist = useCallback((t: string, u: UserView) => {
    window.localStorage.setItem(TOKEN_KEY, t)
    window.localStorage.setItem(USER_KEY, JSON.stringify(u))
    setToken(t)
    setUser(u)
  }, [])

  const openAuth = useCallback((mode: AuthMode = 'login') => {
    setAuthMode(mode)
    setAuthOpen(true)
  }, [])

  const closeAuth = useCallback(() => setAuthOpen(false), [])

  const login = useCallback(
    async (email: string, password: string) => {
      const res = await api.login({ email, password })
      persist(res.token, res.user)
      setAuthOpen(false)
    },
    [persist],
  )

  const signup = useCallback(
    async (name: string, email: string, password: string) => {
      const res = await api.signup({ name, email, password })
      persist(res.token, res.user)
      setAuthOpen(false)
    },
    [persist],
  )

  const logout = useCallback(() => {
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
