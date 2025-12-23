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
│   ├── dev-setup.md
│   ├── integration-testing.md
│   ├── agents.md          (from Agents/Agents.md)
│   └── running-tests.md   (future)
├── operations/            # Deployment, LDAPS, infrastructure
│   ├── ldaps-setup.md     (future: from Dockerfile comments)
│   └── docker-compose-guide.md
└── api/                   # REST API reference
    ├── overview.md
    ├── GET-member.md
    ├── POST-member.md
    ├── PATCH-member.md    (future)
    └── DELETE-member.md   (future)

prompts/                   # Agent runbooks (remains at root for visibility)
├── Plan - LDAP Refactoring.md
├── Plan - PATCH Method Support.md
├── Plan - DELETE Method Support.md
├── Plan - Project Structure Refactoring.md
└── ...

(root)
├── README.md              # Project overview (standard)
├── CHANGELOG.md           # Release history (standard)
├── TODO.md                # Feature roadmap (standard, high visibility)
├── ADR/                   # DEPRECATED (migrate to docs/architecture/ADR/)
└── Agents/                # DEPRECATED (migrate to docs/development/agents.md)
```

Keep `README.md`, `CHANGELOG.md`, and `TODO.md` at root per GitHub conventions.

## Consequences

### Positive
- **Better discoverability**: All documentation in `docs/` with clear subfolders
- **Scalability**: New features (PATCH, DELETE, group management) have designated homes
- **Agent-friendly**: Organized structure makes it easier for agents to parse and reference docs
- **Single source of truth**: Architecture decisions, dev guides, and API reference coexist
- **Growth-ready**: `docs/api/` and `docs/operations/` prepared for expansion

### Negative
- **Migration effort**: Move files, update cross-references, update CI/CD references (if any)
- **Temporary disruption**: Links in code comments and workflows may break during transition
- **ADR discoverability**: ADRs now nested in `docs/architecture/ADR/` instead of root-level `ADR/`

### Mitigation
- Update `README.md` with link to `docs/`
- Create redirects/aliases if necessary
- Validate all links during migration (find . -name "*.md" | xargs grep -l "ADR/\|Agents/\|docs/")
- Test that GitHub Actions and CI workflows can still find docs

## Related ADRs / Links
- [ADR Examples — GitHub](https://github.com/joelparkerhenderson/architecture-decision-record)
- [Agents.md](../Agents/Agents.md) — Agent workflow documentation
- [Plan - Project Structure Refactoring](../prompts/Plan%20-%20Project%20Structure%20Refactoring.md) — Implementation runbook
