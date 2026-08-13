# Domain Docs

**Layout:** Single-context

## Files

| File | Purpose |
|---|---|
| `CONTEXT.md` (repo root) | High-level domain overview — entities, invariants, key terminology |
| `docs/adr/` | Architecture Decision Records — one file per decision |

## Rules for agents

- Read `CONTEXT.md` at the start of any task that touches domain logic.
- Before creating an ADR, check `docs/adr/` to avoid duplicating an existing decision.
- ADR filenames follow the pattern `NNNN-short-title.md` (e.g. `0001-use-uuidv7.md`).
- Do not edit `CONTEXT.md` without explicit user instruction — it is a human-maintained document.
