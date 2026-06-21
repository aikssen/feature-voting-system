// A rotating palette so each feature card gets its own accent colour — used for
// the card's left border, gradient wash, and hover glow (purely decorative).
const CARD_ACCENTS = [
  '#06b6d4', // cyan
  '#8b5cf6', // violet
  '#22c55e', // emerald
  '#f59e0b', // amber
  '#f43f5e', // rose
  '#3b82f6', // blue
  '#ec4899', // pink
  '#14b8a6', // teal
] as const

/** Deterministic accent for a card by its position in the list. */
export function accentFor(index: number): string {
  return CARD_ACCENTS[((index % CARD_ACCENTS.length) + CARD_ACCENTS.length) % CARD_ACCENTS.length]
}
