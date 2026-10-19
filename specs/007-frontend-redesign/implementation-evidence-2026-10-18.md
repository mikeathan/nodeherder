# Entity/history implementation evidence

Scope: FR-20–24 / AC-44–50 on `feat/entity-history-redesign`, baseline
`771d2a9b2f64529d5a927b3d7e1d2ff2f03f5f07`. The owner's implementation/review/commit
instruction supersedes the earlier investigation-only checkpoint. Constitutions
remain proposed; no ratification, maintainer review or exception is invented.

## Visual completeness review — 2026-10-19

The owner identified the missing agreed multicolour meter and requested a comparison
of all remaining families with the accepted references (recovered temporarily from
git history; disposable PNGs remain removed from the repository). AC-55 records scope.

Completed corrections:

- Energy: a full-opacity green/yellow/coral SVG gradient, with per-instance IDs and
  existing theme tokens. The cumulative total stays an exact labelled reading;
  the decorative arc adds no maximum or safety threshold.
- Numeric history: a contrasting marker on the final known sample, with sufficient
  clipping allowance at range endpoints. Missing tails stay missing. Start/middle/end
  time labels replace irregular automatic millisecond ticks. Existing dashed guides
  were verified in the installed Chart.js source and preserved, rather than replaced.
- Shared expose palette: teal air quality, purple action/event charts, amber open and
  teal closed contact, red alert/low-battery and teal clear/dry state. Badges, timelines,
  legends and recent-report dots use the same registry; unknown remains distinct.
- Diagnostics: battery/signal badges, a known-only percentage bar, and relevant
  available link/battery readings. Null never becomes zero; actual zero keeps its meter.
- Settings: up to five metadata options become labelled direct buttons in entity
  views. Rows and longer lists retain the shared select. The selected button remains
  confirmed state until the device update arrives; pending disables duplicate sends.
- Cleanup: removed the unused history colour variable and retained one palette owner.
  No new dependency, transport, retention, resource subscription or device capability.

Verification:

- Chromium desktop/mobile entity/history suite: 59 pass, nine intentional desktop
  skips for mobile-only scenarios. The new meter-colour test initially read before
  the dialog existed; adding visibility synchronisation corrected this test race.
- Final marker/time-label change: four additional device-history browser checks pass.
  Existing gap preservation, repeated-event, native-command, cancellation, shell and
  confirmation checks remain passing in these affected suites.
- Jest: 42 suites / 280 tests pass; lint has zero errors and seven existing script
  warnings; production type check/build pass. Quality evidence is regenerated from
  fresh coverage with no complexity/CRAP violations (maximum C=8 / CRAP=30).
- Visually inspected both-theme meter and battery screenshots, plus the final numeric
  history screenshot showing the complete marker and three labels. Logs/screenshots
  and the temporary browser config are outside the repository.
- Firefox 142.0.1 / Playwright build 1495 was installed to match the locked runner.
  Its browser launch failed before application tests: macOS rejected a plugin-container
  sandbox extension and the SWGL compositor could not map its framebuffer. A bounded
  15-second retry reproduced the launch failure. Firefox rendering is **unverified**;
  the initial eight launch failures and interrupted unbounded retry are not application
  test failures or passing evidence. No physical phone/assistive-tech session was run.
- The full 135-pass browser run above belongs to the preceding repair; it was not
  unnecessarily repeated for these scoped visual changes. No Go code changed, so
  backend checks were not rerun. Relative-link and diff-whitespace checks pass.

Review against NH-01–06/FE-01–05: scoped acceptance and shared presentation ownership;
confirmed values and native commands retained; no ingress, secrets, backend or storage
change; theme/accessibility/empty/zero coverage and real verification recorded. Existing
resource ownership is unchanged. No exceptions or ratification are claimed.

## Mobile repair review — 2026-10-18

Owner screenshots exposed gaps in the prior corrective pass. AC-51–54 in the
[spec](spec.md#mobile-entity-repair--2026-10-18) supersede history-dependent sheets.

| Finding | Cause and final correction |
| --- | --- |
| Views open small, expand later, or remain half-height | `fullscreenMobile` and desktop width depended on loaded samples/availability. Every entity now requests the same full-screen shell from first render through loading, empty, error and offline states. The 899px breakpoint matches the entity layout, including 768px tablets and 844px landscape. Desktop width depends only on the control type. |
| Brightness/temperature dragging feels stepped and disconnected | Painted grip and native-thumb travel used separate geometry; device steps drove visual motion and pending state faded the rail. Pointer capture now uses the visible rail geometry, tracks a continuous local percentage, and snaps only the draft/command to native steps. The fill scales instead of repainting a moving hard-stop gradient. Confirmation preserves the rail and opacity; mode/status rows reserve their space. |
| Controls extend below a short phone screen | Slider height ignored the space needed by readout/mode controls. Its height now accounts for viewport space; the 320×568 control row stays visible. Very short landscape retains internal scrolling and a reachable close button. |
| Activity/alert views lack the accepted visual hierarchy | `EntityReadout` restores state/event/alert badges and keeps environmental views typography-first. Writable mobile views use compact history and omit unrelated readings. |
| Energy has no requested meter visual | `EntityMeter` frames the exact cumulative reading with a compact semicircular SVG arc. The arc is decorative: there is no invented maximum, percentage or safety threshold. Consumption history and missing/zero coverage stay separate. |
| Tests passed despite incorrect layouts | Settled screenshots missed initial geometry; an empty-history test explicitly endorsed compact sheets. Tests now hold the response until first-frame geometry is checked for nine read-only families, then verify unchanged bounds after empty results. |

Final checks (Node 24.21.0; isolated mock hub and Chromium only):

- Jest coverage: 42 suites / 278 tests pass.
- Clean-server full Playwright run: 135 pass; nine intentional desktop skips for
  mobile-only scenarios. Includes both 50-cycle lifecycle checks, confirmation,
  refusal/timeout/cancellation, keyboard, mouse/touch and viewport checks.
- Additional final Playwright run: six pass, including drag-frame alignment with
  reduced motion disabled and meter unknown versus measured zero. More than 12
  distinct rendered drag positions are checked; grip/fill geometry error stays below 2px.
- Frontend lint: zero errors, seven existing script warnings. Type check/build pass.
  Existing duplicate `DeviceCard` auto-import warning remains.
- Mock history contract: one test passes. Changed-function quality: 61 source reports,
  317 touched functions, maximum complexity 8 / CRAP 30, no failures.
- First full run on the reused hot-reloaded Vite process had four sign-in interceptor
  failures: timestamped module URLs bypassed its exact lazy-route match. Fresh isolated
  servers resolved these; the clean full run passes. No application auth code changed.
- Inspected generated mock screenshots for brightness, temperature/humidity, presence,
  energy, battery, smoke and alarm. Final viewport checks include 320×568, 390×844,
  768×1024, 844×390 and 1280×800. The earlier 1920px inspection is historical.
- Removed 12 disposable PNGs, their preview index, and superseded initial investigation/
  readiness reports at the owner's request. Kept contracts, acceptance and final evidence;
  relative file-link and diff-whitespace checks pass. Screenshots/logs stay outside git.

Review: state/services/transport ownership is unchanged; pointer presentation is a
local draft, not confirmed device state. Native steps, one command per completed
gesture, cancellation and confirmation remain shared. New display components own no
requests, timers or global listeners; pointer capture ends on release/cancel/removal.
No dependency, backend, schema, retention or device ingestion change. NH-01–06 and
FE-01–05 checked for this scope; no exceptions or ratification claimed. Go checks were
not rerun for this frontend/document-only repair. Physical iOS/Android and assistive
technology sessions remain unrun; browser motion checks do not establish physical feel.

## Final ownership and review

- `domain/controls.ts` owns native numeric bounds, steps and display conversion;
  `domain/exposes.ts` owns capability/labels/value formatting. `useExposeEdit` owns
  local drafts and delegates to `useDeviceCommand`; hub Vuex remains confirmed truth.
  Both device rows and full-screen controls use `UiRange`, including cancellation.
- `useEntityDialog` selects only relevant supported controls. Mobile entity dialogs
  fill the viewport; desktop adjustable views put history beside the control.
  Read-only views include compact useful summaries, real units and sample inspection.
- `domain/history/*` owns range, numeric reduction, state duration, event count and
  counter consumption rules. Shared `HistoryView`, `NumericHistory`, `StateHistory`
  and one options adapter replace the separate legacy mini/full chart implementations.
- `contracts/history.ts` validates HTTP ingress. `history.service.ts` owns a bounded
  deduplicated request queue, timeout and reference-counted cancellation. Vuex owns
  bounded results; `useHistoryQuery` rejects stale view/session completions. Logout
  clears histories and command timers. No per-tile polling or history persistence.
- Go handlers validate/adapt, the query service resolves devices and owns scan/time
  budgets, domain collectors preserve raw reports, and the repository owns storage.
  [Wire contract](history-contract.md) is additive; legacy HTTP/WS remain compatible.
  Cache snapshots and an actual cache boundary prevent duplicate/lost split samples;
  an empty cache after reopening uses persisted recent data.

The separate cleanup pass removed unused chart/slider implementations, factories,
helpers and their orphaned tests after import scans. Comments explain ownership,
draft/confirmed separation, unit inversion, gaps, counters and lifecycle guards.
Heavy chart code now loads asynchronously instead of being globally registered.
Chart.js 4.5.1 replaces ApexCharts/vue3-apexcharts; its small colour utility is the
only transitive addition. There is no storage schema/retention change. The library
change and measured reason are recorded in ADR-004.

## Acceptance evidence

| Acceptance | Evidence |
| --- | --- |
| AC-44 layout/design | EntityViewDialog/DeviceMetrics browser specs; full-screen 390px, no overflow at 320/360/768/1280/1920 and short 500px landscape; mock screenshots visually inspected during implementation (disposable image artifacts subsequently removed at the owner’s request) |
| AC-45 controls | Native 184/254 brightness displays 72%; Kelvin display/mired command; mouse/touch one-send gesture and cancellation; keyboard, confirmation, timeout/refusal/disconnect, missing-reading/zero tests |
| AC-46 shared history | Shared renderer/domain/service tests; full-history numeric/contact/repeated event switching, custom invalid/valid dates, explicit error/refresh, generation/session invalidation |
| AC-47 producer | API raw repeated enum/truncation/400/500 tests, domain decode tests, service empty/failure tests, repository range/reopen/scan/cancellation regressions; existing API/WS suites remain passing |
| AC-48 energy | Source-owned energy tests: endpoints, reset, explicit gap, no cross-bucket allocation, missing baseline and measured zero; meter total separate from labelled observed consumption |
| AC-49 resources | Scheduler tests: one active, bounded queue/cache, shared cancellation/deadline/reset; 50 completed chart open/close cycles on desktop/mobile; heap/DOM evidence below |
| AC-50 quality | Pinned per-function TS/Vue/mock and Go checks below; no changed function exceeds C=10 or CRAP=30; no coverage exclusion added |

## Checks and runner

Darwin arm64; Node 24.21.0, Go 1.26.1; Playwright 1.56.1 / bundled Chromium
141.0.7390.37. Browser services are isolated mock hub 4110 and frontend 4100.
No household device was commanded by these tests.

- Frontend `npm test -- --runInBand --coverage`: 42 suites / 273 tests pass.
- Frontend `npm run lint`: pass, zero errors; seven existing warnings in unrelated
  `scripts/get-version.cjs` and `scripts/start.mjs`.
- Frontend `npm run build`: type check and production build pass. Main bundle is
  about 488kB / 144kB gzip after lazy chart loading; Chart.js is a separate ~208kB /
  71kB gzip chunk, versus the earlier ~578kB / 157kB Apex chunk. The build no longer
  reports an oversized chunk.
  Existing duplicate `DeviceCard` auto-import warning remains outside this scope.
- Frontend `npm run test:e2e`: all 104 desktop/mobile browser tests pass.
- `node --test tools/server/history.test.mjs`: mock contract test passes.
- Backend `go test ./...`, `go vet ./...`, `go build ./...`: pass.
- Backend `go test -race -coverprofile=/private/tmp/nodeherder-history-go.cover
  ./internal/metrics/... ./internal/api`: pass; temporary stores/mocks only.

Earlier iteration failures were corrected: explicit select names in history, timeout
rollback expectations, and a range-test fixture using an unsupported Playwright fill
operation. A prior auth handoff failure did not recur in the completed full suite.
Local-port sandbox restrictions required the tests to run with local binding enabled.

## Quality measurement

Frontend command, from repo root after fresh coverage:

```sh
node frontend/tools/quality/history-quality.cjs frontend/coverage/coverage-final.json
```

TypeScript 5.9.2 AST, source-line-preserving Vue script extraction, separate nested
functions/callbacks; 243 changed/new functions. Maximum C=9, maximum CRAP=30,
zero violations. 134 functions have Istanbul statement evidence; 109 use explicitly
labelled zero-coverage upper bounds. Removed predecessors remain in the baseline;
baseline upper bounds are not claimed as historical measured scores. See the
[runner documentation](../../frontend/tools/quality/README.md) for exact conventions.

Touched branch coverage includes history service 100%, control conversions 97.61%,
events 100%, ranges 100%, state 95.23%, numeric 94.44%, contract 89.18%, energy
89.47%, useExposeEdit 86.36%, useHistoryQuery 88.88%. Vue rendering is verified in
Playwright; generated templates are not included in the script complexity report.

Go scorer `github.com/hbaldwin98/crap/cmd/crap@8eea294df1123778a77ecb826f0b16d95b65768d`
(reported version `v0.0.0-20260908235400-8eea294df112`), from `backend/`:

```sh
crap -format json -coverage /private/tmp/nodeherder-history-go.cover \
  -diff-base 771d2a9b2f64529d5a927b3d7e1d2ff2f03f5f07 internal/metrics internal/api
```

22 changed functions; maximum C=9 / CRAP=9, zero violations. Coverage is mapped
from the affected race suite; whole-package percentages are not substitutes for
function scores. Reports/coverage are reproducible generated artifacts kept out of git.

## Entity view corrective pass — 2026-10-18

Historical checkpoint: the owner subsequently rejected response-dependent mobile
sheets. The mobile repair below supersedes that sizing decision and its evidence.

Owner review of `7ccdcfcb` found the implemented views did not match the accepted
direction. Reproduction and corrections are recorded in the
[plan](plan.md#entity-view-corrective-pass--2026-10-18), [spec](spec.md#corrective-pass--2026-10-18)
and [tasks](tasks.md#entity-view-corrective-pass-2026-10-18). No HTTP/WS contract,
backend, ingestion, storage or dependency changed.

Observed and fixed: a missing/offline adjustable reading rendered a large `—` plus a
minimum-position, pale rail (the reported "handle above the slider"); the grip was
offset from the fill boundary; the dialog autofocused the range and painted a focus
ring; power/alarm controls lacked visible labels/state; colour temperature omitted
warmth and warm/cool endpoint text; adjustable desktop views used the compact chart;
counter history with no complete interval drew empty axes with `—`; the mock history
fallback collapsed expose-specific `ON`/`OFF` words to booleans; raw boolean HTTP
history rendered as `Unknown` beside a confirmed `On`; the `--nh-kind-switch` colour
was undefined; read-only views without samples used an oversized blank column;
device-card row sliders expanded to a full-height 3rem rail across the row; alert views
did not surface their silence action; missing current readings showed only `Unknown`
despite recorded samples; "Other readings" listed empty values; and busy presence
timelines rendered hundreds of thin buttons; Energy spent most of a narrow view on an
empty 235px chart; and an offline adjustable entity still filled the phone with a
full-height disabled slider and blank space.

Corrective evidence:

| Acceptance | Evidence |
| --- | --- |
| AC-44 layout/design | 320×568, 390×844, 768×1024, 1280×800 and 1920×1080 dialogs: no horizontal overflow; primary controls visible within the body at 320/390/768/1280/1920; desktop (≥900) uses two columns and the full history variant, mobile keeps compact history. Device cards at 1500/390/320 have no horizontal overflow and compact 28px rails; the 320px header hides its duplicate page title instead of truncating to `De…`. Short landscape (667×375, 844×390) uses the existing scrollable sheet with the header/close fixed; controls are reachable by scrolling, not permanently clipped. |
| AC-45 controls | `UiRange`/`EntityAdjustment` tests and browser spec: unknown shows `Unknown` with a muted rail and no grip; grip centre is within 2px of the 72.44% fill boundary; one-send gesture/cancel unchanged; dialog surface, not the range, receives initial focus; energy/read-only writable controls show `Power` + On/Off/Unknown and `Alarm` labels. Device-row sliders are 1.75rem high (112–120px wide in the card test) with a 44px hit area, kind colour and mired/Kelvin direction matching the entity view. An unavailable adjustable entity shows `Unknown`/range in a content-height 597px sheet with no dead rail. |
| AC-46 shared history | Desktop brightness shows `Range`/`Sample average`; mobile uses compact history. Counter history with one sample shows `No complete counter interval in this period.`, `Not enough data`, `0 / 12` and no canvas. Raw boolean state history renders `On`/`Off`, 120 min duration, 1 transition and 100% coverage instead of `Unknown`; read-only entities without samples use the ordinary dialog width and the explicit empty-history block. Missing current readings show `Latest recorded <value> · <time>`; other readings list known values only; >96 state segments use the labelled density view. Tooltip colours come from `--nh-surface`/`--nh-text`/`--nh-border-strong`. |
| AC-48 energy | Observed consumption remains separate from the meter total; unavailable consumption is labelled instead of a bare dash; empty/zero-only interval axes are not rendered. At 597px the meter total, Power/Child lock and consumption summary stay primary with a 105px chart. Mock fallback preserves expose on/off words. |

Checks, Darwin arm64, Node v26.0.0 (workspace `.nvmrc` is 24; actual runner reported),
TypeScript 5.9.2, Playwright 1.56.1/Chromium:

- `npm test -- --runInBand`: 42 suites / 278 tests pass.
- `npm run lint`: 0 errors, 7 pre-existing warnings in `scripts/get-version.cjs` and
  `scripts/start.mjs`.
- `npm run build`: type check and production build pass.
- `node --test tools/server/history.test.mjs`: pass, including expose-specific binary
  on/off words.
- Fresh isolated mock hub 4110 and frontend 4100 (`npm run test:e2e`): 118 passed on
  the final code. A focused `EntityViewDialog.spec.ts` run also passed 24/24. Isolated
  4103/4113 runs use the new
  `E2E_EXTERNAL_SERVERS`/`E2E_BASE_URL`/`E2E_HUB_API` options. No household device was
  contacted; the manual state/empty-state/card/smoke/energy/offline screenshots used
  isolated mocks.
- Quality runner after fresh coverage: 58 reports, 312 touched functions, max C=8,
  max CRAP=30, zero failures.
- Screenshots inspected in dark and light themes at 320/390/1120/1500 widths for
  unknown/offline brightness, known brightness with full/compact history, colour
  temperature, energy, alarm, contact, temperature, battery, smoke, silence, state,
  empty illuminance/presence, dense presence, compact energy at 597px, the offline
  brightness content sheet, and the device-card sliders. The earlier
  empty illuminance/energy screenshots were caused by a stale mock process, not
  production code.

Limits: Chromium desktop/mobile emulation only; physical iOS/Android and assistive
technology sessions were not run. Short landscape requires scrolling to the control
row; it does not overflow horizontally. No backend or Go checks were run because this
correction changed no Go code.

## Performance and lifecycle

The 10,000-point normalization/statistics/extrema-reduction benchmark includes
runtime decoding, sorting and range filtering, three warmups and 20 measured runs.
Payload: 465,281 UTF-8 bytes. p95=8.017ms against the 100ms budget. The 20 measured
runs in milliseconds (sorted, recorded to three decimals):
`7.675, 7.711, 7.735, 7.747, 7.761, 7.780, 7.803, 7.826, 7.854, 7.875,
7.878, 7.897, 7.906, 7.906, 7.908, 7.938, 7.955, 7.986, 8.017, 11.522`.
Each chart renders at most 1,000 numeric points; dense state timelines use a labelled
64-bucket duration view and recorded samples paginate at 100 rows.

Lifecycle browser tests count app-owned modal key handlers, pending history reads,
body scroll restoration and retained canvas DOM after 50 completed charts.
An initial Apex registry-only check missed real mobile retention. Forced-GC DOM
and listener counts caught it in both dev and compiled production. A chart-free
control stayed stable; changing resize/animation teardown and avoiding the Vue
wrapper did not fix it. Those attempted workarounds are not retained. The final
Chart.js renderer passes the stronger regression:

| Project | Cycle 25 heap / nodes / listeners | Cycle 50 heap / nodes / listeners |
| --- | --- | --- |
| Desktop | 15,112,152 bytes / 1,748 / 98 | 15,188,040 bytes / 1,748 / 98 |
| Mobile | 15,193,116 bytes / 1,748 / 96 | 15,271,844 bytes / 1,748 / 96 |

Owned modal handlers, pending history reads and retained canvas elements are zero;
body scroll is restored. The regression requires no more than 100 extra nodes,
five listeners or 2MiB heap growth between cycles 25 and 50. No such native-resource
growth occurred. This is evidence for the tested flow, not universal leak freedom.

## Rule review and limits

NH-01 PASS: amended spec/ADR/plan/tasks before production edits. NH-02 PASS: command
preview/confirmation stays separate, no ingestion/automation timing changes. NH-03
PASS scoped review: validated history ingress, same route access, no new secrets or
history persistence. NH-04 PASS: established service/repository/Vuex boundaries and
documented additive contract. NH-05 PASS: source-owned mock/temp-store regressions
and required checks. NH-06 PASS scoped review: units, bounds, partial/unknown data,
resource caps and code-only recovery; lifecycle evidence accompanies this review.

FE-01/02/03/04/05 PASS scoped review: shared ownership, matching runtime contract,
explicit network/command outcomes, keyboard/touch/focus/viewports/theme tokens,
existing API helper and no new sensitive persistence. BE-01/03/04/05 PASS scoped
review: layer boundaries, synchronous request-owned budgets, race tests, inclusive
ordering/empty/error contract and unchanged access. BE-02 N/A: lifetime ingestion,
automation events and sampling were not changed; full suites verify compatibility.

This completes the scoped entity/history amendment, not unrelated older 007 tasks.
Tests cover Chromium desktop/mobile emulation; physical iOS/Android, assistive-tech
sessions and real-household checks were not run. History is a bounded snapshot with
explicit refresh. Chart rendering remains a lazy optional download. Deploy backend
and frontend together; rolling back code needs no data migration.
