import { describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { VoteButton } from './VoteButton'

describe('VoteButton', () => {
  it('votes on click when allowed', async () => {
    const onVote = vi.fn()
    render(<VoteButton count={3} hasVoted={false} isAuthor={false} onVote={onVote} />)

    const btn = screen.getByRole('button', { name: /vote for this request/i })
    expect(btn).toHaveAttribute('aria-pressed', 'false')

    await userEvent.click(btn)
    expect(onVote).toHaveBeenCalledOnce()
  })

  it('is disabled and pressed once voted (no unvote)', async () => {
    const onVote = vi.fn()
    render(<VoteButton count={4} hasVoted isAuthor={false} onVote={onVote} />)

    const btn = screen.getByRole('button', { name: /voted/i })
    expect(btn).toBeDisabled()
    expect(btn).toHaveAttribute('aria-pressed', 'true')

    await userEvent.click(btn)
    expect(onVote).not.toHaveBeenCalled()
  })

  it('shows no vote control for the author', () => {
    render(<VoteButton count={2} hasVoted={false} isAuthor onVote={vi.fn()} />)
    expect(screen.queryByRole('button')).toBeNull()
    expect(screen.getByText('2')).toBeInTheDocument()
  })
})
