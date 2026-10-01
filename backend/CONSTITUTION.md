# Backend Constitution

Version: 1.0.0 | Status: Proposed for adoption | Last Amended: 2026-10-01
Parent: [NodeHerder v1.0.0](../.specify/memory/constitution.md)
Scope: `backend/` and backend integration/deployment behavior.

Parent governance applies; these additional rules do not certify existing compliance.

## BE-01 — Domain/transport separation

Handlers MUST validate/adapt HTTP/WS/MQTT/MCP inputs and delegate business rules to
established services/controllers; persistence MUST use repository/store boundaries.
Domain tests MUST work without live brokers/servers. New package dependencies MUST
avoid cycles and explain ownership. Plans identify affected layers; tests use fakes.

## BE-02 — Device update semantics

`DeviceLifetimeService` changes MUST verify repeated events, state deduplication,
fresh state before debounce, automation inputs, and sampled storage/broadcasts.
Consider seed/update/disabled/availability paths separately. Timing/cooldown behavior
MUST be specified/tested, not inferred from comments/fields. Regression evidence MUST
cover identical button events, unchanged state, and debounced state in automation reads;
tests: `internal/services/`, `internal/automations/`.

## BE-03 — Concurrency and lifecycle

New goroutines/timers/subscriptions/shared state MUST define ownership, cleanup,
and necessary synchronization. External calls MUST bound timeouts and define retries.
Concurrency tests MUST use controlled clocks/synchronization instead of arbitrary
sleeps where feasible and run race detection on affected packages.

## BE-04 — Storage/query meaning

Metrics queries MUST define aggregation, limits, time-boundary inclusivity, ordering,
and empty results. Schema/key/retention changes MUST include existing-data fixtures
and migration/recovery evidence. Success responses MUST NOT conceal failed writes.
Changed configuration defaults/startup handling MUST be documented.

## BE-05 — API/assistant boundaries

Changed routes MUST document authentication, errors, and success payloads; public
additions/access expansion require NH-03 review. MCP intent MUST validate targets,
metrics, aggregation, and time scope; ambiguity MUST NOT select arbitrary devices.
Assistant failures MUST produce bounded, observable errors without leaking secrets.

## Verification

Toolchain: [go.mod](go.mod). From `backend/`: `go test ./...`; implementation changes:
`go build ./...`; relevant Go changes: `go vet ./...`; concurrency changes:
`go test -race ./<affected-package>/...`. Record exact packages/results in the plan.
Use temporary storage/mocked MQTT; real-device checks require separately scoped manual
targets. Navigation: [repository map](../docs/sdd/repository-map.md).
