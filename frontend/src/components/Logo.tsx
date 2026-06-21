/**
 * SoundFlow mark — an animated equalizer, the product's signature motif
 * (DESIGN.md: musical wave / sound wave). Pure CSS bars, no image asset.
 */
export function Logo({ withWordmark = true }: { withWordmark?: boolean }) {
  return (
    <span className="flex items-center gap-2.5">
      <span
        aria-hidden="true"
        className="flex h-7 w-7 items-end justify-center gap-[3px] rounded-lg bg-surface-2 p-1.5 ring-1 ring-border"
      >
        <span className="eq-bar h-2.5" />
        <span className="eq-bar h-4" />
        <span className="eq-bar h-3" />
        <span className="eq-bar h-4.5" />
      </span>
      {withWordmark && (
        <span className="text-[15px] font-extrabold tracking-tight text-text">
          Sound<span className="text-accent">Flow</span>
        </span>
      )}
    </span>
  )
}
