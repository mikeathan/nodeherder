# Tasks: Wireless observation sources

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: draft

Each task maps to requirement/acceptance IDs or a constitutional gate. Mark done only
with evidence. No task is started; this file sequences authorized implementation.

## Preparation

- [x] T001 — **DONE 2026-11-11**: additive `Capabilities []string` on `devices.Device`
  (ADR-001, accepted). Q2 global and Q3 seven-day retention also resolved. IDs: none
  (unblocked FR-07); evidence: spec.md §Q1 resolution.
- [x] T002 — **DONE**: every rule has a result or justified N/A in `plan.md` §Constitution Check.

## Coverage and implementation

- [x] T003 — **DONE**: identity-resolution table test in
  `backend/internal/observations/ble_test.go`; IDs: AC-03, AC-04; evidence: 7 test functions
  covering payload precedence, address-type classification, unusable addresses, fingerprint
  stability, local names, malformed data and Eddystone UID.
- [x] T004 — **DONE**: `backend/models/observations/` and `ble.go`; IDs: FR-02; evidence:
  `go test ./internal/observations/...` passes; GoDoc on every exported declaration.
- [x] T005 — **DONE**: bounded-queue and cadence tests in `service_test.go` using
  `mocks.NewMockClock`; IDs: AC-05, AC-06, NFR-02.
- [x] T006 — **DONE**: `service.go` (aggregation, bounded queue, change-driven publish);
  IDs: FR-01, FR-03, FR-04, FR-09, NFR-01; evidence: `-race` clean.
- [x] T007 — **DONE**: `models/settings/observation_cache.go` persistence, `observations`
  WS event and the two triage actions; IDs: FR-05, FR-06, AC-07, AC-08; evidence:
  `models/settings` and `internal/ws` tests pass; payload validation in
  `internal/observation_handlers.go`.
- [x] T008 — **DONE**: optional `Observer` seam and BLE AD decoder in
  `backend/internal/deviceproviders/esphome/advertisement.go`; IDs: FR-01, AC-01, AC-02,
  AC-03; evidence: `advertisement_test.go` proves capability is structural (a plain sink does
  not satisfy `Observer`).
- [x] T009 — **DONE**: composition in `backend/internal/hub.go` via
  `WithObservationConsumer`, mirroring the spatial recorder; IDs: FR-05; evidence: full
  backend suite passes with no `AppStore` change.
- [x] T010 — **DONE**: `src/contracts/observations.ts`, `src/types/observations.type.ts`,
  `src/domain/observations/{capability,presentation}.ts`; IDs: FR-07, FR-08, AC-09, AC-11,
  AC-12; evidence: 34 new Jest tests pass.
- [x] T011 — **DONE**: `observations` Vuex module and WS mapping; IDs: FR-05; evidence:
  module registered, event mapped, full suite passes.
- [x] T012 — **DONE**: `WirelessViewer`, `ObservationRadar`, `ObservationTable`,
  `ObservationFilters`, route and capability-gated nav; IDs: FR-07, NFR-04, AC-10;
  evidence: navigation gating tests pass; `npm run build` succeeds.
- [x] T013 — deferred: the per-device "Emit observations" settings row was not added. The
  capability is declared and consumed; the device-page row is presentation only. Recorded as
  follow-up so the gap is explicit rather than implied complete.

## Handoff

- [x] T014 — **Checks run** (actual results in `plan.md` §Traceability): backend
  `go build ./...`, `go vet ./...`, `go test ./...` (52 packages pass),
  `go test -race` on affected packages (clean); frontend `npm test -- --runInBand` (149 suites,
  859 tests pass), `npm run lint` (0 errors; 23 pre-existing warnings elsewhere),
  `npm run build` (succeeds).
- [ ] T015 — Partially done: offline/stale/empty states implemented and lint-clean, but
  keyboard and 390px/1280px browser verification is **unrun**. Gate FE-04 is not yet satisfied.
- [ ] T016 — Pending: docs and final rule audit.

Add discovered work here and update the spec/plan when scope changes.
