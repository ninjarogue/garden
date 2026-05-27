---
scope:
  - .garden/context/**
  - internal/contextcard/**
  - internal/scopeglob/**
tags:
  - context
---

# Context Card Format

Context cards are Markdown files with small YAML frontmatter and a human-readable body.

Required frontmatter:

```yaml
scope:
  - internal/example/**
```

Optional frontmatter:

```yaml
tags:
  - testing
```

Rules enforced by `internal/contextcard`:

- The file slug is the card identity and must be lowercase words separated by hyphens.
- `scope` must be a non-empty YAML list.
- Each `scope` entry must be a valid Garden glob.
- `scope` cannot contain `CHANGE_ME`.
- `tags`, when present, must be a YAML list.
- Scope cannot contain compact-index delimiters or control characters that would make `AGENTS.md` unsyncable.
- Tags are human-only labels and are not rendered into the AGENTS index.
- The Markdown body cannot be empty.

Keep card bodies direct and operational. They should tell future agents what to preserve, what to avoid, and what verification matters for the scoped files.
