import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { axe } from 'jest-axe'
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

function jsonResponse(body: unknown): Response {
  return { ok: true, status: 200, json: () => Promise.resolve(body) } as Response
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

describe('accessibility (axe)', () => {
  beforeEach(() => {
    window.localStorage.clear()
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonResponse(page([feature(), feature({ id: 'f2', title: 'Sleep Timer', is_author: true })]))))
  })
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('the home page has no detectable a11y violations', async () => {
    const { container } = renderApp()
    await screen.findByText('Offline Downloads')
    expect(await axe(container)).toHaveNoViolations()
  })

  it('the auth dialog has no detectable a11y violations', async () => {
    const { container } = renderApp()
    await screen.findByText('Offline Downloads')
    await userEvent.click(screen.getByRole('button', { name: /^log in$/i }))
    await screen.findByRole('dialog')
    expect(await axe(container)).toHaveNoViolations()
  })

  it('the submission form has no detectable a11y violations', async () => {
    // Authenticated so the inline form is reachable.
    window.localStorage.setItem('soundflow_token', 'jwt')
    window.localStorage.setItem('soundflow_user', JSON.stringify({ id: 'u1', name: 'Ever', email: 'e@x.com', created_at: '' }))
    const { container } = renderApp()
    await screen.findByText('Offline Downloads')
    await userEvent.click(screen.getByRole('button', { name: /submit feature request/i }))
    await screen.findByLabelText('Title')
    expect(await axe(container)).toHaveNoViolations()
  })
})
