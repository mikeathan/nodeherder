---
name: go-senior-backend
description: Applies senior-level (10+ years) Go expertise, focusing on SOLID principles, Clean Architecture, and idiomatic performance. Use for architectural reviews, refactoring, or building robust backend services.
---

# Senior Go Backend & Clean Architecture Skill

When writing or reviewing Go code, you must adhere to the following Senior/Staff Engineer standards:

## 1. Clean Architecture & SOLID

- **Interfaces at the Consumer**: Define interfaces where they are _used_, not where they are implemented (idiomatic Go).
- **Dependency Injection**: Use constructor functions (`NewService(...)`) to inject dependencies. No global state or `init()` functions for logic.
- **Single Responsibility**: Keep packages focused. If a package is named `util`, suggest a more specific domain name (e.g., `jsonconv` or `authutils`).

## 2. Idiomatic Performance & Safety

- **Pointer Semantics**: Use pointers only when sharing state or for large structs (>64 bytes). Use value semantics for small, immutable data.
- **Context Propagation**: Always pass `ctx context.Context` as the first argument to functions involving I/O or long-running tasks.
- **Goroutine Management**: Never start a goroutine without knowing how it will stop. Use `WaitGroups` or `Channels` for synchronization.
- **Slice Optimization**: Pre-allocate slice capacity with `make([]T, 0, length)` if the size is known to avoid unnecessary allocations.

## 3. Production-Ready Patterns

- **Table-Driven Tests**: Always generate tests using the table-driven pattern, including `Parallel()` execution where safe.
- **Functional Options**: Use the Functional Options pattern (`WithTimeout`, `WithRetry`) for complex struct initialization instead of multiple constructors.
- **Error Handling**: Use `errors.Is` and `errors.As` for checking sentinel errors or custom error types.

## 4. The "Senior Check" Workflow

Before outputting code, verify:

1. Is it readable? (Clear names > clever code).
2. Is it testable? (Are dependencies mocked/interfaced?).
3. Does it handle the "happy path" last? (Return early on errors).

## Constraints

- No `panic()` in production code. Use `error` return values.
- No `interface{}` (any) unless strictly necessary for generic data handling.
- Documentation must follow `godoc` conventions.
