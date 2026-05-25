# Property-Based Testing Findings

These notes capture the PBT research and implementation findings from this session. The intended use is source material for a future agent skill that helps identify, design, implement, and review property-based tests.

## Terminology

PBT means property-based testing, not product-based testing.

Property-based testing checks that broad behavioral properties hold across many generated inputs. Example-based tests check a few named scenarios. PBT should complement example tests, not replace them.

## Key Findings

- PBT is useful when correctness can be expressed as an invariant, round trip, model equivalence, ordering rule, conservation law, or metamorphic relation.
- PBT is strongest on pure functions and deterministic domain logic: parsers, serializers, validators, reducers, formatters, permission rules, pricing, scheduling, routing, sorting, and state machines.
- PBT can work on IO-heavy systems when effects are modeled or simulated, such as fake filesystems, fake networks, fake clocks, in-memory queues, or reference models.
- Generated failures are most valuable when the tool shrinks them to a minimal counterexample that can become a regression test.
- Agents can help propose candidate functions, identify properties, generate valid input domains, write the test, run it, and investigate failures. Human review remains necessary because weak or false properties produce misleading confidence.

## Source Notes

- Bob Martin/X: property testing is another hardening technique agents can use. He describes agents identifying suitable functions, specifying domain/range, implementing/running property tests, and fixing bugs found this way. Source found at `https://x.com/unclebobmartin/status/2058184649575133291`.
- James Long/X: described using property-based testing for OpenCode with simulated filesystem/network/process behavior and said he is bullish on the approach. Source found at `https://x.com/jlongster/status/2058200254046990772`.
- Hypothesis: PBT is about generated input exploration and behavioral properties, not one specific language or type system. Source: `https://hypothesis.works/articles/what-is-property-based-testing/`.
- fast-check: useful JavaScript/TypeScript framing around arbitraries, generated runs, shrinking, and counterexamples. Source: `https://fast-check.dev/docs/introduction/what-is-property-based-testing/`.
- Scott Wlaschin/F# for Fun and Profit: practical property patterns such as round trips, different paths same destination, invariants, idempotence, structural induction, and test oracles. Source: `https://fsharpforfunandprofit.com/posts/property-based-testing/`.
- Johannes Link/How to Specify It: strong taxonomy for specifying properties, including postconditions, invariants, metamorphic properties, inductive properties, and model-based properties. Source: `https://johanneslink.net/how-to-specify-it/`.

## Property Patterns

Use these patterns when looking for useful properties.

| Pattern | Question | Example |
| --- | --- | --- |
| Round trip | Can data go there and back unchanged? | `parse(render(x)) == x` |
| Idempotence | Does repeating the operation stop changing the result? | `normalize(normalize(x)) == normalize(x)` |
| Invariant | What must always be true about the output? | `sorted(xs)` is ordered and has same elements |
| Conservation | What quantity must be preserved? | Total debit equals total credit |
| Model equivalence | Does implementation match a simple reference model? | Optimized lookup matches map-based lookup |
| Metamorphic relation | Does changing input predictably change output? | Adding unrelated data does not change result |
| Commutativity | Does operation order not matter? | Merging independent sets produces same result |
| Monotonicity | Can output only move one direction as input changes? | More permissions cannot reduce access |
| Validation boundary | Do all accepted generated values satisfy contract? | Valid slug always parses as a card slug |
| Error containment | Do invalid values fail safely? | Malformed frontmatter always returns an error |

## Agent Workflow For Applying PBT

1. Inspect the code and existing tests first.
2. Identify deterministic logic with a clear input/output contract.
3. Prefer one small property over a broad, vague property.
4. Define the generated domain narrowly enough to produce meaningful valid inputs.
5. Add example tests for known edge cases if readability would suffer.
6. Write the property test with strong failure logs and limited runtime.
7. Run the relevant test package, then the full suite when feasible.
8. If a property fails, determine whether the bug is in the code, the generator, or the property.
9. Keep minimized counterexamples as regression examples when they explain a real bug.
10. Do not add dependency-heavy PBT libraries unless the standard test tooling is insufficient.

## Language And Tooling Notes

- Go: start with standard `testing/quick` for simple generated properties. Use native fuzz tests for byte/string parser boundaries or crash discovery. Consider third-party libraries only when shrinking, generator composition, or state-machine testing requires it.
- TypeScript/JavaScript: use `fast-check` for arbitraries, shrinking, async properties, model-based tests, and command-based state-machine tests.
- Python: use `hypothesis` for generated strategies, shrinking, stateful testing, and broad ecosystem support.
- JVM: use jqwik or QuickTheories for property tests; Johannes Link's materials are especially relevant for jqwik.
- .NET/F#: use FsCheck or Hedgehog-style libraries where available.

## Findings From This Repo

Garden is a small Go CLI. It has pure logic that fits PBT well:

- `internal/contextcard`: Markdown frontmatter parsing, card template rendering, slug/title conversion, metadata validation.
- `internal/agents`: compact index rendering, block upsert, marker detection, sync behavior, lint findings.
- `internal/output`: deterministic diff and lint output formatting.

Initial properties added in this session:

- `internal/agents/context_index_property_test.go`: `RenderIndex` output is independent of input card order.
- `internal/contextcard/card_property_test.go`: `renderTemplate` followed by `Parse` round-trips slug, path, scope, tags, and generated body.

Verification command used:

```sh
env GOCACHE=/tmp/garden-go-build go test ./...
```

All tests passed after adding the initial properties.

## Good Future Garden Properties

- `SyncIndex` idempotence: syncing an already-synced document should not change it again.
- `UpsertBlock` preservation: content before and after the Garden managed block should remain unchanged.
- Marker validation: documents with duplicated, missing, or reversed markers should always return controlled malformed-marker errors.
- `Parse` invalid frontmatter containment: generated malformed frontmatter should fail without panics and with actionable errors.
- `unifiedDiff` stability: equal content always reports no changes; non-equal content always includes removed and added lines.

## Proposed Skill Shape

Possible skill name: `property-based-testing`.

Possible trigger description:

```yaml
name: property-based-testing
description: Identify, design, implement, and review property-based tests. Use when adding PBT coverage, choosing properties for pure functions or state machines, debugging generated counterexamples, selecting tools such as testing/quick, Go fuzzing, fast-check, Hypothesis, jqwik, or FsCheck, or converting example-based tests into stronger generated properties.
```

Suggested `SKILL.md` body should stay short and include:

- A brief workflow: inspect existing tests, pick candidate, select property, write generator, run, triage failures.
- A property pattern checklist.
- Tool selection guidance by language.
- Review warnings: avoid tautologies, overfitted generators, generated domains that exclude edge cases, and properties that simply restate implementation.

Suggested reference files:

- `references/property-patterns.md`: detailed catalog of round trip, invariants, metamorphic, model-based, and stateful properties.
- `references/go.md`: `testing/quick`, fuzzing, and Go generator examples.
- `references/typescript.md`: `fast-check` examples.
- `references/python.md`: Hypothesis examples.
- `references/review-checklist.md`: how to review PBT tests for weak properties and bad generators.

Suggested scripts are optional. A script may not be necessary at first. If repeated setup becomes common, add helper snippets or templates rather than a fragile automation script.

## Skill Guardrails

- Do not force PBT everywhere. If no meaningful property is visible, keep example tests.
- Do not add new test dependencies without checking the repo's existing stack and asking when the choice has tradeoffs.
- Do not treat generated random testing as proof. It is a bug-finding and hardening technique.
- Prefer deterministic seeds or clear reproduction output when the test framework supports it.
- Keep generated input domains valid unless the specific property is about invalid input handling.
- Log enough context for failures to be actionable.
