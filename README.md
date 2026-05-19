# Garden

Garden is a local context and memory router for coding agents.

Core model:

```txt
.garden/ = source of truth
garden pack = retrieve only relevant context
agent uses pack = no giant AGENTS.md
```

The product question:

```txt
What should the agent know right now?
```

## Direction

Garden turns a growing pile of repo memory into a tiny task-specific context pack.

Start local-first:

```txt
garden init
garden remember "Use server query functions instead of inline SQL" --path "src/routes/**" --tag database
garden pack --path src/routes/api/users.ts --task "add user endpoint"
```

## MVP Principles

- Manual memory entry first.
- Flat-file storage first.
- Deterministic retrieval first.
- No exports to existing agent systems yet.
- No indexing, database, vector search, MCP, or SaaS dashboard in the first slice.

## Product Bet

Your repo memory can grow without making every agent session dumber.
