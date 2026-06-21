import { request } from './client'
import type { AuthResponse, FeatureView, ListParams, Page, VoteResponse } from './types'

export const api = {
  signup(body: { name: string; email: string; password: string }) {
    return request<AuthResponse>('/auth/signup', { method: 'POST', body })
  },

  login(body: { email: string; password: string }) {
    return request<AuthResponse>('/auth/login', { method: 'POST', body })
  },

  listFeatures(params: ListParams, opts: { token?: string | null; signal?: AbortSignal } = {}) {
    const qs = new URLSearchParams()
    if (params.search) qs.set('search', params.search)
    if (params.sort) qs.set('sort', params.sort)
    if (params.page) qs.set('page', String(params.page))
    if (params.limit) qs.set('limit', String(params.limit))
    const suffix = qs.toString() ? `?${qs}` : ''
    return request<Page<FeatureView>>(`/features${suffix}`, { token: opts.token, signal: opts.signal })
  },

  getFeature(id: string, token?: string | null) {
    return request<FeatureView>(`/features/${id}`, { token })
  },

  createFeature(body: { title: string; description: string }, token: string) {
    return request<FeatureView>('/features', { method: 'POST', body, token })
  },

  vote(featureId: string, token: string) {
    return request<VoteResponse>(`/features/${featureId}/vote`, { method: 'POST', token })
  },
}

export { ApiError } from './client'
export type * from './types'
