# Garden

Garden is a local context and memory router for coding agents.

Core model:

```txt
.garden/ = source of truth
garden pack = ranked context pack from matching stored memories
agent uses pack = no giant AGENTS.md
```

The product question:

```txt
What should the agent know right now?
```

## Direction

Garden turns a growing pile of repo memory into a tiny task-specific context pack.

Garden is most useful for what the codebase does not reliably say. High-signal memories include rationale, gotchas, sparse intended patterns, local exceptions, workflow constraints, and product intent. Users can store anything; the first version retrieves from `.garden/memories.json` using explicit metadata like scope, task text, tags, priority, and budget.

Start local-first:

```txt
garden init
garden remember "Route files should not import DB clients directly; query modules enforce tenant scoping and audit logging." --scope "src/routes/**" --tag database
garden pack --path src/routes/api/users.ts --task "add user endpoint"
```

`garden remember` requires either `--scope` or `--always`. `--scope` and `--tag` can be repeated. `--always` means the memory can be considered for any path, not that it is forced into every pack.

`garden pack` outputs structured Markdown for LLM context:

```md
<garden_context_pack>
# Garden Context Pack

Path: `src/routes/api/users.ts`
Task: add user endpoint

## Relevant Memories
- Route files should not import DB clients directly; query modules enforce tenant scoping and audit logging.
</garden_context_pack>
```

## MVP Principles

- Manual memory entry first.
- Flat-file storage first.
- Deterministic retrieval first.
- Prove the simplest useful loop: remember scoped memory, then retrieve relevant stored memories for this path/task.
- No exports to existing agent systems yet.
- No refinement workflow, indexing, JSON output/API mode, database, vector search, MCP, or SaaS dashboard in the first slice.

## Product Bet

Your repo memory can grow without making every agent session dumber.
