# Compact Index Syntax

Garden uses a Vercel-style compact index syntax for generated always-on indexes in `AGENTS.md`.

This is not a new universal standard. It is a compact, line-oriented, pipe-delimited index pattern optimized for agent context density.

## Base Pattern

Follow the Vercel-style pattern:

```txt
[Index Title]|root:<path>
|IMPORTANT:<instruction>
|<path-or-scope>:{<comma-separated-items>}
```

This base syntax is allowed:

- bracketed index title
- `|root:<path>` metadata
- `|IMPORTANT:<instruction>` priority instruction
- one indexed row per path, scope, or directory
- `{...}` for compact grouped items

## Where To Use It

Use compact syntax only for generated indexes that are loaded every session, such as the Garden context-card index.

Do not use compact syntax for the human-readable repo constitution. Project purpose, validation commands, project structure, conventions, and docs pointers should stay in normal Markdown.

## Vocabulary Approval Rule

Any Garden-specific compact-index vocabulary must be explicitly approved before implementation and documented in this file.

Garden-specific vocabulary means compact field names or row qualifiers that Garden invents beyond the base Vercel-style pattern. Examples that require approval before use:

- `cmd:`
- `mem:`
- `tags:`
- `ids:`
- `scope:`
- `score:`

Do not introduce stable compact-index vocabulary only in code, examples, or generated output. Add it here with meaning, allowed values, and an example.

## Approved Garden Vocabulary

No Garden-specific compact field vocabulary has been approved yet.

Current examples should use the base pattern only:

```txt
[Garden Context Index]|root:.garden/context
|IMPORTANT:Before editing a listed area, inspect the matching context card
|src/routes/**:.garden/context/routes-query-modules.md
|internal/contextcard/**:.garden/context/context-card-format.md
```

Card paths like `.garden/context/routes-query-modules.md` are plain index items, not compact-index field vocabulary.

## Future Candidates

No future Garden vocabulary is planned right now.

If a candidate is approved later, move it to `Approved Garden Vocabulary` with a precise definition.
