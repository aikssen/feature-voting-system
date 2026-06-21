/** First character of a name, uppercased — used for the user badge. */
export function initialOf(name: string): string {
  return name.trim().charAt(0).toUpperCase() || '?'
}

const UNITS: [Intl.RelativeTimeFormatUnit, number][] = [
  ['year', 60 * 60 * 24 * 365],
  ['month', 60 * 60 * 24 * 30],
  ['week', 60 * 60 * 24 * 7],
  ['day', 60 * 60 * 24],
  ['hour', 60 * 60],
  ['minute', 60],
]

const rtf = new Intl.RelativeTimeFormat('en', { numeric: 'auto' })

/** Renders an ISO-8601 UTC timestamp as a browser-local relative time. */
export function relativeTime(iso: string): string {
  const then = new Date(iso).getTime()
  if (Number.isNaN(then)) return ''
  const seconds = Math.round((then - Date.now()) / 1000)
  const abs = Math.abs(seconds)
  if (abs < 45) return 'just now'
  for (const [unit, secs] of UNITS) {
    if (abs >= secs) {
      return rtf.format(Math.round(seconds / secs), unit)
    }
  }
  return 'just now'
}

/** "124 votes" / "1 vote". */
export function voteLabel(count: number): string {
  return `${count} ${count === 1 ? 'vote' : 'votes'}`
}
