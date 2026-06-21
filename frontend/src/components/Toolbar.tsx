import type { Sort } from '../api/types'

const FILTERS: { value: Sort; label: string }[] = [
  { value: 'trending', label: 'Trending' },
  { value: 'most_voted', label: 'Most Voted' },
  { value: 'newest', label: 'Newest' },
]

interface ToolbarProps {
  search: string
  sort: Sort
  onSearch: (v: string) => void
  onSort: (s: Sort) => void
}

export function Toolbar({ search, sort, onSearch, onSort }: ToolbarProps) {
  return (
    <div className="flex flex-col gap-3">
      <div className="relative">
        <SearchIcon />
        <input
          type="search"
          value={search}
          onChange={(e) => onSearch(e.target.value)}
          placeholder="Search feature requests…"
          aria-label="Search feature requests"
          className="w-full rounded-xl border border-border bg-surface/60 py-3 pl-11 pr-4 text-sm text-text placeholder:text-faint transition-colors focus:border-accent focus:outline-none"
        />
      </div>

      {/* Single-select sort → radiogroup semantics (not tabs, which imply panels). */}
      <div role="radiogroup" aria-label="Sort feature requests" className="flex flex-wrap gap-1.5">
        {FILTERS.map((f) => {
          const active = sort === f.value
          return (
            <button
              key={f.value}
              type="button"
              role="radio"
              aria-checked={active}
              onClick={() => onSort(f.value)}
              className={`rounded-lg px-3 py-2 text-sm font-medium transition-colors ${
                active
                  ? 'bg-accent/15 text-accent ring-1 ring-accent/30'
                  : 'text-muted hover:bg-surface-2/60 hover:text-text'
              }`}
            >
              {f.label}
            </button>
          )
        })}
      </div>
    </div>
  )
}

function SearchIcon() {
  return (
    <svg
      aria-hidden="true"
      viewBox="0 0 24 24"
      fill="none"
      className="pointer-events-none absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-faint"
    >
      <circle cx="11" cy="11" r="7" stroke="currentColor" strokeWidth="2" />
      <path d="m20 20-3-3" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
    </svg>
  )
}
