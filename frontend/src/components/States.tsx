import { Button } from './ui'

/** Skeleton list shown while the first page loads. */
export function FeatureSkeleton() {
  return (
    <div className="flex flex-col gap-3" aria-hidden="true">
      {Array.from({ length: 5 }).map((_, i) => (
        <div key={i} className="flex gap-3 rounded-2xl border border-border bg-surface/40 p-4">
          <div className="h-14 w-14 shrink-0 animate-pulse rounded-xl bg-surface-2" />
          <div className="flex-1 space-y-2.5 py-1">
            <div className="h-4 w-1/2 animate-pulse rounded bg-surface-2" />
            <div className="h-3 w-3/4 animate-pulse rounded bg-surface-2" />
            <div className="h-3 w-1/4 animate-pulse rounded bg-surface-2" />
          </div>
        </div>
      ))}
    </div>
  )
}

export function EmptyState({ searching, onClear }: { searching: boolean; onClear: () => void }) {
  return (
    <div className="flex flex-col items-center gap-3 rounded-2xl border border-dashed border-border bg-surface/30 px-6 py-16 text-center">
      <div aria-hidden="true" className="flex h-12 items-end gap-1.5">
        <span className="eq-bar h-6" />
        <span className="eq-bar h-10" />
        <span className="eq-bar h-4" />
        <span className="eq-bar h-8" />
      </div>
      {searching ? (
        <>
          <p className="text-text">No requests match your search.</p>
          <Button variant="ghost" onClick={onClear}>
            Clear search
          </Button>
        </>
      ) : (
        <>
          <p className="text-text">No feature requests yet.</p>
          <p className="text-sm text-muted">Be the first to share an idea for SoundFlow.</p>
        </>
      )}
    </div>
  )
}

export function ErrorState({ message, onRetry }: { message: string; onRetry: () => void }) {
  return (
    <div
      role="alert"
      className="flex flex-col items-center gap-3 rounded-2xl border border-danger/30 bg-danger/5 px-6 py-14 text-center"
    >
      <p className="text-text">{message}</p>
      <Button variant="ghost" onClick={onRetry}>
        Try again
      </Button>
    </div>
  )
}
