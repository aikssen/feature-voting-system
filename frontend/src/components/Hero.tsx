import { Button } from './ui'

export function Hero({ onSubmit }: { onSubmit: () => void }) {
  return (
    <section className="relative overflow-hidden">
      {/* Neon stage glow + faint vinyl grooves — DESIGN.md hero imagery, all CSS. */}
      <div aria-hidden="true" className="pointer-events-none absolute inset-0 -z-10">
        <div className="absolute left-1/2 top-[-6rem] h-72 w-72 -translate-x-1/2 rounded-full bg-accent/20 blur-[90px]" />
        <div className="absolute right-8 top-10 h-48 w-48 rounded-full bg-success/15 blur-[80px]" />
        <div
          className="absolute inset-0 opacity-[0.06]"
          style={{
            backgroundImage:
              'repeating-radial-gradient(circle at 50% 30%, #fff 0 1px, transparent 1px 14px)',
          }}
        />
      </div>

      <div className="mx-auto max-w-5xl px-4 pb-10 pt-16 text-center sm:px-6 sm:pt-24">
        <span className="inline-flex items-center gap-2 rounded-full border border-border bg-surface/60 px-3 py-1 text-xs font-medium text-muted">
          <span className="h-1.5 w-1.5 rounded-full bg-success" />
          Community feature voting
        </span>

        <h1 className="mx-auto mt-6 max-w-2xl text-balance text-4xl font-extrabold leading-[1.05] tracking-tight sm:text-6xl">
          Shape the Future of <span className="text-accent">SoundFlow</span>
        </h1>

        <p className="mx-auto mt-5 max-w-xl text-pretty text-base text-muted sm:text-lg">
          Help us decide which features we build next. Submit ideas, discover what the community
          wants, and vote on what matters most.
        </p>

        <div className="mt-8 flex justify-center">
          <Button onClick={onSubmit} className="px-5 py-3 text-base">
            Submit Feature Request
          </Button>
        </div>
      </div>
    </section>
  )
}
