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
- [x] T013 — **Superseded by T019**. The per-device row is now required rather than
  optional: it must show capability status and live counters, because a device can be
  connected yet silent.

## Reassessment round (2026-11-11, after hardware decision)

- [x] T017 (D-1, blocking bug) — **DONE**: `Service.RegisterSource` was never called in
  production, so `snapshot.sources` was always empty and the filter component always
  rendered its empty state. Added the `sourceRegistry` optional interface, announced on
  subscribe and withdrawn on session exit. IDs: AC-07, FR-05. Evidence:
  `TestObservationSourceRegistersOnceWhileObserving`,
  `TestNonObservingSinkIsNeverRegistered`, `TestUnregisterSourceRemovesDescriptor`.
- [x] T018 (D-2) — **DONE**: per-source `SourceDiagnostics` (advertisements, decoded,
  dropped, lastAdvertisement) reach the snapshot; `Dropped` is now wired from `hub.go`;
  drops are attributed to their source; the viewer states each source's delivery status
  and explains silence as a likely `bluetooth_proxy` omission. IDs: NFR-01, FE-03.
  Evidence: `TestSourceDiagnosticsReachTheSnapshot`,
  `TestRecordSourceActivityIgnoresUnknownSources`, `TestDiagnosticsSurviveReRegister`,
  `TestDropsAreAttributedToTheirSource`, plus four contract tests and five presentation
  tests.
- [ ] T019 — Device-page observation row: capability status and counters. **Open.** The only
  remaining code item. Not blocking: the viewer states per-source status, so the device page
  row is convenience rather than diagnosis.
- [x] T020 (FE-04) — **DONE**: `e2e/wireless.spec.ts` renders `/wireless` in a real browser at
  1280px and 390px and now covers 13 cases, including isolation, Focus distance and trend,
  Escape, and both viewports. The mock emits `observations`. Gate FE-04 is satisfied.
- [ ] T021 — Docs and final rule audit. **Partially done**: spec, plan, ADRs and contract are
  current; a final pass over the full diff is outstanding.
- [x] T022 — **DONE on real hardware**: the Everything Presence Lite connected, the
  subscription was accepted and advertisements decoded. This is what exposed D-10. Assumption
  A1 (message id 93) is confirmed.

### Bug-fix round (2026-11-11)

- [x] T023 (D-4) — **DONE**: diagnostics were missing from `renderSignature`, so a source
  whose advertisements all failed to decode never republished. That defeated D-2 precisely
  in the diagnosis case. Counters now participate in change detection. Evidence:
  `TestDiagnosticsAloneTriggerAPublish`, `TestUnchangedDiagnosticsDoNotRepublish`.
- [x] T024 (D-5) — **DONE**: identity RSSI was a lifetime maximum; now tracks the newest
  reading so distance follows movement. Evidence: `TestRSSITracksNewestReading`.
- [x] T025 (D-6) — **DONE**: advertised local name now becomes the identity label. Evidence:
  `TestDecodeAdvertisementsKeepsAdvertisedName`.
- [x] T026 (D-7) — **DONE**: diagnostics were counted twice (10 advertisements reported 18).
  Counting now has one owner. Evidence: `TestForwardingDoesNotDoubleCount`.
- [x] T027 (D-8, blocking) — **DONE**: `useStore()` returned undefined because the app installs
  the store under a Symbol injection key; the sidebar crashed on mount. Both components now
  import the `@/store` singleton. Evidence: `layout.spec.ts` failed before, passes after.
- [x] T028 (D-9) — **DONE**: removed the never-dispatched `stale` flag; the viewer derives
  from `ws/getConnectionStatus`.
- [x] T020 (FE-04) — **DONE**: `e2e/wireless.spec.ts` renders `/wireless` in a real browser at
  1280px and 390px, asserting identity rendering, the bearing-less disclaimer, facet filters
  appearing (D-1 regression guard), the only-new filter, keyboard focus on triage, and no
  horizontal overflow. The mock hub now emits `observations`. Evidence: 6/6 desktop and 6/6
  mobile; full suites 168 and 177 pass with 0 failures.

## Proximity round (2026-11-12)

Driven by using the viewer against real hardware. Triage and grouping did not make it a
tool: it was a census of MAC addresses.

- [x] T029 — **DONE**: "only new" filtered on the triage decision and matched every identity on
  a first run. Novelty now comes from `firstSeen` and a five-minute window. Evidence:
  `isRecentArrival` tests; e2e asserts an older identity is excluded.
- [x] T030 — **DONE**: pause/freeze, grouping and a list view. Freezing keeps a copy so triage
  still lands on screen, and a watcher drops rows that leave the live snapshot.
- [x] T031 (D-11) — **DONE**: isolation. Selecting an identity scopes every view to it; Escape
  and an explicit control clear it; a selection that ages out says so.
- [x] T032 — **DONE**: RSSI trend. A bounded 20-reading window per identity, older-half versus
  newer-half means, a 3 dB deadband and a 45 s staleness horizon. `unknown` is never shown as
  `steady`. Evidence: nine trend tests.
- [x] T033 — **DONE**: Focus view. Range rings without a bearing, distance through the shared
  `formatQuantity`, near/mid/far band, arrow-plus-word trend, signal sparkline. One column at
  every width. Evidence: 13 desktop and 13 mobile e2e.
- [x] T034 — **DONE**: trend and its sparkline participate in `renderSignature`; without that
  they would freeze on a live device. Evidence: `TestDiagnosticsAloneTriggerAPublish` pattern.

## Handoff

- [x] T014 — **Checks run**: backend `go build`, `go vet`, `go test ./...` (52 packages),
  `go test -race` on affected packages (clean); frontend `npm test -- --runInBand` (149 suites,
  887 tests), `npm run lint` (0 errors), `npm run build`, `vue-tsc` and `type-check:e2e` clean,
  Playwright 175 desktop / 13 mobile wireless (0 failures).
- [x] T015 — **DONE** via T020: keyboard focus and 390px/1280px verified in a browser.
- [ ] T016 — Distinct from T021: the final full-diff rule audit before merge. Open.

Add discovered work here and update the spec/plan when scope changes.
