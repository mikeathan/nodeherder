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

## Review round (2026-11-18) — the integration page, judged by using it

Spec [US-05](spec.md), plan §11, [ADR-005](adr-005-census-is-read-not-stored.md).
Ordered so the payload is honest before any surface renders it.

### Backend: one truth, and it is current

- [x] T035 — **DONE**: `RecordSourceActivity` now honours the same `Active` gate as
  `ObserveAdvertisements`. IDs: AC-24, D-13. Evidence: `TestDisabledGateStopsCounters`.
- [x] T036 — **DONE**: `publishLocal` records only the transient source error and clears any
  stored snapshot; `instanceDTO` resolves the resource on read through `resolveLocal`, with
  no identity allocator (`nil`), so nothing is reserved from a read transaction. IDs: AC-23,
  D-12, ADR-005. Evidence: `TestLocalProviderCensusIsReadOnEveryRead` (two reads after a
  hub-side change, and asserts no snapshot is persisted), `TestDisabledLocalProviderPublishesNoCensus`.
- [x] T037 — **DONE**: `Snapshot` gains `Observing`; `flush` refreshes the gate cache and
  clears sources, aggregates, classes and arrivals when it closes; the flag participates in
  `renderSignature` so the clearing broadcast is not suppressed as a duplicate. IDs: AC-24,
  D-20. Evidence: `TestClosingTheGateClearsTheCensus`, `TestSilentRoomIsNotAnAbsentObserver`.
- [x] T038 — **DONE**: `InstanceDTO.Continuous` from `Definition.Local`; both client
  derivations deleted. IDs: AC-25, D-14. Evidence:
  `TestLocalProviderReportsWhetherItIsContinuous` (single and listed), contract test for
  ingress validation.

### Frontend: the page stops offering controls that cannot act

- [x] T039 — **DONE**: `ConnectionSettings` takes `continuous` instead of deriving `local`
  from a definition that never arrives. Address row and schedule fieldset are absent for a
  continuous provider, with a sentence saying what it is. IDs: AC-25, D-14. Evidence:
  `ConnectionsPage.spec.ts` asserts `Change address` and `Automatic refresh` have count 0.
- [x] T040 — **DONE**: enable/disable is a `UiSettingRow` + `UiToggle` labelled "Observing",
  with help text naming what turning it off does. IDs: AC-25, D-15. Evidence: both settings
  specs drive the switch by role and check its state.
- [x] T041 — **DONE**: `providerName` resolves from the catalog, so the pane no longer shows
  the raw key `wireless_observation/hub`. IDs: D-16.
- [x] T042 — **DONE**: `Results.vue` shows the live count as the headline, explained hidden
  devices, source names, a link to the radar, and packet totals inside a `<details>`
  labelled "Packets" with the lifetime caveat. IDs: AC-26, D-17.
- [x] T043 — **DONE**: `HomeObservationCard` uses only defined theme tokens, is a
  whole-tile link with an overlay, has no inner primary button, and reaches "Not observing".
  `.is-observation` joins the full-width layout rule. IDs: AC-27, D-18, D-20. Evidence:
  the new Home spec measures the computed border colour.
- [x] T044 — **DONE**: `integrations/observation/help.ts` declares three topics: which
  sensors qualify, what the numbers mean, and what turning it off does. IDs: D-19.

### Guard

- [x] T045 — **DONE**: `src/__tests__/theme/tokens.test.ts`. Two assertions: the style layer
  defines every property it reads, and every Home component reads only defined properties.
  It found a fourth unreported instance — `HomePresenceCard` used `--nh-radius` (undefined)
  instead of `--nh-radius-l` — which is fixed here. Pre-existing offenders outside Home are
  reported, not fixed: the `components/wireless` tree has the same class of bug across
  `--accent-primary`, `--border-subtle`, `--text-muted` and friends, and the history charts
  rely on JS-set properties.

### Verification for this round (actual results, 2026-11-18)

- `backend/`: `go build ./...` PASS; `go test ./...` PASS (all packages);
  `go test -race ./internal/observations/... ./internal/integrations/...` PASS, no races.
- `frontend/`: `npm test -- --runInBand` PASS 152 suites / 996 tests; `npm run lint` PASS,
  0 errors (23 pre-existing warnings in untouched files); `npm run build` PASS;
  `npm run type-check:e2e` PASS.
- `frontend/` Playwright: `--project=desktop e2e/{home,ConnectionsPage,wireless}.spec.ts`
  PASS 72/72; `--project=mobile e2e/{home,ConnectionsPage}.spec.ts` PASS 40/40.
- Negative control: reinstating `border: 1px solid var(--border-subtle, #dee2e6)` made the
  new Home spec fail with `rgb(222,226,230)` against the theme's `rgb(207,228,238)`, so the
  regression is genuinely detected rather than merely present.
- **UNRUN**: a live EPL check of the three surfaces agreeing within one publish window.

## Review round 2026-11-18 (second) — navigation and one rendering bug

Spec [US-05](spec.md) AC-28…AC-30, plan §12.

- [x] T046 — **DONE**: `IntegrationWidget.destination` lets a capability point its board
  entry at its own live surface; `IntegrationEntry` falls back to the integration page and
  carries a settings action whenever it has been redirected. Whole-row target via a link
  overlay, with the actions above it. IDs: AC-28. Evidence: the board spec now clicks the
  entry and lands on `/wireless`, and a second spec proves the settings action still
  reaches the integration page.
- [x] T047 — **DONE**: the radar always offers "Integration settings", and renders an
  explicit notice with a link to Setup when `observing` is false. The link resolves the
  instance by provider key (`integrations/instanceForProviderKey`) and falls back to the
  catalog when none exists; the page reads the instance list on mount because it can be
  opened directly. IDs: AC-29. Evidence: `wireless.spec.ts` toggles the fixture off and
  asserts both the notice and the link.
- [x] T048 — **DONE**: the board row's device count was written to `value` while the entry
  renders `date`, so the count never appeared and the timestamp printed twice. IDs: AC-30.
  Evidence: the Go-fixture board spec asserts `1 device in range` on Results.
- [x] T049 — **DONE**: "23devices in range" was Vue condensing the whitespace between the
  count and its unit; the headline is now one string. IDs: AC-30.

### Verification for this round (actual results, 2026-11-18)

- `frontend/`: `npm test -- --runInBand` PASS 153 suites / 1001 tests; `npm run lint`
  PASS 0 errors; `npm run build` PASS; `npm run type-check:e2e` PASS.
- Playwright desktop PASS 74/74 (`ConnectionsPage`, `wireless`, `home`); mobile PASS 60/60
  (`ConnectionsPage`, `wireless`).
- `backend/` unchanged this round; `go build ./...` and `go test ./...` pass.

## Review round 2026-11-18 (third) — tense, and a request for a document that cannot exist

Spec [US-05](spec.md) AC-31…AC-32, plan §13.

- [x] T050 — **DONE**: "23 devices in range" directly above "No sensor is reporting" was a
  self-contradiction. Identities outlive their source by up to ten minutes by design
  (`UnregisterSource` keeps aggregates so a proxy restart does not blank the view), so the count
  describes the recent past while the source line describes now. Headline and status note are now
  computed together in `censusSummary`, which uses "heard earlier" when nothing is reporting and
  explains the drop-off. IDs: AC-31. Evidence: seven `censusSummary`/`censusReporting` tests,
  including the exact reported shape.
- [x] T051 — **DONE**: "Provider not found" on a local provider. `hasDocument` asked the catalog,
  and `provider?.origin !== 'local'` is true when the catalog has not loaded — so any route that
  skipped the catalog (a Home card, a deep link, or disabling) requested a definition that cannot
  exist. The instance's own `continuous` flag is now the authority, and the catalog is only a
  fallback for a backend that predates it. IDs: AC-32. Evidence: a browser test that deep-links the
  instance and then disables it, asserting no error and a rendered pane.

Also corrected: the Go e2e fixture described one identity on a source with **no delivery counters**,
which is the same contradictory shape the page is now right to call "heard earlier". It now carries
realistic counters, so the fixture cannot assert a state the product correctly refuses to describe.

### Verification for this round (actual results, 2026-11-18)

- `frontend/`: `npm test -- --runInBand` PASS 153 suites / 1007 tests; `npm run lint` PASS 0 errors;
  `npm run build` PASS; `npm run type-check:e2e` PASS.
- Playwright desktop PASS 75/75; mobile PASS 61/61.
- `backend/`: `go test ./internal/` PASS.

## Full-branch review 2026-11-18 — defects found by reading the whole diff

Asked for a full review of the branch for bugs, performance, dead code and rule violations.
Seven real defects, two of them severe and none of them visible in the UI until the right state
was reached.

| ID | Severity | Defect | Fix |
| --- | --- | --- | --- |
| D-25 | **critical** | `decodeAdvertisements` never set `Observation.ObservedAt`. The aggregator substitutes its own clock, so nothing looked wrong — but the recorder files evidence by that field, so every reading landed on day `0001-01-01` and hour 0. The accumulator never promoted, the write floor rejected the row, no summary was ever stored, and **classification could never work**: every device stayed "learning" forever with no error anywhere. | Stamped at decode, plus `TestDecodeAdvertisementsStampsObservedAt`. |
| D-26 | high | Closing the gate cleared `s.sources`, but only the live session can announce a descriptor and it does so once per connection (guarded by `announced`). Disabling then re-enabling left `sources` empty until the device reconnected — potentially hours — taking the filter facets, the diagnostic counters and the silent-source warning with it. | Kept sources on close; only observations are cleared. `TestReEnablingTheGateRestoresObservation`. |
| D-27 | medium | A local provider that failed once kept reporting the failure forever. `publishLocal` persists `LastError` and the DTO seeded its status from it; the successful read path never cleared it. The card stayed "attention", and the controls that would clear it are hidden for a continuous provider. | Cleared on a successful read. `TestRecoveredLocalProviderStopsReportingFailure`. |
| D-28 | medium | A **disabled** local provider fell through to projecting a stored census. ADR-005 claimed the read path "ignores it for local providers"; it did not. A record left by an earlier build showed "39 devices" for a radio nobody was reading. | A `local` case that never projects. `TestDisabledLocalProviderNeverProjectsAStoredCensus`. |
| D-29 | medium | `record()` accepted RSSI 0 as a real reading while every other consumer rejects it as unmeasurable. A single zero could make the median 0, render "0 dBm" and plot the device on top of the radar. | Skipped; the packet still counts. `TestZeroRSSIIsNotAZeroReading`. |
| D-30 | medium | `providerName` fell back to the raw key on any page reached directly, because only the board ever loads the catalog. Setup showed `wireless_observation/hub`. | The page fetches the catalog itself when absent. |
| D-31 | low | `censusSummary` hardcoded the plural in the hidden branch ("0 of 1 devices shown" beside "1 device is hidden"), and left "advertisements arriving but none decoding" with no note at all. | Noun follows the count; a note for the undecoded state. |

Dead code removed: `ObservationTable.vue` (orphaned when the cards view replaced it, zero
imports), `observation.Encode()` and `DailySummary.Var()` (never called),
`hasObservationSources` (gated a nav item that no longer exists), `Observation.Payload` (never
written), and twelve `export` keywords in `presentation.ts` that nothing imports. The stale
`ObservationTable` entry disappears from `components.d.ts` on the next build.

Considered and deliberately **not** changed:

- **Caching the gate.** `RecordSourceActivity` and `ObserveAdvertisements` consult a gate that
  opens a BoltDB read transaction, once per batch. I started adding a cache, and three existing
  tests correctly failed: they flip the gate and expect the next batch to respect it, and a cache
  makes an operator's "disable" take effect a tick late. A read transaction on an indexed key is
  microseconds against a batch arriving a few times a second, so the cache bought nothing and
  cost correctness. Reverted, with the reasoning recorded at `isActive`.
- `MediumWiFiProbe` stayed. It is a documented reservation for a future source, referenced by
  spec 064's extension rule, and removing it would make that path non-additive.
- Three unused icon keys (`copy`, `delay`, `google`) predate this branch and are unrelated.
- `SummaryRepository.SaveTransition` has no production caller, but the store method it wraps is
  tested and the classification spec requires transitions. Removing it would delete a capability
  rather than dead code.

### Verification for this round (actual results, 2026-11-18)

- `backend/`: `go build ./...` PASS; `go vet` on affected packages PASS; `go test ./...` PASS;
  `go test -race ./internal/observations/... ./internal/integrations/... ./internal/deviceproviders/...`
  PASS, no races.
- `frontend/`: `npm test -- --runInBand` PASS 153 suites / 1006 tests; `npm run lint` PASS
  0 errors; `npm run build` PASS.
- Playwright desktop PASS 75/75 (`ConnectionsPage`, `wireless`, `home`); mobile PASS 61/61.

### CI rehearsal (actual results, 2026-11-18)

Every job in `.github/workflows/ci.yml` was run locally before opening the PR.

| Job | Result |
| --- | --- |
| Workflow and Shell Checks | `actionlint` (v1.7.12, installed for this check) PASS; `bash -n` over every tracked `.sh` PASS; `shellcheck scripts/manage-stack.sh` PASS |
| Deployment Tooling | `sync-reporting-defaults --check` PASS; its own tests 4/4; `manage-stack` + `dev` tests 70/70; Compose configuration 15/15 |
| Backend Tests and Build | `go mod verify` PASS; `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build` PASS; `go test -count=1 ./...` PASS (all packages); patched ESPHome module `-race` PASS; `go vet ./...` PASS |
| Backend Race Detection | `go test -race -count=1 ./...` PASS, whole tree |
| Frontend Tests, Lint and Build | `npm run lint` 0 errors; `jest --ci --runInBand` PASS 153 suites / 1006 tests; `type-check:e2e` PASS; production build PASS |
| Frontend E2E | desktop shards 1-4 PASS (63, 42 + 9 skipped, 44, 50); mobile shards 1-4 PASS (63, 50 + 1 flaky, 44, 50) |
| Containers | backend and frontend images build; `-h` smoke check PASS; `nginx -t` PASS; built `index.html` present |

Two findings from the rehearsal, and one of them was a real defect:

- **Found and fixed a handle leak in this feature.** `useObservationViewer` started a one-second
  interval and only cleared it on component unmount. Unit tests call it outside any component, so
  each of the 16 call sites leaked an interval and **Jest could not exit** — `--detectOpenHandles`
  named it directly. The composable now exposes `stop()`, and the suite tears down every instance
  it creates. The full run now exits by itself in ~3s instead of hanging. The previous rounds hid
  this behind `--forceExit`, which is exactly the flag that stops you noticing it; CI would have
  hung on the 15-minute job timeout.
- **Three apparent failures were local environment, not code**, and were confirmed against `main`
  in a detached worktree: the proxy-peer tests read a `TRUSTED_PROXY_CIDRS` value from this
  machine's `backend/.env` (they pass 70/70 with it unset), and the Compose tests need
  `NODEHERDER_COMPOSE_BINARY` on macOS because the fake `DOCKER_CONFIG` defeats plugin discovery
  (they pass 15/15 with it set). Neither touches this branch — it changes zero files under
  `scripts/`.

One E2E failure was a flake: `devices.spec.ts › spatial mappings save and show only current
provider reports` failed once on a mobile shard and passed twice in isolation. This branch changes
no file in that spec or in the spatial code.
