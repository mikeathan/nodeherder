# Tasks: Frontend redesign and customisable design system

## Mobile entity repair — 2026-10-18

- [x] Reproduce delayed/empty-history sizing across entity families with isolated mocks.
- [x] AC-51/54: remove response-dependent shell sizing; cover all family states.
- [x] AC-52: continuous slider presentation; verify step, capture, cancel and confirmation.
- [x] AC-53: add truthful meter reading visual with missing/zero coverage.
- [x] Review, visually inspect, run checks, document evidence and commit scoped work.

### Visual completeness follow-up — AC-55

- [x] Restore decorative meter colours and shared chart guides/final-sample markers.
- [x] Correct state/event/air palettes and battery diagnostic presentation.
- [x] Add metadata-driven direct choices to entity views; preserve confirmation.
- [x] Verify both themes, mobile/desktop, review quality and commit.

### Drag stability repair — AC-56

- [x] Reproduce pending draft reset; compare native control with the old implementation.
- [x] Restore one native drag owner and retain unconfirmed pending drafts.
- [x] Verify brightness/temperature gestures, failure/cancel and quality; commit.

## Entity/chart implementation — 2026-10-17

Completed scope and checks: [2026-10-18 evidence](implementation-evidence-2026-10-18.md).

- [x] T300 — Freeze additive history/control contracts and capture baseline complexity;
  FR-20–24, NH-01/04, AC-50; done: amendment, producer fixtures, recorded metrics.
- [x] T301 — Bounded history HTTP producer, raw enum events, real errors and companion
  regressions; FR-23, AC-47; depends T300; done: Go contract/race tests.
- [x] T302 — Pure control conversions and history/energy/state calculations; FR-21/22/24,
  AC-45/46/48; depends T300; done: source-owned boundary/regression tests.
- [x] T303 — Shared request lifecycle/Vuex history, decoding, cancellation/timeout/cache;
  FR-22, AC-46/49; depends T301/302; done: failure/concurrency/reset coverage.
- [x] T304 — Shared numeric/state/event renderers and full-screen entity controls,
  full device history; FR-20–22, AC-44/45/46; depends T302/303; done: mock browser visuals.
- [x] T305 — Remove unused legacy paths, mock parity, quality cleanup and review;
  AC-49/50; depends T304; done: import scan, scores, resource/viewport regressions.
- [x] T306 — Required tests/lint/build/race, final acceptance/rule evidence and commit;
  NH-05, AC-44–50; depends T305; done: actual results and commit recorded.

## Entity view corrective pass (2026-10-18)

Owner review evidence, reproduction and expected behavior:
[plan](plan.md#entity-view-corrective-pass--2026-10-18), [spec](spec.md#corrective-pass--2026-10-18).
No contract or backend change; all tasks are frontend presentation/behavior corrections.

- [x] T307 — Make unknown/offline and preview states truthful in `UiRange` and
  `EntityAdjustment`: muted rail, hidden fill/grip until a real draft/reading, tested
  percent conversion, grip centred on the boundary, pointer release/cancel cleanup.
  IDs: FR-20/21, AC-44/45; paths: `components/ui/UiRange.vue`,
  `components/entity/EntityAdjustment.vue`, `domain/controls.ts`; done: `rangePercent`
  unit tests plus grip-geometry and gesture browser assertions.
- [x] T308 — Restore brightness/colour-temperature information: muted `Unknown`
  readout, warmth descriptor, `Warm`/`Cool` endpoint labels, visible power state, and
  labelled writable controls for read-only entities. IDs: FR-20/21, AC-44/45; paths:
  `EntityAdjustment.vue`, `EntityViewDialog.vue`, new `EntityPowerButton.vue`,
  `domain/exposes.ts`; done: browser assertions for Unknown/no-fill, warmth/endpoints,
  visible power/alarm labels.
- [x] T309 — Correct history truthfulness: no empty counter axes, explicit
  unavailable consumption, theme-token tooltip colours, compact/full selection by
  viewport. IDs: FR-22/24, AC-46/48; paths: `HistoryView.vue`, `NumericHistory.vue`,
  `options.ts`, `EntityViewDialog.vue`, new `composables/useMediaQuery.ts`; done:
  no-complete-interval browser test, desktop summary vs mobile compact assertions.
- [x] T310 — Fix focus/mobile composition: initial focus on the dialog surface, no
  slider outline on open, 44px circular mode/power targets, essential controls reachable
  at 320/390 and short landscape, desktop two-column use of the larger chart.
  IDs: FR-20, AC-44, FE-04; paths: `UiDialog.vue`, `EntityViewDialog.vue`,
  `EntityPowerButton.vue`; done: dialog-surface focus assertion, 320×568 control-reach
  test, 390/1120 screenshots, 320 header title hidden to avoid truncation.
- [x] T311 — Add source-owned regressions for the corrected behavior; IDs: AC-45/46,
  NH-05; paths: `src/__tests__/domain/controls.test.ts`,
  `src/__tests__/domain/exposes.test.ts`,
  `src/__tests__/domain/history/state.test.ts`,
  `src/__tests__/components/chart/history/options.test.ts`,
  `frontend/e2e/EntityViewDialog.spec.ts`; done: new assertions fail on the pre-fix
  behaviour and pass on the final code.
- [x] T313 — Normalise raw boolean history to live binary words, restore the
  `--nh-kind-switch` colour, and use the ordinary dialog width with an explicit empty
  state for read-only views without samples. IDs: FR-22, AC-46/48, NH-02/04; paths:
  `domain/exposes.ts`, `domain/history/state.ts`, `assets/styles/tokens.css`,
  `HistoryView.vue`, `EntityViewDialog.vue`; done: unit assertions for boolean/word
  history, state-timeline screenshot, empty-history browser test.
- [x] T314 — Restore compact device-row sliders: 1.75rem visual rail, 44px touch
  target, kind colour, and mired/Kelvin inversion matching the entity view.
  IDs: FR-21, AC-45, FE-04; paths: `components/entity/ExposeControl.vue`,
  `components/ui/UiRange.vue`; done: 28px rails at 112–122px, affected browser specs.
- [x] T315 — Alert-family action rows, latest-recorded fallback, known-only other
  readings and dense-timeline threshold. IDs: FR-20/22, AC-44/46, NH-06; paths:
  `composables/useEntityDialog.ts`, `EntityViewDialog.vue`, `StateHistory.vue`;
  done: smoke silence browser test, latest-recorded browser test, density screenshot.
- [x] T316 — Make energy history compact at narrow widths, keep its summary visible,
  and replace zero-only interval axes with an explicit notice. IDs: FR-22/24, AC-46/48;
  paths: `NumericHistory.vue`, `HistoryView.vue`, `EntityViewDialog.vue`; done:
  compact energy browser checks and no-complete-interval regression.
- [x] T317 — Historical, superseded by AC-51/52 above: collapse unavailable adjustable rails to a range line and use a
  content-height mobile sheet when there is no usable control/history. IDs: FR-20/21,
  AC-44/45, FE-04; paths: `EntityAdjustment.vue`, `EntityViewDialog.vue`; done:
  offline-brightness 597px sheet check and entity-dialog browser suite.
- [x] T312 — Run fresh isolated mock and exact checks, record acceptance/rule evidence
  and update the evidence document; IDs: AC-44–50, NH-01/05; depends T307–T317; done:
  118 Playwright tests, 42 Jest suites/278 tests, lint/build, quality max C=8/CRAP=30;
  evidence recorded.

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: Main redesign merged; scoped backend repair implemented; follow-up open

## Current status (2026-10-12)

Main redesign merged as #54 (`fcdebca2`), implemented on
`claude/eloquent-lovelace-rus03g`. Active repair on `investigate/ui-device-updates`:
T206 backend portion implemented; frontend rollback/refusal/disconnect and browser latency/
load evidence remain open. Broader entity/history work is deferred. Findings, measured
quality, exact checks and limitations live in the [plan](plan.md#device-update-repair-2026-10-12);
timing/compatibility live in [ADR-005](adr-005-live-device-updates.md).

- [x] T206 backend — Separate live reports from sampled metrics; preserve fresh automation,
  writable same-value confirmations, zero and repeated events; move routine chatter to trace.
- [x] Verify backend regressions, ordered MQTT/dial scenarios, full tests/build/vet and
  affected services/controllers/automations/store race checks on Go 1.24.8.
- [ ] T201 remaining — Branch evidence, Vue source mapping/mounting, TS/Vue/mock tooling
  and CI threshold gate. Generated reports remain temporary, not committed snapshots.
- [ ] T203/T204 remaining — Live effective settings/metadata and browser/ingress timing;
  controlled reproductions are complete, but deployment-specific uncertainties remain.
- [ ] T206 remaining — Frontend numeric rollback/refusal/disconnect and confirmed feedback.
- [ ] T213/T214 — Browser/load/accessibility evidence and final acceptance/rule review.

These partial milestones do not complete the parent T201/T203/T204/T206 tasks, the wider
follow-up or human review. Detailed dependencies and acceptance remain below.

## Phase 0 — Documentation and samples (authorised 2026-10-09)

- [x] T001 — Document current code paths, contracts and gaps in `current-state.md`;
  gate: NH-01/NH-04; done: file committed.
- [x] T002 — Draft spec, plan, design system, device support, automation editor docs and
  ADR-001; gate: NH-01; done: files committed, open questions listed.
- [x] T003 — Build 5 static design samples (plus standalone HTML files with an All pages view)
  covering every FR-01 screen with customiser; IDs: AC-01, AC-02; done: rendered in headless Chromium for every design ×
  screen (see plan traceability).
- [x] T003b — Round 2 after owner feedback (2026-10-10): keep Hearth, add four structurally
  different designs (`design-directions.md`, prototype layout modules); done: verified in Chromium.
- [x] T003c — Round 3 (2026-10-10): Hearth + Panel merge with logo, shared themes, activity
  switch and phone editor layout; done: verified in Chromium (desktop + 375 px).
- [x] T004 — Owner review: Hearth + Panel chosen 2026-10-10; done: Q-01 resolved in spec.


## Original draft tasks (superseded)

T010…016, T020…022, T030, T040…042, T050 and T060 were superseded by the
T100…110 implementation sequence below. T061's manual viewport/keyboard/screen-reader/
offline checks carry into T213; T062's final rule/map review carries into T214.
The draft is not a second active checklist; original acceptance remains in the spec/plan.

## Implementation (authorised 2026-10-10)

- [x] T100 — Characterization Jest tests for WS-emitting store actions, WS routing, auth
  service; gate: NH-05; done: tests pass on unchanged code.
- [x] T101 — Playwright harness (`frontend/e2e`, `.env.e2e`, `npm run test:e2e`) with auth,
  home, automations, device flows passing on the current UI; ADR-002; NFR-06.
- [x] T102 — Backend `DashboardGroup.order` + Go test; mock hub pass-through; FR-09.
- [x] T103 — `src/domain/*`, `src/theme/*` with Jest tests first; FR-02, FR-04, FR-05, FR-07…09, FR-12.
- [x] T104 — Token CSS, themes, PrimeVue preset, fonts, UI primitives; remove legacy theme.
- [x] T105 — Shell (sidebar, header, logo, connection, permit join), sign-in page; FR-11, FR-13.
- [x] T106 — Overview with activity switch; Home with drag/keyboard reorder; devices; device
  list; device page; FR-01, FR-09, FR-12.
- [x] T107 — Panel mode; FR-10.
- [x] T108 — Automations list and single-page editor; FR-06, AC-10…15.
- [ ] T109 — Partial: Settings/Appearance and Console restyling are implemented.
  Assistant and legacy dialog layouts still need the T212 follow-up.
- [x] T110 — Update e2e page objects; full e2e at two viewports; CI job; docs.
- [ ] T111 — Final review and rule/acceptance evidence, carried into T214; human
  review and proposed constitution/ADR approvals remain outstanding.

## Follow-up preparation (investigation checkpoint, 2026-10-11)

Investigation on `investigate/ui-device-updates` reproduced default sampling suppressing
live broadcasts, slider confirmation timeouts and pre-match dispatch logs. Mouse/touch/
keyboard reproductions passed; backend characterizations were converted to passing repair
regressions on 2026-10-12. See [findings](plan.md#investigation-findings-2026-10-11).
Live effective-configuration/timing evidence remains open; wider work is deferred.

- [ ] T200 — Inventory actual consumers and touched functions in entity/chart, command,
  automation, metrics store/service, mock JS and conditional Go paths; IDs: AC-38, NH-04;
  done: versioned file/function inventory against `fcdebca2`, exclusions justified.
- [ ] T201 — Establish pinned coverage/complexity/CRAP runner and CI design; include TS,
  Vue script functions/callbacks and mock JS, and Go if changed; prove SFC source mapping
  and Jest mounting compatibility, record dependency cost in ADR-004; IDs: AC-38, NFR-07;
  depends: T200; done: reproducible baseline, exact commands/tool versions in
  the plan, measured fixture proving gates fail correctly. No invented scores.
- [ ] T202 — Capture backend HTTP raw numeric/binary/enum/error/empty/limit/boundary fixtures
  from `internal/metrics/{query,services,storage}` and API; specify mock parity and verify
  API/MCP consumers; IDs: FR-16, AC-33…35, BE-04/05; depends: T200; done: checked producer/
  consumer types, required failure fixes and backend scan-cost assessment recorded.
- [ ] T203 — Controlled delayed illuminance replay for Living room presence sensor; inspect
  effective category/device/expose sampling, metadata and UI latency; include final update
  followed by silence and unchanged next reports; paths: `device_lifetime.go`, settings
  cache/config, controllers/WS, hub store/entity views; IDs: FR-18, AC-40/41, BE-02/03;
  depends: T200; done: root-cause/timing evidence or documented sensor-only delay, failing
  regression for application defect; timing ADR before any sampling/broadcast change.
- [ ] T204 — Reproduce brightness slider failure with mouse/touch/keyboard and compare
  target/expose/type/range/MQTT payload against working dial fixtures; correlate supplied
  `0xa4c13894070052fc` dispatch logs with matched triggers/actions/echoes; paths: entity
  controls/commands, hub/MQTT/controllers/automations; IDs: FR-19, AC-42/43, NH-02;
  depends: T200; done: diagnosed cause or bounded unresolved evidence, failing regressions
  for demonstrated defects, no assumption that dispatch logs prove action execution.
- [ ] T205 — Mocked entity/full-history/mini-chart desktop and phone prototype; use existing
  tokens, show numeric/state/event and empty/error/partial/offline states; IDs: AC-28…36;
  depends: T200; done: reviewable artifacts and recorded layout choices, no real-device calls.

## Follow-up implementation sequence (requires preparation gates)

Every task below also maps to AC-38/NFR-07: all its touched functions meet the per-function
gate with before/after evidence, meaningful failure coverage and tracked replacements.
Do not defer refactoring to the final task or omit store/composable/Vue/mock/Go functions.

- [ ] T206 — Fix diagnosed slider/illuminance defects and shared command rollback/refusal/
  confirmation; preserve dial payloads, repeated events and automation order; paths from
  T203/204 plus `useDeviceCommand`, `ExposeControl`; IDs: AC-29/40…43, FR-18/19;
  depends: T201,T203,T204 and any timing ADR; done: failing regressions pass, controlled
  latency/deadline and affected Go race results recorded; no unmeasured global debounce change.
- [ ] T207 — Fix disconnected automation/layout persistence, preserve drafts and explain
  failure, share manual-run lock, use API helper, validate draft/schedule times and add
  Europe/London schedule timeline; paths: automation/dashboard store/composables, editor,
  `automation-trigger.service.ts`; IDs: AC-37, FR-17; depends: T201; done: offline/error/
  duplicate Run/schedule/invalid-save regression and browser evidence.
- [ ] T208 — Shared typed entity presentation/capability and detail model; migrate Home,
  device rows/cards/controls, Panel and legacy entity dialog to one model/command path;
  consolidate formatter/unit/binary/icon semantics; IDs: FR-14, AC-28…30; depends:
  T201,T205,T206; done: all surfaces agree, unknown/missing/zero/disabled/offline fixtures,
  exact expose commands and keyboard/focus evidence.
- [x] T209 — **Superseded by [029](../029-history-and-mobile-ux/plan.md)** (existing baseline plus its implementation and verification tasks); History query correctness and service/store lifecycle: actual HTTP decoder,
  generation/keyed status, abort/timeout/429/empty/error handling, budgets/cache cleanup;
  update mock HTTP route and required backend error handling; paths: history service/
  contract, metrics store, `useHistoryQuery`, Go query service/API; IDs: FR-16, AC-33…35,
  NFR-08; depends: T201,T202; done: actual producer fixtures pass, obsolete replies cannot
  overwrite views, storage failures are observable; compatible backend checks pass.
- [x] T210 — **Superseded by [029](../029-history-and-mobile-ux/plan.md)** (existing baseline plus its implementation and verification tasks); Pure history domain: timezone/range boundaries, numeric and state/event
  normalization, gap/unknown policies, labelled stats and extrema-preserving decimation;
  paths: `src/domain/history/`, entity metadata; IDs: AC-31…33, NFR-08; depends: T201,T202;
  done: DST/zero/sparse/duplicate/out-of-order/partial/extrema/known-duration fixtures pass.
- [x] T211 — **Superseded by [029](../029-history-and-mobile-ux/plan.md)** (existing baseline plus its implementation and verification tasks); Completely rebuild full history and shared compact variants: toolbar,
  inspection/zoom/reset, state timeline/event list, theme-reactive renderer, accessible
  sample table, all query states; migrate entity dialog and `DeviceMetrics.vue`; IDs:
  FR-15, AC-30…36; depends: T205,T208,T209,T210; done: full/mini semantic equality and
  keyboard/touch/viewport evidence, no old chart implementation quietly retained in dialogs.
- [ ] T212 — History-dependent work moves to [029](../029-history-and-mobile-ux/plan.md); remaining Assistant/dialog layout work and measured palette/contrast fixes;
  remove unused legacy entity/chart/formatter paths only after all imports/consumers migrate;
  IDs: AC-36/39, FR-17, T109; depends: T208,T211; done: visual/focus checks, palette/accents
  AA assertions, import scan and build; no replacement semantics hidden by compatibility aliases.

## Follow-up verification and handoff

- [ ] T213 — Extend browser flows for edited automation payload, slider scale/gestures/
  confirmation/timeout, panel toggles, history range races/failure/empty/partial and reconnect;
  run frontend suites/lint/build/e2e, relevant Go test/build/vet/race, pinned quality gates,
  p95 history/render budget, 100 tiles × 10 msg/s, gzip comparison; IDs: AC-28…43,
  NFR-01…04/06…08, T061; depends: T206…212; done: exact environment/commands and actual
  passed/failed/unrun results, plus 360/768/1280/1920 keyboard/screen-reader/mock screenshots.
- [ ] T214 — Update plan, task status, quality/acceptance evidence, docs/repository
  map; recheck every NH/FE and applicable BE rule and remaining original requirements;
  IDs: SC-03, NH-01/04/05, T111; depends: T213; done: evidence-linked completion with
  outstanding risks and actual human-review status, no invented approval or ratification.



## Merged redesign checkpoint (historical)

T100…110 delivered wire-contract/Jest tests, desktop/mobile browser harness, dashboard
ordering, domain/theme layers, shared UI/composables, shell/login/Overview/Home/device
screens, Panel mode and the single-page automation editor. Area order saves, unchanged
automation round trips, activity switching and 11-screen overflow checks were covered.
Full UI details and acceptance are retained in the spec/plan.

Final pass removed legacy UI, popup OAuth and Roboto; adopted ADR-003 fonts; shortened
tile labels with descriptions as tooltips; hid absent last-seen times; and animated mobile
drawer closure. The intermediate checkpoint had 135 Jest tests; final verification had
136 Jest tests and 26 browser tests at 1280×800/390×844, lint with no errors, frontend
build and Go tests passing. Review on 2026-10-11 recorded seven existing lint warnings.
Manual keyboard/screen-reader/offline checks at 360/768/1920, Assistant/dialog layouts,
actual automation-edit/panel-command/reconnect browser coverage and human review remain
follow-up work (T109/T111/T213). These historical checks do not prove current repair latency.

## Completed artifact/document cleanup (2026-10-12)

- [x] Remove completed-redesign prototype HTML/assets/generators and raw quality JSON/checker.
- [x] Consolidate implementation progress here and investigation/quality evidence in the plan.
- [x] Remove duplicate status/investigation/quality documents and update their references.
- [x] Verify local links/references and `git diff --check`; no application suites rerun
  for this documentation-only cleanup.
