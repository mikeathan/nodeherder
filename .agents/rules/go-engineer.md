# Go rules

Scope: Go implementation, refactoring, review. Authority:
[backend constitution](../../backend/CONSTITUTION.md), including commands and BE-03 lifecycle.

- Keep packages focused; rename/restructure only for scoped benefit.
- Define small consumer interfaces when substitution aids testing. Inject dependencies
  through constructors; avoid new global mutable business state. Use functional options
  for meaningful independent configuration.
- Choose pointers for identity, mutation, sharing, or measured copy costs, without
  fixed size thresholds. Preallocate useful known capacities.
- Propagate context through I/O and long-running work; preserve compatible signatures
  when changing them expands scope.
- Return contextual errors, wrapping causes with `%w`; inspect with `errors.Is/As`.
  Return early on failure; avoid panics for expected errors.
- Use table-driven tests for repeated cases; parallelize only isolated state.
  Reuse suitable `backend/testing/` and `backend/mocks/` helpers; inspect their cleanup.
  Reuse `utils.Clock` (`backend/utils/time.go`) for controllable time where applicable.
- Prefer concrete types; reserve `any` for heterogeneous external data or genuine
  generic boundaries. Format changed Go with `gofmt`.

- Add GoDoc to every new or changed exported function, method, type and constant. Describe the
  declaration's purpose and contract, including units/timezones, errors, mutation/ownership and
  meaningful nil/pointer semantics. Document important struct invariants, not just field spellings.
- Explain non-obvious private algorithms and control flow with concise comments: concurrency and
  cleanup, revision/generation fencing, transactions, credential boundaries, date handling and
  shared-slice copying. Avoid comments that only narrate an assignment or obvious syntax.

Review readability, ownership, error paths, and regression evidence. Before finishing, audit touched
public declarations for useful GoDoc and complex flows for explanatory comments; passing tests alone
does not satisfy the documentation/readability rules.
