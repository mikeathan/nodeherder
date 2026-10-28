---
name: clean-code
description: Apply and review NodeHerder's coding practices when implementing, refactoring, or reviewing Go or Vue/TypeScript code.
---

# Clean code

Before implementing, refactoring, or reviewing code, classify the affected behavior
and read the applicable constitutions and engineering rules routed by
[`AGENTS.md`](../../../AGENTS.md). Cross-stack work needs both language rules and
the integration rule. These files remain authoritative; this skill does not replace them.

When shaping or reviewing the change:

- Use intention-revealing names, focused functions, explicit side effects, and the
  existing ownership boundaries. Prefer a small, clear change to speculative abstraction.
- Trace error handling, external inputs, state ownership, cleanup, and meaningful edge
  cases. Apply the relevant language and integration rules to the actual change.
- Review the full task diff, including staged changes when relevant, for readability,
  duplication, accidental scope changes, and stale comments. Preserve unrelated work.
- Check affected public documentation and meaningful regression coverage against the
  applicable rules. Run the required checks and report actual results and gaps.

Use judgment: a smell is a prompt to investigate, not an automatic reason to rewrite
unrelated code. Follow the user's scope and the repository's existing conventions.
