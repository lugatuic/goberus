# ADR Template

Use this template when documenting architecture decisions. Agents will parse this format to understand design rationale.

```markdown
# [Decision Title — short, declarative phrase]

## Status
[One of: Proposed, Accepted, Deprecated, Superseded By #XXX]

## Context
[What is the issue that we're seeing that is motivating this decision?]
[What is the background or history that led to this decision?]
[What constraints exist?]
[What alternatives were considered?]

## Decision
[What is the change that we're proposing and/or doing?]
[Be explicit about the decision; avoid vagueness.]

## Consequences
[What becomes easier or possible after this decision?]
[What becomes harder or more complex?]
[What are long-term trade-offs?]
[What follow-up decisions are required?]

## Rationale
[Optional: Why this decision over alternatives?]
[Reference external sources, discussions, or evidence.]

## Related ADRs / Links
[Links to related decisions, documentation, or external references.]
```

## Naming Convention
Use format: `YYYY-MM-DD_decision-title-slugified.md`

Example: `2025-12-18_middleware_design_pattern.md`

## Agent Consumption
Agents parsing ADRs should extract:
1. **Decision Title** (from H1)
2. **Status** (enum: Proposed, Accepted, Deprecated, Superseded By)
3. **Context** (understand the problem)
4. **Decision** (what changed)
5. **Consequences** (understand trade-offs)

Use this structure for programmatic parsing and validation.
