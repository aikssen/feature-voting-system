import { useCallback, useEffect, useRef, useState } from 'react'
import { api, ApiError } from '../api'
import type { FeatureView, Page, Sort } from '../api/types'
import { useAuth } from '../auth/useAuth'
import { useToast } from '../components/toast/useToast'
import { createLogger } from '../lib/logger'
import { useDebounced } from '../lib/useDebounced'

export const PAGE_LIMIT = 12

const log = createLogger('features')

interface UseFeatures {
  data: Page<FeatureView> | null
  loading: boolean
  error: ApiError | null
  search: string
  sort: Sort
  page: number
  setSearch: (v: string) => void
  setSort: (s: Sort) => void
  setPage: (p: number) => void
  refresh: () => void
  vote: (id: string) => Promise<void>
}

function patchFeature(page: Page<FeatureView> | null, id: string, fn: (f: FeatureView) => FeatureView): Page<FeatureView> | null {
  if (!page) return page
  return { ...page, items: page.items.map((f) => (f.id === id ? fn(f) : f)) }
}

export function useFeatures(): UseFeatures {
  const { token, isAuthenticated, openAuth, logout } = useAuth()
  const toast = useToast()

  const [search, setSearchRaw] = useState('')
  const [sort, setSortRaw] = useState<Sort>('trending')
  const [page, setPage] = useState(1)
  const [data, setData] = useState<Page<FeatureView> | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<ApiError | null>(null)
  const [reloadKey, setReloadKey] = useState(0)

  const debouncedSearch = useDebounced(search, 300)
  const tokenRef = useRef(token)
  tokenRef.current = token

  // Changing the query or ordering returns to the first page (DECISIONS.md R5:
  // page-based pagination keeps dynamic-sort reshuffle visually bounded).
  const setSearch = useCallback((v: string) => {
    setSearchRaw(v)
    setPage(1)
  }, [])
  const setSort = useCallback((s: Sort) => {
    setSortRaw(s)
    setPage(1)
  }, [])
  const refresh = useCallback(() => setReloadKey((k) => k + 1), [])

  useEffect(() => {
    const controller = new AbortController()
    setLoading(true)
    setError(null)
    log.debug('loading feature list', { search: debouncedSearch, sort, page, limit: PAGE_LIMIT })

    api
      .listFeatures({ search: debouncedSearch, sort, page, limit: PAGE_LIMIT }, { token: tokenRef.current, signal: controller.signal })
      .then((res) => {
        log.info('feature list loaded', { returned: res.items.length, total: res.total, page: res.page, sort })
        setData(res)
        setLoading(false)
      })
      .catch((err) => {
        if (err instanceof DOMException && err.name === 'AbortError') return
        log.error('feature list failed', {
          sort,
          page,
          code: err instanceof ApiError ? err.code : 'UNKNOWN',
          correlation_id: err instanceof ApiError ? err.correlationId : undefined,
        })
        setError(err instanceof ApiError ? err : new ApiError(0, 'NETWORK_ERROR', 'Cannot reach the server.'))
        setLoading(false)
      })

    return () => controller.abort()
  }, [debouncedSearch, sort, page, token, reloadKey])

  const vote = useCallback(
    async (id: string) => {
      if (!isAuthenticated || !token) {
        log.debug('vote blocked: anonymous, prompting login', { feature_id: id })
        openAuth('login')
        return
      }

      const snapshot = data?.items.find((f) => f.id === id)
      if (!snapshot || snapshot.has_voted || snapshot.is_author) {
        log.debug('vote skipped', {
          feature_id: id,
          reason: !snapshot ? 'not_in_current_page' : snapshot.is_author ? 'is_author' : 'already_voted',
        })
        return
      }

      // Optimistic update (DESIGN.md — voting should feel immediate).
      log.debug('vote applied optimistically', { feature_id: id, from_votes: snapshot.total_votes })
      setData((prev) => patchFeature(prev, id, (f) => ({ ...f, has_voted: true, total_votes: f.total_votes + 1 })))

      try {
        const res = await api.vote(id, token)
        log.info('vote confirmed', { feature_id: id, total_votes: res.total_votes })
        setData((prev) => patchFeature(prev, id, (f) => ({ ...f, has_voted: res.has_voted, total_votes: res.total_votes })))
        toast.success('Vote counted')
      } catch (err) {
        // Roll back to the captured snapshot.
        const code = err instanceof ApiError ? err.code : 'UNKNOWN'
        const correlationId = err instanceof ApiError ? err.correlationId : undefined
        log.warn('vote rejected, rolling back optimistic update', { feature_id: id, code, correlation_id: correlationId })
        setData((prev) => patchFeature(prev, id, () => snapshot))
        if (err instanceof ApiError && err.status === 401) {
          log.info('session expired during vote, logging out', { feature_id: id })
          logout()
          openAuth('login')
          toast.error('Your session expired. Please log in again.')
        } else if (err instanceof ApiError) {
          toast.error(err.message)
        } else {
          log.error('vote failed with an unexpected error', { feature_id: id, error: String(err) })
          toast.error('Could not record your vote. Please try again.')
        }
      }
    },
    [data, isAuthenticated, token, openAuth, logout, toast],
  )

  return { data, loading, error, search, sort, page, setSearch, setSort, setPage, refresh, vote }
}
