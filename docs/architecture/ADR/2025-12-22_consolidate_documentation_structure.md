# Consolidate documentation structure

## Status
Accepted (2025-12-22)

## Context
The project has documentation scattered across 7 root-level locations:
- `ADR/` — 1 file (middleware design pattern)
- `Agents/` — 1 file (agent workflow notes)
- `docs/` — 2 files (dev setup, integration testing)
- `TODO.md` — Feature roadmap (root)
- `CHANGELOG.md` — Release history (root)
- `README.md` — Project overview (root)
- `prompts/` — Agent runbooks and plans (newly moved from .github/prompts)

This fragmentation creates friction for new contributors, makes discoverability difficult, and will worsen as features like PATCH, DELETE, and group management are added. Documentation lacks a clear organizational hierarchy, mixing architecture decisions, development guides, operational procedures, and feature plans in random locations.

Inspired by [ADR examples](https://github.com/joelparkerhenderson/architecture-decision-record/tree/main/locales/en/examples), we need a consolidated structure that:
- Groups related documentation
- Provides clear navigation
- Scales as the project grows
- Separates concerns (architecture, development, operations, API)

## Decision
Adopt a consolidated documentation structure organized by concern:

```
docs/
├── architecture/          # Architecture decisions & design patterns
│   └── ADR/
│       ├── 0000_template.md
│       ├── 2025-12-18_middleware_design_pattern.md
│       └── 2025-12-22_consolidate_documentation_structure.md
├── development/           # Developer guides & workflows
│   ├── agents.md          (from Agents/Agents.md — single source of truth for agent behavior + workflow)
│   ├── dev-setup.md
│   └── integration-testing.md
└── prompts/               # Agent runbooks & implementation plans (from .github/prompts/)
    └── Plan - IntegrationTestsAndCICD.prompt.md

(root)
├── README.md              # Project overview (standard)
├── CHANGELOG.md           # Release history (standard)
├── TODO.md                # Feature roadmap (standard, high visibility)
└── version.txt            # Semver version (read by Makefile/Dockerfile)
```

Keep `README.md`, `CHANGELOG.md`, `TODO.md`, and `version.txt` at root per GitHub conventions.
Additional `docs/` subdirectories (e.g., `api/`, `operations/`) should only be created when there is actual content to put in them.

## Consequences

### Positive
- **Better discoverability**: All documentation in `docs/` with clear subfolders
- **Scalability**: New subdirectories can be added when content exists to fill them
- **Agent-friendly**: `agents.md` is the single source of truth for agent behavior and workflow
- **No empty placeholders**: Directories only exist when they contain real content

### Negative
- **Migration effort**: Moved files, updated cross-references and CI/CD references
- **ADR discoverability**: ADRs now nested in `docs/architecture/ADR/` instead of root-level `ADR/`

### Mitigation
- Updated `README.md` with link to `docs/`
- Validated all links after migration

## Related ADRs / Links
- [ADR Examples — GitHub](https://github.com/joelparkerhenderson/architecture-decision-record)
- [agents.md](../../development/agents.md) — Agent workflow documentation
- [Plan - IntegrationTestsAndCICD](../../prompts/Plan%20-%20IntegrationTestsAndCICD.prompt.md) — CI/CD implementation plan
