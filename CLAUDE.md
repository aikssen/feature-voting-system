# CLAUDE.md — How AI Agents Work in This Project

This file defines **operating rules for AI agents** working in this repo. It is not
a requirements doc — for *what* to build, read the documents below. This file is
about *how* to work. When in doubt, follow this file; for product/technical detail,
follow the linked docs.

## Documentation map & source of truth

Read these before acting. Do not restate their contents elsewhere — link to them.

| Doc | Owns |
|-----|------|
| `PROJECT.md` | Product scope, user journey, business rules, non-goals |
| `DESIGN.md` | UX, visual identity, layout, frontend behavior |
| `ARCHITECTURE.md` | System design, stack, structure, API surface, data model |
| `DECISIONS.md` | Resolved questions + the **frozen** API / error / data contract |
| `TASKS.md` | Implementation plan, phases, progress |

**Conflict resolution order:** `DECISIONS.md` → `PROJECT.md` / `DESIGN.md` /
`ARCHITECTURE.md` → `TASKS.md`. If `DECISIONS.md` contradicts an older doc, the
contract in `DECISIONS.md` wins. If you find a contradiction the contract does not
cover, **stop and surface it** rather than guessing.

## Prompt logging (required)

Every time you receive a new instruction or prompt, append it to `prompts.txt` in
the project root with an ISO-8601 timestamp and a brief summary of what you did in
response. Create the file if it doesn't exist. Keep this log updated across all
sessions. Do this even for planning/analysis-only turns.

## Working norms

- **Plan before building.** For non-trivial work, state the approach and the files
  you'll touch first. Prefer a short plan over a large speculative diff.
- **Build to the frozen contract.** Endpoints, error envelope, status codes,
  validation, and data model are defined in `DECISIONS.md`. Match them exactly. If
  the contract needs to change, change `DECISIONS.md` first (with rationale), then
  the code — never let code and contract silently diverge.
- **Keep docs consistent.** When a decision changes scope or contract, update the
  owning doc in the same turn. Don't leave two docs disagreeing.
- **Respect scope.** Honor the non-goals in `PROJECT.md` and `ARCHITECTURE.md`. Do
  not add abstractions, endpoints, or dependencies that aren't required. Flag
  overengineering rather than introducing it.
- **Confirm before irreversible or scope-expanding actions** — new dependencies,
  schema changes beyond the contract, deleting files you didn't create, anything
  outward-facing. Approval for one step is not approval for the next.

## Guardrails

- **Don't break the "no edit/delete" rules.** Per `PROJECT.md`, feature requests
  and votes are create-only; unvote/edit/delete are out of scope. Don't add them.
- **Never weaken the voting invariants.** One vote per user per feature (DB-enforced)
  and no self-voting are correctness requirements, not preferences.
- **Secrets stay in env.** No hardcoded JWT secrets, DB credentials, or tokens in
  code or commits. Never log tokens or password hashes.
- **Trust the database for uniqueness.** App-layer checks are for friendly errors;
  the unique constraints are the source of truth. Don't rely on app checks alone
  for concurrency-sensitive rules.

## Conventions

- **Ports & layout** are defined in `ARCHITECTURE.md` / `DECISIONS.md` — read them
  there, don't memorize copies here.
- **Match surrounding style.** Write code that reads like the code already in each
  package (naming, structure, comment density). Backend follows vertical-slice
  layout (`handler` / `service` / `repository` / `model` per feature).
- **Tests:** unit-only and mock-based (no DB writes in unit tests). Prioritize the
  service layer and the business/auth/voting rules called out in `ARCHITECTURE.md`.

## Definition of done (per change)

1. Behavior matches the frozen contract in `DECISIONS.md`.
2. Relevant unit tests added/updated and passing.
3. Affected docs updated; no doc-vs-code or doc-vs-doc contradiction introduced.
4. `prompts.txt` updated for the turn.
