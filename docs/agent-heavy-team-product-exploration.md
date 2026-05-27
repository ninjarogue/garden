# Agent-Heavy Team Product Exploration

We explored a new direction for Garden aimed at small dev teams relying heavily on coding agents.

Core thesis:

```txt
Rule files are advisory constraints.
Tests and analysis are physical constraints.
Garden should connect the two.
```

The stronger product idea is that Garden should not try to make agents obey rules directly. Instead, it should tell teams:

- Which repo constraints apply to a change.
- Which context cards explain those constraints.
- Which checks enforce them.
- Whether the PR touched verification surfaces like tests, CI, lint config, or check definitions.

We discussed several feature ideas:

- **Changed-files context check:** map changed files to relevant `.garden/context/*.md` cards using deterministic scope globs.
- **Approved check registry:** cards reference check IDs, while `.garden/checks.yml` defines the commands.
- **PR Constraint Report:** CI generates a Markdown report showing changed areas, relevant cards, required checks, check results, and verification surfaces changed.
- **Trust zones:** classify changed files as implementation, tests, CI, config, generated, agent-context, etc.
- **Agent work receipt:** a local/CI summary of matched cards, checks run, unverified constraints, and verification surface changes.

We clarified that "weakened checks" was too strong. Better wording:

```txt
whether the PR changed verification surfaces
```

That means Garden does not claim intent or prove weakening. It simply flags when a PR touches things that enforce constraints, such as tests, workflows, lint configs, or check definitions.

For CI/reporting, the recommended first step is:

```sh
garden pr report --changed-file-list changed-files.txt --format markdown
```

Then GitHub Actions can append it to:

```sh
$GITHUB_STEP_SUMMARY
```

Later, the same Markdown could become a sticky PR comment, but Garden should start with stdout/Markdown output rather than owning GitHub API integration immediately.
