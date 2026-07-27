import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import App from './App'
import { AuthProvider } from './auth/AuthContext'
import { ToastProvider } from './components/toast/ToastContext'
import type { FeatureView, Page } from './api/types'

function feature(overrides: Partial<FeatureView> = {}): FeatureView {
  return {
    id: 'f1',
    title: 'Offline Downloads',
    description: 'Download playlists for travel.',
    author: { id: 'u2', name: 'Mia' },
    created_at: '2026-06-20T12:00:00Z',
    total_votes: 3,
    trending_score: 1.2,
    has_voted: false,
    is_author: false,
    rank: 1,
    ...overrides,
  }
}

function page(items: FeatureView[]): Page<FeatureView> {
  return { items, page: 1, limit: 12, total: items.length, total_pages: 1, has_next: false }
}

function jsonResponse(status: number, body: unknown): Response {
  // headers must exist — the api client reads the echoed correlation id from it.
  return { ok: status >= 200 && status < 300, status, headers: new Headers(), json: () => Promise.resolve(body) } as Response
}

function renderApp() {
  return render(
    <ToastProvider>
      <AuthProvider>
        <App />
      </AuthProvider>
    </ToastProvider>,
  )
}

describe('App', () => {
  beforeEach(() => {
    window.localStorage.clear()
  })
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('renders the feature list from the API', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonResponse(200, page([feature()]))))
    renderApp()

    expect(await screen.findByText('Offline Downloads')).toBeInTheDocument()
    expect(screen.getByText('Mia')).toBeInTheDocument()
  })

  it('optimistically increments the vote count when an authenticated user votes', async () => {
    // Seed an authenticated session.
    window.localStorage.setItem('soundflow_token', 'jwt')
    window.localStorage.setItem('soundflow_user', JSON.stringify({ id: 'u1', name: 'Ever', email: 'e@x.com', created_at: '' }))

    const fetchMock = vi.fn((url: string, init?: RequestInit) => {
      if (init?.method === 'POST' && url.includes('/vote')) {
        return Promise.resolve(jsonResponse(201, { feature_id: 'f1', total_votes: 4, has_voted: true }))
      }
      return Promise.resolve(jsonResponse(200, page([feature()])))
    })
    vi.stubGlobal('fetch', fetchMock)

    renderApp()

    const voteBtn = await screen.findByRole('button', { name: /vote for this request/i })
    expect(voteBtn).toHaveTextContent('3')

    await userEvent.click(voteBtn)

    // Optimistic + confirmed state: count is 4 and the control is now "Voted".
    await waitFor(() => expect(screen.getByRole('button', { name: /voted/i })).toHaveTextContent('4'))
    expect(fetchMock).toHaveBeenCalledWith(expect.stringContaining('/features/f1/vote'), expect.objectContaining({ method: 'POST' }))
  })

  it('prompts login when an anonymous user tries to vote', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonResponse(200, page([feature()]))))
    renderApp()

    const voteBtn = await screen.findByRole('button', { name: /vote for this request/i })
    await userEvent.click(voteBtn)

    // The auth dialog opens instead of casting a vote.
    expect(await screen.findByRole('dialog')).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: /welcome back/i })).toBeInTheDocument()
  })
})
