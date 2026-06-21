# Quality & Testing — Phase 5

This document records the quality validation for the SoundFlow Feature Voting
System: automated test coverage plus the accessibility, responsive, and mobile
usability review.

## Automated tests

| Suite | Scope | Command |
|-------|-------|---------|
| Backend unit | Auth / feature / vote services + business rules (mock repos, no DB) | `cd backend && go test ./...` |
| Backend HTTP E2E | Full contract against a live API + Postgres (29 checks) | `./scripts/smoke.sh` |
| Frontend unit/component | format, API client + error envelope, VoteButton, App integration | `cd frontend && pnpm test` |
| Frontend accessibility | axe assertions on home, auth dialog, submission form | `cd frontend && pnpm test` |

Business rules explicitly covered (per `ARCHITECTURE.md` testing priorities):
duplicate-vote rejection, self-vote rejection, auth required for vote/create,
required-field validation, and normalized duplicate-title rejection.

## Accessibility validation

Automated with **axe-core** (`jest-axe`) over the rendered app — home page, the
auth dialog, and the inline submission form all pass with **no violations**.
Issues found during the audit and fixed:

- **Heading order** — the page jumped `h1` (hero) → `h3` (card titles). Added an
  `h2` for the feature-requests region so levels increase by one.
- **Sort control semantics** — the filters used `role="tab"` (which implies tab
  panels). Changed to a `radiogroup` of `role="radio"` buttons with `aria-checked`,
  which correctly models single-select sorting.
- **Colour contrast** — the faint caption colour (`#71717a`) sat just under WCAG
  AA (~4.2:1) on the near-black background. Raised to `#8b8b95` (≥ 4.5:1 on both
  the page and card surfaces).

Verified by review (not flagged by axe, but checked):

- Single `h1`; landmarks present (`banner` header, `main`, `dialog`, live region
  for toasts). A **skip link** jumps keyboard users to the feature list.
- All interactive controls have accessible names (vote button `aria-label` +
  `aria-pressed`, icon-only search labelled, logo link labelled). Decorative SVGs
  and the equalizer mark are `aria-hidden`.
- Form fields use real `<label>`s; errors use `aria-describedby` + `aria-invalid`.
- The auth dialog is `role="dialog"` `aria-modal`, traps focus, closes on Escape,
  locks scroll, and restores focus to the trigger.
- Visible focus styles (`:focus-visible`) on a dark theme; `prefers-reduced-motion`
  disables animations. `lang="en"` and `color-scheme: dark` are set.

## Responsive layout & mobile usability

Mobile-first by construction (base styles target mobile, `sm:` adapts upward),
per `DESIGN.md`. Reviewed and confirmed:

- **Viewport** — `width=device-width, initial-scale=1, viewport-fit=cover`.
- **No fixed-width overflow** — audited for hard-coded pixel widths; none that
  break small screens. Long titles/descriptions wrap (`min-w-0` flex content).
- **No hover-only actions** — voting and every action work on tap; hover is purely
  decorative enhancement.
- **Layout adapts** — hero type scales (`text-4xl` → `sm:text-6xl`); the auth/
  submission surfaces render as a bottom sheet on mobile and a centred dialog on
  larger screens; the header collapses the username on narrow screens.
- **Touch targets** — vote control and primary buttons are comfortably tappable;
  the sort control grew to `py-2` and now wraps on very narrow screens; long
  usernames truncate instead of pushing the header.

## How to run everything

```bash
# Backend
cd backend && go test ./...

# Frontend (unit + component + a11y)
cd frontend && pnpm test
cd frontend && pnpm typecheck && pnpm build

# Full-stack HTTP E2E (needs the stack running, e.g. docker compose up)
./scripts/smoke.sh
```

## Notes / limitations

- axe runs in jsdom, which has no layout engine, so the colour-contrast rule
  can't execute there — contrast was therefore verified manually against WCAG AA.
- Visual responsive behaviour across real devices is a manual check; the review
  above is structural. With the stack running (`docker compose up`), the app can
  be exercised at http://localhost:5173 (log in with `ever@example.com` / `Pa$s`).
