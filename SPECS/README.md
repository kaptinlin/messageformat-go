# SPECS

SPECS is the source of truth for MessageFormat Go design contracts. User guides live in [`docs/`](../docs/), and the root [`README.md`](../README.md) stays focused on installation, examples, and development commands.

## Index

| Spec | Owns |
|------|------|
| [`00-overview.md`](00-overview.md) | Project scope, compatibility targets, runtime guarantees |
| [`20-api-contracts.md`](20-api-contracts.md) | Public API, data model, functions, values, errors |
| [`40-architecture.md`](40-architecture.md) | Processing pipeline, package boundaries, intl bridge, tests |

## Rules

- Put design contracts here when they can be violated by code or documentation changes.
- Keep tutorials, examples, and command usage in README or `docs/`.
- Keep AI-agent workflow instructions in `AGENTS.md` and `CLAUDE.md`.
- Prefer updating one owning spec over duplicating the same rule across files.
