import type { CSSProperties } from 'react'
import type { FeatureView } from '../api/types'
import { relativeTime } from '../lib/format'
import { VoteButton } from './VoteButton'

interface FeatureCardProps {
  feature: FeatureView
  onVote: (id: string) => void
  style?: CSSProperties
}

export function FeatureCard({ feature, onVote, style }: FeatureCardProps) {
  return (
    <article
      style={style}
      className="animate-rise group flex gap-3 rounded-2xl border border-border bg-surface/70 p-4 transition-colors duration-200 hover:border-border-strong hover:bg-surface-2/50"
    >
      <VoteButton
        count={feature.total_votes}
        hasVoted={feature.has_voted}
        isAuthor={feature.is_author}
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
