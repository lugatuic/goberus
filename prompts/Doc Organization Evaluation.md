# Documentation Folder Organization Evaluation

## Current State

**Root-level docs/config:**
- `ADR/` — 1 file (middleware design pattern)
- `Agents/` — 1 file (agents overview)
- `docs/` — 2 files (dev setup, integration testing)
- `TODO.md` — Project issues/features (root)
- `CHANGELOG.md` — Release history (root)
- `README.md` — Project overview (root)
- `prompts/` — Agent runbooks/plans (NEW: moved from .github/prompts)

**Config:**
- `config/` — Go package, not docs

**Total scattered docs:** 7 locations

## Analysis

### By Folder

1. **ADR/** ✅ Should consolidate
   - Purpose: Architecture Decision Records (important long-term)
   - Current: 1 file (2025-12-18_middleware_design_pattern.md)
   - Action: Move to `docs/architecture/ADR/` or `docs/adr/`
   - Reasoning: Architectural decisions belong with other design docs

2. **Agents/** ⚠️ Needs clarification
   - Purpose: Agent overview/guide (unclear — only 1 .md file)
   - Current: Agents.md
   - Questions:
     - Is this for future AI/automation agents?
     - Does it document how agents interact with the project?
     - Is it just a note?
   - Recommendation: Move to `docs/development/agents.md` or keep as separate `agents/` for agent source code (if agents are scripts)

3. **docs/** ✅ Foundation for consolidation
   - Current: dev-setup.md, integration-testing.md
   - Action: Keep as base; add subdirs (architecture/, development/, operations/, api/)

4. **prompts/** ✅ Already moved
   - Purpose: Agent runbooks, implementation plans
   - Current: 4 Plan files
   - Action: Keep at root (for visibility and agent access)

5. **TODO.md** ⚠️ Should consolidate
   - Purpose: Issue tracking, feature roadmap
   - Current: Project-wide issues
   - Options:
     - **Option A**: Keep as root (GitHub best practice — visible on repo landing page)
     - **Option B**: Move to `docs/project-status.md` or `docs/roadmap.md`
   - Recommendation: **Keep at root** — standard convention, high visibility

6. **CHANGELOG.md** ✅ Keep at root
   - Standard convention, GitHub recognizes it

7. **README.md** ✅ Keep at root
   - Standard convention, GitHub recognizes it

## Recommended Structure

```
goberus/
├── README.md                    # Project overview
├── CHANGELOG.md                 # Release history
├── TODO.md                      # Feature roadmap (KEEP AT ROOT)
├── VERSION
├── docs/                        # Consolidated documentation
│   ├── architecture/
│   │   └── ADR/
│   │       └── 2025-12-18_middleware_design_pattern.md  (MOVE from ADR/)
│   ├── development/
│   │   ├── dev-setup.md         (MOVE from docs/)
│   │   ├── integration-testing.md (MOVE from docs/)
│   │   ├── agents.md            (MOVE from Agents/Agents.md)
│   │   └── running-tests.md     (NEW)
│   ├── operations/
│   │   ├── ldaps-setup.md       (NEW: from Dockerfile comments)
│   │   └── docker-guide.md
│   └── api/
│       ├── overview.md
│       ├── GET-member.md
│       ├── POST-member.md
│       ├── PATCH-member.md      (NEW)
│       └── DELETE-member.md     (NEW)
├── prompts/                     # Agent runbooks
│   ├── Plan - LDAP Refactoring.md
│   ├── Plan - PATCH Method Support.md
│   ├── Plan - DELETE Method Support.md
│   ├── Plan - Project Structure Refactoring.md
│   └── ...
├── ADR/                         # DELETE (move content to docs/architecture/ADR/)
├── Agents/                      # DELETE (move content to docs/development/agents.md)
└── ... (rest of project)
```

## Migration Strategy

### Phase 1a: Immediate (this PR - optional)
- [ ] Create `docs/` subdirectories (architecture/, development/, operations/, api/)
- [ ] Keep existing docs in place (don't move yet)
- [ ] Document the new structure for review

### Phase 1b: After approval (separate PR)
- [ ] Move `ADR/` → `docs/architecture/ADR/`
- [ ] Move `Agents/Agents.md` → `docs/development/agents.md`
- [ ] Move `docs/dev-setup.md` and `docs/integration-testing.md` to `docs/development/`
- [ ] Delete now-empty `ADR/` and `Agents/` folders
- [ ] Update all links in docs and code (find . -name "*.md" | xargs grep -l "ADR/\|Agents/")
- [ ] Update `.github/workflows` if any reference old paths
- [ ] Update `README.md` with new doc structure

## Validation Checklist
- [ ] All files preserved (no data loss)
- [ ] All internal links updated (docs, code comments)
- [ ] `git log` shows proper history/blame
- [ ] GitHub recognizes `README.md`, `CHANGELOG.md` at root
- [ ] Test that prompts/ is accessible from GitHub Actions

## Open Questions
1. **Agents.md**: Is this for AI agents, deployment agents, or something else? Clarify purpose.
2. **API docs**: Should these be auto-generated from code (Swagger/OpenAPI) or hand-written?
3. **Operations docs**: Should Docker setup, LDAPS cert handling, production deployment go here?
4. **Config package**: Should there be a `docs/api/config.md` explaining the Config struct?

## Timeline
- **This PR**: Propose structure, create skeleton directories
- **Follow-up PR**: Migrate files, update links, validate
- **Phase 2+ PRs**: Add new docs as features are implemented (PATCH, DELETE, etc.)
