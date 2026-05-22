# Learn-Up Visual Explainer

Main goal: after tracing a path down, climb back up from the innermost layer and understand what each layer is responsible for.

Source concept: Mitchell Hashimoto's "trace down, learn up" from `https://mitchellh.com/writing/contributing-to-complex-projects`.

Related notes:

- `trace-down-visual-explainer.md` covers the outside-in path-following diagram.
- `trace-down-learn-up-reference.md` records the source concept and examples.

## Success Means

- Start at the innermost point found during trace down.
- Move from concrete implementation toward user-visible behavior.
- Explain what each layer teaches us about the layer above it.
- Keep one simple climb upward.
- Use plain wording over technical completeness.
- Clarity beats brevity; use more words when needed.
- Do not repeat the trace-down execution path unless it helps anchor the climb.

## Learn-Up Diagram Rule Set

- Start with the lowest-level implementation detail from the trace-down diagram.
- Use boxes for layer responsibility: thing + what we learn from it.
- Use arrows for support relationships: how the lower layer helps the next layer above it.
- Keep arrows short but true: `is read by`, `supports`, `scopes`, `feeds`, `powers`, `explains`.
- If an arrow implies the wrong actor, rewrite it.
- If a box label needs more words to be clear, use more words.
- Avoid generic product summaries until the final node.
- The final node can name the user-visible behavior.

## Box Label Patterns

Good learn-up labels:

- `Postgres tables / show what data actually exists`
- `project context / keeps reads scoped to Taxonomy`
- `dashboard query / asks for the pieces the route needs`
- `server function / keeps database work on the server`
- `route loader / gets data before the page renders`
- `Dashboard page / shows current project health`

Avoid labels that only restate execution order:

- `dashboard query / runs query`
- `server function / calls query`
- `route loader / calls server function`

Those belong in trace-down, not learn-up.

## Arrow Relationship Rules

- Arrows should describe support, not just sequence.
- The arrow label must be true from source to target.
- Good examples:
- `Postgres tables` -> `need scope from` -> `project context`
- `project context` -> `scopes` -> `dashboard query`
- `dashboard query` -> `supports` -> `server function`
- `server function` -> `feeds` -> `route loader`
- `route loader` -> `powers` -> `Dashboard page`

## Dashboard Commit Example

For `355c680 feat: read dashboard data from database`:

1. `Postgres tables / show what data actually exists`
2. `project context / keeps reads scoped to Taxonomy`
3. `dashboard query / asks for the pieces the route needs`
4. `server function / keeps database work on the server`
5. `route loader / gets data before the page renders`
6. `Dashboard page / shows current project health`

## Label Cutoff Rules

- Labels must never be cut off.
- Add `<br/>` line breaks at natural phrase boundaries.
- Keep each line short enough to fit the node comfortably.
- Increase diagram width, node spacing, or reduce font size slightly if needed.
- Reopen the diagram in the browser and visually check every node before calling it done.

## Label Test

- Does the diagram start at the innermost layer?
- Does each box say what that layer teaches us?
- Does each arrow say how the lower layer supports the next layer?
- Is every word visible in the rendered diagram?
- Does the final node connect back to user-visible behavior?
