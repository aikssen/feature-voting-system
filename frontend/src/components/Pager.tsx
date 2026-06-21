import type { ReactNode } from 'react'

interface PagerProps {
  page: number
  totalPages: number
  onPage: (p: number) => void
}

export function Pager({ page, totalPages, onPage }: PagerProps) {
  if (totalPages <= 1) return null

  return (
    <nav className="flex items-center justify-center gap-4 pt-2" aria-label="Pagination">
      <PagerButton disabled={page <= 1} onClick={() => onPage(page - 1)}>
        ← Prev
      </PagerButton>
      <span className="font-mono text-xs text-muted" aria-live="polite">
        Page {page} of {totalPages}
      </span>
      <PagerButton disabled={page >= totalPages} onClick={() => onPage(page + 1)}>
        Next →
      </PagerButton>
    </nav>
  )
}

function PagerButton({ disabled, onClick, children }: { disabled: boolean; onClick: () => void; children: ReactNode }) {
  return (
    <button
      onClick={onClick}
      disabled={disabled}
      className="rounded-lg border border-border bg-surface-2/50 px-3 py-1.5 text-sm text-text transition-colors hover:border-border-strong disabled:cursor-not-allowed disabled:opacity-40"
    >
      {children}
    </button>
  )
}
