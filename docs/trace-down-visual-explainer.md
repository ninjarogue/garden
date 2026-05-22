# Trace-Down Visual Explainer

Main goal: make the visual explainer help us understand the commit path fast.

Sibling note: `learn-up-visual-explainer.md` covers the separate climb back up through layer responsibilities.

Success means:

- One clear entry point.
- One simple trace-down flow.
- Only important files and concepts.
- No decorative noise.
- No "learn up" until we explicitly switch to it.
- Do not mix responsibility/meaning explanations into the trace-down diagram.
- Plain terms over technical completeness.
- Clarity beats brevity; simple labels can be longer when needed.
- Diagram answers: "where do I start, where does it go, what matters next?"

## Current Experiment

- Start with command flow.
- Optimize for: what command runs, what file it reads, what database effect happens.
- Use short labels with purpose, e.g. `drizzle.config.ts loads DB URL + schema path`.
- Compare against tiny labels or more annotated labels later if needed.
- First pass should show the happy path only: `db:migrate` to database changes.

## Experiment Results

- The simple command-flow diagram worked better than the first, broader visual explainer.
- Left-to-right happy path made the trace easier to follow.
- Short labels with purpose were clearer than detailed explanation blocks.
- Adding a separate "questions to ask" strip did not work well; it added another thing to read and increased cognitive load.
- Ordinary wording matters more than technical precision in the first diagram.
- Arrow labels should be short plain relationship phrases, not jargon.
- Square labels can carry extra words when that makes the step easier to understand.
- The goal is not the fewest words; the goal is the clearest first pass.
- Better example from seed refactor: `adds` -> `seed builder / adds IDs, dates, links`.
- Worse example from seed refactor: `fills in` -> `seed builder / IDs + links`.
- Important correction: arrows must not imply the wrong actor. `seed content` does not `add`; it `goes into` the builder. The builder adds database details.

## Current Rule

- Visual explainer shows the path, not the FAQ.
- Visual explainer shows what runs, not why each layer exists.
- Keep the diagram clean and let questions emerge naturally from the nodes.
- When a node raises a question, zoom into that node verbally or with a tiny focused diagram.

## Trace-Down Diagram Rule Set

- Start with the entry point command or route.
- Show the happy path first.
- Use node labels with purpose: thing + what it does.
- If simplicity and clarity conflict, choose clarity.
- Square labels should be as long as needed to be plain and unambiguous; arrows should stay short.
- Labels must never be cut off. If a label clips, fix the layout before accepting the diagram.
- Use ordinary words first; avoid internal terms like `row expansion`, `normalized rows`, `semantic content`, `orchestration`, `adapter`, or `transaction` in the first diagram unless the user already said them.
- Prefer plain box wording: `opens the database connection`, `turns seed content into table data`, `human-written seed content`, `adds IDs, dates, lifecycle, links`, `arrays ready for database insert`, `deletes old Taxonomy seed, inserts new one`.
- Make arrows short plain relationship phrases: `runs`, `calls`, `reads from`, `goes into`, `turns into`, `returns`, `saves to`.
- Pair short arrows with clear boxes. Good examples: `seed content` -> `goes into` -> `seed builder / adds database details`, then `seed builder` -> `returns` -> `table data / rows ready to insert`.
- The arrow label must be true of the source-to-target relationship.
- Do not make the arrow do work that belongs inside a box.
- Do not let the arrow imply the wrong actor.
- If the arrow repeats the box, delete it or make it explain the relationship.
- If the arrow and box together need explanation, first try a slightly longer plain box label.
- Add short arrow labels only when they explain why one step leads to the next.
- Keep arrow labels to roughly 1-5 plain words; use the shortest phrase that stays true and clear.
- Avoid side panels, FAQ strips, broad summaries, or extra branches unless the user asks.
- If a step raises a question, pause and explain that step before expanding the diagram.
- Stop at the output/result of the path; move responsibility/meaning into a learn-up diagram.

## Label Cutoff Rules

- Never solve cutoff by making wording unclear.
- Add `<br/>` line breaks at natural phrase boundaries.
- Keep each line short enough to fit the node comfortably.
- Increase diagram width, node spacing, or reduce font size slightly if needed.
- Reopen the diagram in the browser and visually check every node before calling it done.
- If Mermaid sizes a node badly, prefer clearer multi-line labels over one long label.

## Label Test

- Would a non-specialist understand the label without asking what the words mean?
- Is the arrow a plain relationship instead of a concept?
- Is the arrow true about the thing it starts from and the thing it points to?
- Does the box say the concrete result of that step?
- Is every word visible in the rendered diagram?
- If a shorter label creates a question, use a longer plain label.
- If the label is clear but not tiny, keep it.
