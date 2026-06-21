import { useEffect, useRef, useState } from 'react'

interface VoteButtonProps {
  count: number
  hasVoted: boolean
  isAuthor: boolean
  onVote: () => void
}

export function VoteButton({ count, hasVoted, isAuthor, onVote }: VoteButtonProps) {
  const [pop, setPop] = useState(false)
  const prevVoted = useRef(hasVoted)

  // Pop the count whenever this card transitions into the voted state.
  useEffect(() => {
    if (hasVoted && !prevVoted.current) {
      setPop(true)
      const t = setTimeout(() => setPop(false), 320)
      return () => clearTimeout(t)
    }
    prevVoted.current = hasVoted
  }, [hasVoted])

  const base =
    'group/vote flex w-14 shrink-0 flex-col items-center gap-0.5 rounded-xl border px-2 py-2.5 transition-all duration-150'

  if (isAuthor) {
    return (
      <div
        className={`${base} cursor-default border-border bg-surface-2/40 text-faint`}
        title="You can't vote on your own request"
        aria-label={`${count} votes. This is your request.`}
      >
        <Triangle filled={false} />
        <span className="font-mono text-sm font-medium text-muted">{count}</span>
      </div>
    )
  }

  return (
    <button
      type="button"
      onClick={onVote}
      aria-pressed={hasVoted}
      aria-label={hasVoted ? `Voted. ${count} votes` : `Vote for this request. ${count} votes`}
      disabled={hasVoted}
      className={`${base} ${
        hasVoted
          ? 'cursor-default border-success/40 bg-success/15 text-success'
          : 'border-border bg-surface-2/40 text-muted hover:-translate-y-0.5 hover:border-accent/50 hover:text-accent'
      }`}
    >
      <Triangle filled={hasVoted} />
      <span className={`font-mono text-sm font-semibold ${pop ? 'animate-vote-pop' : ''} ${hasVoted ? 'text-success' : 'text-text'}`}>
        {count}
      </span>
    </button>
  )
}

function Triangle({ filled }: { filled: boolean }) {
  return (
    <svg viewBox="0 0 24 24" className="h-3.5 w-3.5" fill={filled ? 'currentColor' : 'none'} aria-hidden="true">
      <path d="M12 5l7 12H5z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round" />
    </svg>
  )
}
