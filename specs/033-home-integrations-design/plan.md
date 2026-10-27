# Plan: Individual Home cards and shared panel layout

Spec: [spec.md](spec.md) | Status: Implemented and verified | Updated: 2026-10-27
Constitutions: proposed v1.0.0; no ratification or exception approval claimed.

## Scope and decision

The user explicitly selected **Implement in NodeHerder now**, superseding preview-only scope.
Use this branch's existing spec/plan/tasks set and the repository `sdd-plan` skill. Preserve
unrelated AGENTS/SDD/Claude workspace changes. No sub-agents, dependencies or package-lock changes.

Integration outputs are individual cards, inside any named area or directly on Home. There is
no mandatory Household group. Existing inline area handles/arrows/rename/delete, dashed outlines,
tile removal, Add tiles and New area/Export/Import/Done remain. Heading counts are removed.
Panel uses the same accepted layout and independent per-card visibility.

[ADR-001](adr-001-home-layout.md) selects a private optional Home aggregate in the existing settings
Bolt record. It is authoritative after the first accepted Home save. The same transaction stores
a derived device-only `hub.dashboardGroups` compatibility projection. No integration-key dependency
for device Home. Public settings snapshots strip the private field, including settings-save replies.
The [wire/migration contract](contracts/home-layout.md) specifies bounds, auth, errors and recovery.

## Ownership and implementation sequence

1. **Domain/cache:** `backend/models/settings/home.go` owns typed references, validation,
   deterministic read-only legacy migration, owned copies and device projection. `home_cache.go`
   owns accepted revision/ETag transactions and legacy adapters under the existing cache lock.
   Candidate state is persisted before publication; errors/conflicts retain accepted state.
2. **Ingress:** `backend/internal/api/home.go` has small consumer interfaces and protected local
   GET/PUT ingress. Validate new source identities against accepted local devices/resources;
   preserve already-accepted missing references as removable placeholders. Strict bounded JSON,
   cookie Origin/request marker and If-Match are required. No provider calls or new goroutines.
   Existing router/middleware wiring registers auth and CORS headers; no public access expansion.
3. **Frontend state:** `types/home.type.ts`, `contracts/home.ts`, `services/home.service.ts` and
   `store/modules/home/index.ts` own typed validated transport and accepted state. Session/read
   generation fences discard late replies; logout clears private state. Proposals are owned copies.
   Accepted replies also refresh the derived device-only hub view for Overview counts.
4. **Shared projections:** `domain/home-layout.ts` owns immutable placement commands and bounds.
   `home-compatibility.ts` owns legacy conversion/device projection; `home-panel.ts` owns panel
   filtering/control adapters; `home-picker.ts` owns selectable source choices. `domain/home-cards.ts`
   indexes accepted results once per resource. Existing waste
   date/status rules are shared with the Integrations board; no copied values in layout. Missing,
   empty, disabled, stale and failed results stay distinct. No invented dates or index identity.
5. **Page lifetime:** `useHomeLayout.ts` owns one AbortController and bounded local projection reads
   per mounted Home/panel page. Focus reloads are suppressed during edits/saves; hidden pages pause
   projection reads. Abort/unwatch/listener/timer cleanup occurs on logout/unmount. `useHomeGrid.ts`
   owns one ResizeObserver, batched measurements and coalesced frames, with replacement/disposal
   cleanup and generation guards. No per-card transport, polling, timers or observers.
6. **UI:** GroupDashboard composes the page and owns only rendering/drag/file-input adapters.
   `useHomeEditor.ts` owns dialogs and named commands; `useHomeLayoutFiles.ts` owns file confirmation
   and import lifetime. `services/home-export.service.ts` owns browser file IO and URL cleanup;
   `contracts/home-export.ts` owns format validation. AreaCard, HomeStandaloneCard and the focused
   picker/card/move components preserve established controls. Save-as-you-go publishes only accepted replies and restores failed drag copies.
   Cards move within/between groups or directly to Home; keyboard/tap Move supports panel visibility.
   Reference-only export and validated/confirmed Home/legacy import retain source ownership.
   PanelMode retains device climate/button semantics, clock/status/dock/keyboard navigation;
   mixed cards retain placement order and summaries span available columns. Detail links preserve
   allowlisted Home/panel return context through Results/Setup/operations/Help.

## Alternatives, compatibility and recovery

See the ADR for rejected competing device overlays, broadening legacy WS payloads, optional
integration-DB ownership and a new database. Before first Home save, legacy writes keep legacy
representation; deterministic content ETags detect intervening edits. Afterwards save/rename/
delete/import adapt to canonical layout atomically. Legacy imports retain integration placements
in matching areas or move displaced integrations directly to Home. Placement removal never deletes
sources. Existing integration-board preferences do not silently pin/hide Home cards.

Back up `settings.db` before upgrading. An older binary can use the derived device groups, but
subsequent old-binary writes may discard the unknown Home field. Restore the matching settings
backup before re-upgrading to recover those private placements. Integration encryption/AAD/schema
and provider execution are unchanged. Unknown/corrupt Home schemas fail visibly without defaults
silently overwriting them. No deployment or actual household migration was performed.

## Acceptance and rule review

| Rules | Design review before implementation | Final implementation evidence |
| --- | --- | --- |
| NH-01 | Updated spec, ADR, contract, plan/tasks before production code | Same branch scope; AC-18–23 reopened for actual implementation |
| NH-02 / BE-02 | Preserve device services, acknowledgements and automation semantics | EntityTile/PanelButton/PanelDial pipelines retained; full Go suite and browser device command regression |
| NH-03 / FE-05 / BE-05 | Protected bounded private API; no source snapshots in layout/public config | Auth/CSRF/CORS/admission/privacy tests; allowlisted return navigation; no household endpoints |
| NH-04 / FE-01/02 / BE-01 | One canonical cache owner, explicit wire and projections | Typed ingress, Vuex/services/domain owners and consumer interfaces; mocks updated |
| NH-05 | Source-owned unit/transport/cache tests and isolated UI planned | Actual checks below; historical preview evidence kept separate |
| NH-06 / BE-04 | Atomic candidate-before-publish, deterministic migration, compatibility and backup | Write-failure/concurrency/restart/legacy-adapter fixtures; confirmed import/group deletion |
| FE-03 | Loading/offline/conflict/rollback and independent source failure | Store/composable tests and browser conflict recovery; no automatic write retries |
| FE-04 | Shared named controls; keyboard alternatives; required responsive widths/themes | Keyboard Move, touch/pointer area drag, inline controls and Home/panel screenshots/geometry |
| BE-03 | No backend workers; existing cache mutex; owned snapshots | Concurrent save test, affected race detector; frontend cancellation/listener/observer/frame tests |

No approved exceptions. Before finishing, review changed public APIs for useful GoDoc/JSDoc,
complex copying/fencing/transaction comments, readable control flow, source-owned tests and links.
Constitution status remains proposed; this work does not claim governance approval.

## Actual checks

Commands use installed toolchains and isolated temporary storage/MQTT/provider fixtures. Go uses
`GOCACHE=/private/tmp/nodeherder-home-go-cache` because the default cache is sandbox-inaccessible.
HTTP/browser fixtures need permitted loopback execution outside the sandbox.

- `backend: go test ./...` — passed. Initial sandbox-only run failed binding loopback; permitted
  isolated run passed. `go build ./...` and `go vet ./...` — passed.
- `backend: go test -race ./models/settings ./internal/api ./internal` — passed after final cache
  changes. Covers migration/order, owned snapshots, privacy, legacy adapters, failure atomicity,
  concurrent save fencing, restart persistence, auth/CSRF/CORS, malformed and unknown source refs.
- `frontend: npm test -- --runInBand` — 77 suites / **458 tests passed**. An earlier concurrent run
  exceeded an existing unrelated history benchmark's 100ms p95 budget (118.65ms); clean reruns
  passed without changing that benchmark. New tests cover wire bounds/Unicode, service errors/
  cancellation, stable source mapping, immutable movement, mixed panel ordering, session/read fences,
  failed saves, page request counts and timer/listener/observer/frame cleanup.
- `frontend: npm run lint` — passed with **29 existing warnings**, no errors. `npm run build`
  (Vue/TypeScript plus Vite) — passed. Existing duplicate DeviceCard auto-import warning remains.
- `frontend: npm run type-check:e2e` — passed.
- `frontend: npm run test:e2e -- home.spec.ts panel.spec.ts ConnectionsPage.spec.ts --reporter=list`
  — **78 passed** across desktop/mobile. After the final legacy-view projection, Home/panel/Overview
  confirmation — **36 passed**. Expanded matrix/touch area-drag confirmation — **4 passed**.
  Home normal/edit and panel fit 390/1024/1280/1440px in both themes. These runs exercise individual
  service/summary selection, group/direct movement, per-card panel visibility, accepted reload,
  export/import roundtrip, keyboard Move, detail/Setup/back, rename/delete, conflict rollback,
  normal device commands and existing integration lifecycle/editor/navigation regressions.
  Atomic cross-group card-drag confirmation — **2 passed**, mouse and emulated touch; each
  saves exactly one transaction and survives reload. Native drag admission is awaited before
  releasing the gesture; the first fast-mouse harness attempt ended before asynchronous admission.
  Final settled desktop/mobile screenshot/matrix pass — **2 passed**; inspected Home normal/edit
  and panel summaries. Sidebar/theme transitions are awaited before geometry/screenshots.
  Final scrolling-container/card-bound matrix — **2 passed**, both input projects, four widths
  and both themes in Home normal/edit and panel. An intermediate assertion incorrectly treated
  the remove badges extending into page padding as clipping; actual app scroll areas and card
  bounds fit. The regression checks those boundaries while retaining the existing badges.
  Initial selector/restoration assertions were corrected; an intervening dev-server HMR reload
  invalidated one earlier run. Completed runs above passed with source stable.
- `git diff --check` — passed. No real household devices, providers or services tested.

## Limits and future work

Supported new integration cards are individual waste services and an optional schedule summary;
weather/energy need real contracts/renderers. Lifecycle tests are deterministic resource cleanup
checks, not a production heap or long-duration wall-panel soak. No real touch hardware, assistive
technology or production performance guarantee is inferred. Browser privacy includes references
only; no new layout/source state is persisted to localStorage/sessionStorage. Existing numeric
panel-page/browser-theme preferences retain their existing ownership.

## Research and historical artifacts

[Home Assistant investigation](designs/home-assistant-investigation.md) distinguishes live demo
observations, official documentation and NodeHerder proposals. [Source investigation](designs/investigation.md)
records the original owners. [Card-first preview](designs/home-cards.html) remains an offline sample;
[three-design comparison](designs/home-integrations.html) contains the superseded Household option.
[Historical preview verification](preview-verification.md) preserves the previous 274/258/273/164
check results; none are counted as production acceptance.

## Panel summary follow-up (AC-24)

The supplied panel screenshot exposes excessive horizontal stretching that overflow checks do
not detect. Before editing: cap only `.is-panel.is-summary` in IntegrationTile.vue at 36rem,
keep its existing full-row grid placement for safe narrow layouts, and increase panel collection
row typography/spacing. No layout/source/transport/storage changes. A bounded tile is preferred
to changing the shared panel grid, which could disturb existing device controls or create implicit
columns at narrow widths. Update the existing Home E2E flow to assert readable bounded width at
390/1024/1280/1440/1680px and inspect desktop/mobile screenshots. Reuse source-owned tests; no
CSS-mirroring unit test. Run frontend unit/lint/build, E2E type check and relevant isolated browser
checks. NH-01/04/05 and FE-04 apply directly; NH-02/03/06 and FE-01/02/03/05 boundaries are retained.
Backend rules are N/A for this CSS-only correction.

Actual checks: the pre-fix browser regression failed at the panel row text size (12.8px versus
16px required); supplied screenshot/source show the edge-to-edge expansion. After the correction,
all 77 frontend suites / 458 tests pass, lint passes with the same 29 existing warnings, production
build and E2E type check pass. The 1680px three-bin screenshot shows the 576px summary cap and
readable aligned rows. Full Home/panel browser run: 36 passed; desktop/mobile screenshots inspected.
The first browser launch could not bind the occupied development ports; the user closed their
running app and the isolated run proceeded. No household devices/providers were exercised.

## Frontend refactor audit and owners (AC-25)

The user correctly identified mixed responsibilities and dense code that passing tests did not
resolve. Reread AGENTS.md, frontend constitution, Vue rules and repository SDD skill. This is an
implementation refactor, not approval of an exception or a new feature/branch planning set.

Pre-refactor audit: GroupDashboard combined domain edits, dialogs, file parsing/validation, DOM downloads,
confirmation and view rendering; templates issue saves and create proposals. home-layout combined
placement commands, compatibility and panel projection. AddTilesDialog built domain choices
inline. Contract parsing used vague helpers and a large callback; page lifetime used boolean error
routing and imperative editing state. Public helpers lacked precise types/docs in several places.

Refactor before further acceptance: keep immutable commands in home-layout; move legacy conversion
into home-compatibility and panel projection into home-panel. Add a focused picker projection.
Use useHomeEditor for dialog/action orchestration and useHomeLayoutFiles plus a browser file
service/ingress parser for transfer. Extract standalone block presentation; GroupDashboard becomes
composition/view wiring with named handlers. Session loading/cancellation stays in useHomeLayout,
with edit-pausing supplied by its caller and readonly accepted feedback. Simplify ingress helpers
and source mapping, name intermediate concepts, use explicit branches/braces, and format Vue
markup for human reading. Do not replace Vuex/UI controls or introduce a framework/generic engine.

Existing/relocated companion tests follow production owners; add meaningful regressions for
orchestration, file cancellation/conflicts and picker filtering. Preserve private reference-only
export, legacy adapters, immutable accepted state, one page request scope/observer, device command
semantics and panel selection/keyboard controls. NH-01/04/05 and FE-01–05 apply; other existing
boundaries remain as previously reviewed. Run full frontend unit/lint/build/E2E type checks and
isolated Home/panel/integration browser regressions. No backend/device changes are planned.

### Refactor completion and review

Implemented the owners above and reviewed all touched frontend production modules, their DTOs,
Vue templates, mocks and source-owned tests. GroupDashboard's script now composes page concerns
and delegates named actions; no JSON parsing, DOM downloads, domain proposal construction or
inline save/mutation expressions remain in its template. Standalone rendering has a presentational
component. Picker/source projection, placement commands, compatibility and panel projection have
separate owners. HomeCardSource distinguishes source identity from a HomeCard placement, eliminating
invented IDs/visibility in the source catalog. Shared layout limits have one frontend owner.

Ingress uses named bounded readers and a typed membership context, not a large mapping callback
or unchecked response cast. HTTP transport, file ingress and browser IO are separate. File imports
capture the accepted revision, serialize overlapping reads, suppress abandoned callbacks and revoke
URLs in a finally block. Request-session publishing has one Vuex guard. Editor-owned refresh pausing
is explicit; page feedback refs are readonly. Observer replacement/disposal, cancellation and all
existing source/device behavior remain covered. Public APIs and non-obvious ownership/fencing paths
have useful JSDoc; Vue markup is formatted without compressed adjacent component/closing tags.

Actual final checks:
- `npm test -- --runInBand`: **84 suites / 478 tests passed**. New/relocated companions cover the
  new owners, immutable commands, picker exclusion/deduplication, source/placement distinction,
  import format/privacy, stale confirmation, cancellation, overlapping file reads and URL failure cleanup.
- `npm run lint`: passed, **0 errors / 28 warnings** outside the new Home modules. `npm run build`
  and `npm run type-check:e2e`: passed. No dependency/package-lock/framework change.
- `npm run test:e2e -- home.spec.ts panel.spec.ts ConnectionsPage.spec.ts overview.spec.ts --reporter=list`:
  **82 passed**, desktop and mobile, with the app source stable for the run. Includes the complete
  Home/panel editor/import/drag/conflict/navigation flow, three-bin compact summary geometry in both
  themes through 1680px, existing device commands and integration/Overview regressions.
- Final `git diff --check`, local links and screenshot review are recorded at handoff. Go suites are
  not repeated for this frontend-only refactor; their previous production evidence remains above.
  No household device/provider tests. Heap/long-duration hardware/assistive-technology limits remain
  as previously recorded; passing checks are not a claim of zero undiscovered defects.

No exceptions were requested or invented. Preserve unrelated AGENTS/SDD/Claude workspace changes.

## Narrower desktop summary (AC-24 follow-up)

The user reviewed the 36rem panel summary and requested the agreed narrower 360–400px desktop
width. Set the panel summary cap to 24rem (384px at a 16px root size), preserving the existing
mobile fit, content and full-row grid placement. This is scoped to the existing presentation owner;
no source, persistence or panel navigation behavior changes. Extend the existing browser geometry
check to assert the new cap and mobile fit across the agreed viewport/theme matrix. Run frontend
unit/lint/build/type checks and isolated browser validation; inspect desktop/mobile screenshots.
NH-01/04/05 and FE-04 apply; unchanged data/command/trust boundaries retain prior evidence.
Actual checks: the pre-change desktop regression measured 576px against the new 384px maximum
and failed as expected. After the change, the isolated Home/panel browser flow passed on desktop
and mobile (2 projects), checking three collection rows at 390/1024/1280/1440/1680px in both
themes; 1680px desktop and mobile screenshots were inspected. Frontend unit tests passed
(84 suites / 478 tests); lint passed with 28 existing warnings; build and E2E type check passed.
No real household devices or providers were used. Backend checks were not repeated for this
frontend-only presentation change.

## Home card spacing (AC-26)

The reported screenshot shows the standalone collection summary displaced from adjacent device
cards. The outer grid uses a larger column gap than the shared tile gap and rounds block heights to
24px steps. Before editing, reproduce with unequal-height areas and a standalone summary. Set
column spacing to the shared tile gap and use fine-grained grid rows; include one tile gap in each
block's measured border-box height. The existing page-owned ResizeObserver continues to batch
measurements and release targets/frames on replacement or unmount. No layout membership, API,
source or panel behavior changes. Verify actual card bounding boxes, no overlaps, drag/edit and
mobile/desktop sizes; update source-owned observer tests and isolated browser evidence. NH-01/04/05,
FE-01/04 and BE-03 lifecycle reasoning apply; other boundaries are unchanged. The border-box
observer also catches density changes, and the edit outline offset fits compact spacing.

Before the fix, the isolated browser regression measured a 10.39px difference between the outer
and inner horizontal gaps; the source-owned observer regression also failed when padding was
excluded. After the fix, the browser checks passed at 1280px and 1440px for comfortable, compact
and spacious density, including aligned rows and columns. Normal and compact edit-mode screenshots
were inspected. The Home and panel browser suite passed 38 tests across desktop and mobile projects,
including responsive geometry and drag/reorder checks. The frontend unit suite passed 478 tests;
lint passed with 28 existing warnings, build and E2E type-check passed. The browser fixtures used
synthetic devices and providers. No backend changes or household devices were involved.


## Standalone drag control (AC-27)

Place the existing standalone block drag handle and reorder arrows in a visible toolbar above
the card. Retain the separate card-placement controls for moving between areas and their existing
behavior. Exercise drag persistence and keyboard alternatives using isolated browser fixtures;
check normal/edit layouts and run the required frontend checks. NH-01/04/05 and FE-01/04 apply.
An isolated desktop browser test verified the visible handle, drag to a new block position, saved
order after reload, and the earlier arrow. The card-placement control remained present. The
frontend suite passed 478 unit tests; build, E2E type-check and lint passed (28 existing lint
warnings). The default fixture port was occupied, so the browser check used a separately started
synthetic Go fixture and Vite server, with identity verification and routed requests. No household
devices or provider services were used.


## Consistent edit controls (AC-28)

Restore compact tile spacing by overlaying card-placement controls on hover or keyboard focus.
Render standalone block controls with the existing area header visual treatment, and preserve
the area drag handle and named keyboard controls. Verify area and standalone dragging with mixed
blocks, card transfer, persistence, and edit-mode geometry in isolated fixtures. NH-01/04/05 and
FE-01/04 apply. Card drag uses Sortable's fallback mode to keep nested-grid transfers working with
positioned controls. The isolated Home suite passed 18 desktop/mobile tests, including mixed area
and standalone dragging with persisted order, card transfer, responsive geometry and failed-save
recovery. The compact edit screenshot was inspected. Frontend unit tests passed 478 tests; build,
lint and E2E type-check passed (lint reports 28 existing warnings). The alternate-port fixture
sign-in failed before testing, so the final browser run used the standard isolated fixtures on free
ports. No household devices or providers were used.


## Consistent standalone headers (AC-29)

Share the area editing surface and header text color with standalone blocks. Add a header delete
action using the existing confirmation and placement-removal paths, retaining sources. Verify both
themes, cancel/confirm behavior, and existing drag/layout interactions in isolated fixtures.
NH-01/04/05 and FE-01/04 apply. Isolated browser checks confirmed matching text, background and
outline colors in both themes, cancellation and confirmed deletion, source preservation, and
persisted reorder. The full Home browser suite passed 18 desktop/mobile tests; the edit screenshot
was inspected. Frontend unit tests passed 479 tests; build, lint and E2E type-check passed (28
existing lint warnings). The alternate-port fixture sign-in did not complete; final browser checks
used standard isolated fixtures once their ports were free. No household devices were used.


## First-row card alignment (AC-30)

The first grid row includes area headings while standalone cards have no heading in normal mode.
Use the page-owned grid observer to mark standalone blocks occupying that first row and reserve
the same heading space; do not add that space to cards in later rows. Reassess placement after
layout or grid-width changes, batch reads before writes, and preserve observer/frame cleanup.
Verify top and lower rows across reorder, viewport widths and densities, alongside edit mode and
frontend checks. NH-01/04/05 and FE-01/04 apply. Isolated browser checks passed for the card
moving between top and lower rows as the viewport changed, reordering it to the first position,
three densities and widths from 1280 to 1680px. The first-row screenshot was inspected. A separate
isolated browser check passed edit-mode area/card dragging, saved ordering, and deletion. Frontend
unit tests passed 479 tests; build, lint and E2E type-check passed (28 existing lint warnings).
Browser fixtures used alternate local ports while the user's app was running; no household devices
or providers were used.
