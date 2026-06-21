import type { CSSProperties } from 'react'
import type { FeatureView } from '../api/types'
import { relativeTime } from '../lib/format'
import { accentFor } from '../lib/cardAccents'
import { VoteButton } from './VoteButton'

interface FeatureCardProps {
  feature: FeatureView
  index: number
  onVote: (id: string) => void
}

export function FeatureCard({ feature, index, onVote }: FeatureCardProps) {
  const accent = accentFor(index)
  const style = {
    '--card-accent': accent,
    animationDelay: `${Math.min(index, 8) * 35}ms`,
  } as CSSProperties

  return (
    <article style={style} className="feature-card animate-rise group flex h-full gap-3 p-4">
      <VoteButton
        count={feature.total_votes}
        hasVoted={feature.has_voted}
        isAuthor={feature.is_author}
        accent={accent}
        onVote={() => onVote(feature.id)}
      />

      <div className="min-w-0 flex-1">
        <h3 className="text-pretty font-semibold leading-snug text-text">{feature.title}</h3>
        <p className="mt-1 text-sm leading-relaxed text-muted">{feature.description}</p>
        <div className="mt-3 flex items-center gap-1.5 text-xs text-faint">
          <span className="font-medium text-muted">{feature.author.name}</span>
          <span aria-hidden="true">·</span>
          <time dateTime={feature.created_at}>{relativeTime(feature.created_at)}</time>
        </div>
      </div>
    </article>
  )
}
