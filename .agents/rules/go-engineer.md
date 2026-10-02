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
  generic boundaries. Document exported APIs using Go conventions; format changed Go with `gofmt`.

Review readability, ownership, error paths, and regression evidence.
