# SPECS

SPECS is the source of truth for MessageFormat Go design contracts. User guides live in [`docs/`](../docs/), and the root [`README.md`](../README.md) stays focused on installation, examples, and development commands.

## Index

| Spec | Owns |
|------|------|
| [`00-overview.md`](00-overview.md) | Product and module scope, compatibility targets, runtime and verification guarantees |
| [`20-api-contracts.md`](20-api-contracts.md) | Root and MF1 public APIs, data model, functions, values, ownership, and errors |
| [`40-architecture.md`](40-architecture.md) | Processing pipeline, package/module boundaries, intl bridge, tests, and CI ownership |

## Rules

- Put design contracts here when they can be violated by code or documentation changes.
- Keep tutorials, examples, and command usage in README or `docs/`.
- Keep AI-agent workflow instructions in `AGENTS.md` and `CLAUDE.md`.
- Prefer updating one owning spec over duplicating the same rule across files.
