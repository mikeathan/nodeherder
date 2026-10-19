# Plan: Frontend redesign and customisable design system

## Mobile entity repair — 2026-10-18

Apply AC-51–54 in shared UiDialog/EntityViewDialog (stable shell), UiRange and pure
control geometry (continuous pointer presentation with native step commands), and a
small SVG EntityMeter (reading visual; no new chart dependency or invented scale).
EntityReadout owns the state/alert badge and environmental reading presentation;
mobile writable views retain controls plus compact history.
Keep Vuex confirmed state, command confirmation, raw history semantics and bounded
query ownership. Existing uncommitted frontend repairs are retained and reviewed.

Verify first-frame and settled geometry with delayed/empty/error history for each
family on mobile, touch/mouse/keyboard/cancel command semantics, meter unknown/zero,
small/landscape/desktop layouts, screenshots and browser errors. Run relevant Jest,
Playwright, lint/build and changed-function quality checks. Prior passing evidence
is historical and does not prove this follow-up visually matches the approved design.
NH-01/04: scoped acceptance and existing boundaries; NH-02/06, FE-01/03: native
steps, no fabricated readings, unchanged transport/persistence; NH-03/FE-02/05:
no new ingress/secrets; NH-05/FE-04: isolated interaction/viewport verification.
No exceptions or adoption are claimed. Rollback is code-only.

Visual completeness (AC-55): keep the shared expose registry as the owner of state
palettes, use theme tokens for SVG meter stops and chart styling, retain chart gaps,
and add battery presentation to EntityReadout. ExposeControl gets an explicit
expanded presentation for short choices in entity views; device rows retain compact
controls. Source-owned domain/options tests plus mock browser checks cover labels,
colours, battery missing/zero and confirmed-setting semantics. No new dependency,
transport, timing or persistence change; NH-01–06 / FE-01–05 checks remain applicable.

AC-56 repair: remove the added pointer-position engine from UiRange, restoring one
native range/local-value owner and change-only commits. Keep shared modern styling,
keyboard focus, cancellation and native bounds. useExposeEdit retains its draft while
pending even when an unrelated/unconfirmed device report arrives. Add a source-owned
regression before fixing the draft reset, then verify both native drag directions,
coarse/decimal bounds and native command units against isolated mocks. No backend,
dependency or command transport change; NH-02/05/06 and FE-01/03/04 remain gates.

Final implementation review: [2026-10-18 evidence](implementation-evidence-2026-10-18.md).

## Implementation authorised — 2026-10-17

Implement FR-20–24/AC-44–50 on `feat/entity-history-redesign`; the final implementation
evidence above records the completed change map. Use Chart.js 4.5.1 with a shared theme-aware adapter after the Apex mobile teardown
regression (ADR-004), plus native accessible
range inputs and proportional HTML state timelines. Pure functions own history and
control conversions; Vuex owns query results and hub confirmed state. Browser handles
live outside serialisable store state and are disposed by explicit subscriptions.
HTTP history is additive, bounded and cancellable at collector callbacks; no migration.
Energy bucket semantics and raw enums follow the amendment in spec.md. Backend changes
ship with frontend and old HTTP/WS consumers remain compatible. Rollback is code-only.

Before implementation rule check: NH-01 PASS (amendment/tasks); NH-02 PASS design
(native commands, confirmation, no ingestion changes); NH-03 PASS design (same route,
runtime decode, no logs/persistence); NH-04 PASS design (existing boundaries, ADR-004
amended); NH-05 PASS design (source-owned tests and planned commands); NH-06 PASS design
(bounded resources, units/partial/gaps). FE-01–05 and BE-01–05 PASS design against the
planned ownership/verification boundaries. This records the pre-implementation
checkpoint; final execution evidence is linked above. No
ratification or exception approval is claimed. Review these gates against final diff.

Verification: Jest coverage, frontend lint/build, mock Playwright desktop/mobile plus
viewport/leak checks; Go tests/vet/build and affected metrics race tests. Quality runner
uses TypeScript AST for TS/JS/Vue script complexity, Istanbul statement coverage where
available, zero-coverage upper bounds otherwise; Go scorer pinned/recorded in evidence.
Baseline is branch parent recorded by git, not an unrelated repository. Vue behavior
is exercised through Playwright; adding a second unit framework is unnecessary.

## Entity view corrective pass — 2026-10-18

Owner review of `7ccdcfcb` found that the implemented entity views do not match the
accepted v3/v4/v5 direction and that some states are misleading. Observed against the
isolated mock at 1120×850 and 390×844:

- An offline or missing adjustable reading renders a large `—` where the value belongs.
  At the same time the native range falls back to `min`, so the whole vertical rail
  uses `--nh-range-empty` and the grip sits at the bottom. The dash reads as a floating
  handle above the slider (the owner's "handle above the slider").
- The vertical grip is offset by three percentage points below the fill boundary and is
  not centred on it; the horizontal grip has the same offset.
- `UiDialog` focuses the first body input on open, so the hidden native range draws a
  focus outline around the whole slider before any keyboard interaction.
- The power/state control has no visible On/Off/Unknown text, and a read-only entity
  whose primary control is writable (alarm) renders an unlabelled switch.
- Colour temperature omits the accepted warm/cool endpoint labels and the warmth
  descriptor below the value.
- Adjustable desktop views keep the compact history variant, leaving most of the
  second column empty instead of using a larger chart and summary.
- A counter series with no complete interval shows empty bar axes, `—` consumption and
  no explanation.
- The illuminance/energy screenshots used a stale local mock process. Current mock code
  returns in-range samples; verification must start from a fresh isolated mock.

Corrections, all implementation-level against existing FR-20/21/22/24 and AC-44/45/46/48;
no new HTTP/WS contract, ingestion, automation, storage or dependency change:

1. `UiRange` gains an explicit unknown/disabled presentation: the rail is muted, the
   fill and grip are hidden, and a preview value reveals the real fill. The grip is
   centred on the boundary with `calc()` in both orientations. Percent conversion moves
   to a tested `domain/controls` helper.
2. `EntityAdjustment` shows a muted `Unknown` readout (never a fake minimum), a warmth
   descriptor for colour temperature, and `Warm`/`Cool` endpoint labels. The status line
   is reserved for sending/release feedback. When controls are unavailable it collapses
   the full-height rail to a compact range line instead of a dead slider.
3. `EntityViewDialog` uses circular labelled mode controls, a labelled power control
   with On/Off/Unknown state, labelled writable controls for read-only entities, and a
   compact "Other readings" summary where the device has other measurements. Desktop
   uses the full history variant; a small `useMediaQuery` composable selects it.
   Read-only views without samples use the ordinary dialog width and an explicit empty
   state instead of an oversized blank history column. Unavailable/empty mobile views
   use a content-height sheet; full-screen is reserved for a usable control or history.
4. `UiDialog` moves initial focus to the dialog surface, not the first body input, while
   keeping focus trapped and restored; an explicit `[autofocus]` still wins.
5. `HistoryView`/`NumericHistory`/`options` explain counter series with no complete
   interval instead of drawing empty axes, distinguish unavailable consumption from
   measured zero, and read tooltip colours from theme tokens.
6. Raw HTTP binary history is typed as boolean while live z2m values often use
   `ON`/`OFF`; `binaryStateOf`/formatting now accept both so state timelines and recent
   reports no longer show `Unknown` beside a confirmed `On`. The missing
   `--nh-kind-switch` token is defined so On/Off timelines and legend dots have a colour.
7. Device-row sliders keep the compact horizontal treatment: 1.75rem visual rail with
   a 44px touch target, kind colour, and mired/Kelvin direction inverted consistently
   with the entity view. The device card no longer expands the range to the row width.
8. Alert-family views expose their writable silence/test action as a labelled row.
   A missing current reading shows the latest recorded sample with time, other readings
   list only known values, and very busy timelines (>96 segments) use the labelled
   64-bucket activity density instead of hundreds of thin buttons.
9. Energy history is compact at narrow widths and shows a notice instead of empty axes
   when no interval has measurable consumption, while the meter total, controls and
   consumption summary stay primary.
10. Regression coverage extends `domain/controls`, `domain/exposes`, history state,
   chart options and `EntityViewDialog.spec.ts` (unknown/offline, grip geometry, focus,
   labels, empty-history width, latest-recorded fallback, alert action, desktop summary,
   320/390 widths) plus the existing device-row `ExposeControl` browser flows.
   Verification starts fresh mock/frontend processes.

### Rule check

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | This section, spec corrective note and T307–T317 reproduce the defects and expected behavior | PASS | PASS |
| NH-02 | Commands still go through `useExposeEdit`/`useDeviceCommand`; unknown is never shown as confirmed; labels only | PASS | PASS |
| NH-03 | No new entry points, payload fields, logs or persistence; mock-only verification | PASS | PASS |
| NH-04 | Existing component/domain ownership; no new dependency; shared chart options retained | PASS | PASS |
| NH-05 | Source-owned unit tests plus Playwright desktop/mobile; exact commands recorded | PASS | PASS |
| NH-06 | Unknown vs zero, no fake fill, explicit no-complete-interval text; bounded rendering unchanged | PASS | PASS |
| FE-01/02/03 | Presentation only; existing contracts; requested/confirmed separation documented | PASS | PASS |
| FE-04 | Accessible names, keyboard focus, 44px targets, 320/390/768/1280/1920 checks | PASS | PASS |
| FE-05 | No secrets, bundle or persistence change | PASS | PASS |
| BE-01–05 | N/A: no backend change in this correction | N/A | N/A |

Verification (2026-10-18, final code): Jest 42 suites / 278 tests, frontend lint 0 errors
(7 pre-existing warnings), build/type-check pass, mock history contract pass, full
Playwright suite 118 passed on fresh mock/frontend processes, and the quality runner
reports 312 touched functions, max C=8 / max CRAP=30, zero failures. Desktop/mobile
screenshots confirm unknown/known brightness, colour temperature, energy, alarm, state
history and empty-history read-only states at 320/390/768/1280/1920 widths. No household
device was contacted; only the isolated mock ran for browser evidence.

## Entity/chart redesign investigation — 2026-10-17

The initial investigation established the accepted control-first mobile direction
and shared history architecture. Implementation was subsequently authorised and
completed; see the final evidence above. Initial concept images and superseded
investigation/readiness reports were removed at the owner’s request. Git history
preserves the original preparation material.

Spec: [spec.md](spec.md) | Tasks: [tasks.md](tasks.md) | Status: Main redesign merged; scoped backend repair implemented; wider follow-up deferred
Author: Claude Code / Codex | Updated: 2026-10-12 | Follow-up baseline: `fcdebca2` (#54)
Constitutions: parent/frontend/backend 1.0.0, proposed; none claimed ratified.

Active repair: delayed illuminance, brightness confirmations and routine log noise.
The owner requested implementation after the 2026-10-11 investigation; the backend
portion is implemented on `investigate/ui-device-updates`. Findings and verification:
[device-update repair](#device-update-repair-2026-10-12). Progress and remaining work:
[tasks](tasks.md#current-status-2026-10-12).

The main Hearth + Panel redesign merged as #54 (`fcdebca2`). Original design/execution
sections below are historical. The broader [follow-up plan](#follow-up-plan-2026-10-11),
[entity/history design](entities-and-history.md) and [ADR-004](adr-004-history-and-quality.md)
remain deferred; they are not completed by this repair.

Supporting documents: [current state](current-state.md) ·
[design system](design-system.md) · [device & protocol support](device-support.md) ·
[automation editor](automation-editor.md) · [ADR-001 theming](adr-001-theming-tokens.md) ·
historical prototype review (samples removed after the main redesign).

Progress: Phase 0 (documentation + samples) complete. Q-01 resolved 2026-10-10 (Hearth + Panel);
the owner authorised implementation. See [Implementation design](#implementation-design-2026-10-10).

## Context and design

### Observed facts before the original redesign (historical)

See [current-state.md](current-state.md). In short: Vue 3 + Vuex + PrimeVue 4 with a
660-line Material preset, hard-coded component colours, dark toggle only; nav drawer +
menubar shell; group dashboard of compact `EntityCard`s; automation editor as a panel
stack; `connection_type` is `mqtt` (Zigbee2MQTT) or `http` (`/api/collect`, used by Wi-Fi
devices).

### Proposed architecture (smallest design meeting FR-01…FR-08)

```text
             ┌──────────── theme layer (ADR-001) ────────────┐
 nodeherder_theme (localStorage) → useThemeSettings() → <html data-design data-mode
             data-density style="--nh-accent…"> → CSS tokens (--nh-*) → PrimeVue preset
             (semantic tokens mapped to var(--nh-*)) + nh-ui primitives
             └───────────────────────────────────────────────┘
 Vuex hub/ws/automations (unchanged truth) ─► pure selectors (src/selectors/*)
        │                                     ├ networkSummary(devices)
        │                                     ├ describeTrigger(trigger, devices)
        │                                     └ validateAutomation(automation, devices)
        ├► registries (src/registry/): protocols, capabilities (expose → icon/format/control)
        └► composables: useDeviceCommand (pending/confirmed), useAutomationDraft (dirty,
           validation, leave guard), useActivityFeed (bounded in-memory ring buffer)
 Screens (src/components/…) consume primitives + selectors; no screen talks to WS directly.
```

1. **Theme & tokens** — CSS custom properties in three tiers (primitive palette →
   semantic roles → component tokens) in `src/assets/styles/tokens/`. Each design
   ("skin") is a token file + a shell layout variant. PrimeVue keeps doing behaviour and
   accessibility; its preset is replaced by a small `definePreset(Aura, …)` whose
   semantic colours point at `var(--nh-*)`, so runtime accent changes need no rebuild.
   `services/theme.service.ts` becomes `composables/useThemeSettings.ts` (migrates
   `localStorage.theme`). Details: [design-system.md](design-system.md), ADR-001.
2. **Shell** — `components/layout/AppShell.vue` with variants `sidebar` (HA/Z2M style),
   `rail`, `topbar`, `dock` chosen by design; shared `useNavigationItems` stays the
   single nav source. Header shows connection state text+icon (AC-17), permit-join
   countdown chip, theme quick toggle, user menu.
3. **UI primitives** — `components/ui/`: `NhTile`, `NhCard`, `NhStat`, `NhBadge`,
   `NhChip`, `NhToggle`, `NhSegmented`, `NhSlider`, `NhEmpty`, `NhStatusDot`,
   `NhSection`. Thin wrappers over PrimeVue where one exists (FE-04 "use shared
   controls"); they only consume tokens.
4. **Registries** — `registry/protocols.ts` (label, icon, diagnostics, fallback) and
   `registry/capabilities.ts` (expose → presentation/control), replacing scattered
   logic in `sensor-formatter.ts`, `ConnectionType.vue`, `device.config.ts`.
   [device-support.md](device-support.md).
5. **Commands** — `useDeviceCommand(deviceId, expose)` returns
   `{request(value), pending, error}`; it dispatches existing `hub/setDeviceValue`,
   marks pending, resolves on matching `deviceUpdated` value, times out (Q-05, 5 s) to
   an error toast + revert. No retries (FE-03 "retries must not duplicate commands").
6. **Overview & activity** — `selectors/network.ts` (counts, low LQI/battery);
   `useActivityFeed` subscribes to `hub/updateDevice` via `store.subscribe`, keeps the
   last 200 deltas in memory only (FE-05), never throttles the store (NH-02/NFR-04).
7. **Automation editor** — single-page "When / If / Then" flow replacing the panel stack;
   draft model, sentence summary, inline validation, multi-action ordering (Q-02),
   schedule timeline, raw JSON view (read-only), Run with in-flight lock.
   [automation-editor.md](automation-editor.md).
8. **Screens** — rebuilt on the above: Login, Overview (new), Home/Areas, Devices,
   Device list (search, filters, sortable columns, protocol/power/LQI chips), Device
   page (header with status chips + tabs), Automations list (status, trigger count,
   last edited), Editor, Assistant, Console (virtualised list, level chips, pause),
   Settings (+ new Appearance tab), dialogs, toasts, permit join.

### Alternatives considered

| Option | Why not chosen |
| --- | --- |
| Switch to Vuetify/Quasar/Tailwind UI kit | Replaces component system → FE-01 ADR, large migration, no specified benefit over PrimeVue |
| Keep MaterialBlue preset, restyle per component | Cannot support runtime accent/density/design switching; perpetuates hard-coded colours |
| PrimeVue unstyled mode + own CSS for all components | Max control but rebuilds styling of ~25 PrimeVue widgets; reconsider only if a sample's look cannot be reached via tokens (recorded in ADR-001) |
| Free-form Lovelace grid (drag/resize anywhere) | Needs new persisted layout contract (Q-04/Q-06); areas + ordered tiles meet AC-03 |
| Visual node-graph automation editor | High cost, poor on mobile; block flow covers the linear trigger→conditions→actions model exactly |

### Design samples (Phase 0 deliverable)

Phase 0 provided five static prototypes sharing mock data and a renderer, each with
its own skin and layout. The owner requested removal of the prototype pages, assets
and generator scripts on 2026-10-12 after completion of the main redesign. The
following records the historical selection and review evidence.

Round 3 (2026-10-10, current recommendation): **Hearth + Panel**, see
[design-directions.md](design-directions.md#round-3-2026-10-10-hearth--panel).

Round 1 (2026-10-09) offered five skins on one layout. Owner feedback (2026-10-10): keep Hearth;
the rest must be structurally different. Round 2 is described in
[design-directions.md](design-directions.md):

| # | Name | Structure | Navigation | Automation editor |
| --- | --- | --- | --- | --- |
| 1 | Hearth (kept) | Areas of tiles | Sidebar + top bar | When / If / Then flow cards |
| 2 | Floorplan | Drawn floor plan + room inspector | Top tabs (bottom on phones) | 5-step wizard |
| 3 | Workbench | Explorer tree, open tabs, property grids, log panel | Activity bar + tree + status bar | Outline + inspector + live JSON |
| 4 | Brief | One column of sentences | Command palette (Ctrl K) + text menu | Fill-in-the-blanks sentence |
| 5 | Deck | Wall-tablet room pages, dials | Room pager + bottom dock | Node canvas + drawer |

Each prototype had a layout module over a shared renderer for state, contracts,
actions and editor field builders, so every editor emitted the same `saveAutomation` payload. In production the equivalent split is shared
composables/selectors plus per-layout route components (FE-01).
Every sample had a live customiser (mode, accent, density, radius, font scale, preset).
These prototypes served as review aids before implementation and have now been removed.

## Constitution Check

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | This spec + open questions; Phase 0 is documentation (lightweight path) | PASS for Phase 0; Phases 1+ BLOCKED on Q-01 | Pending |
| NH-02 | `useDeviceCommand` pending ≠ confirmed; activity feed reads after store commit; no UI throttling of store | PASS (design) | Pending AC-04/NFR-04 evidence |
| NH-03 | No new endpoints; no secrets; assistant text stays text | PASS (design) | Pending |
| NH-04 | Vuex/WS/HTTP boundaries kept; selectors are pure & unit-testable; ADR-001 for theming dependency/boundary | PASS (design); ADR-001 pending review | Pending |
| NH-05 | Traceability table below; Jest for selectors/registries/round-trip; manual procedures for UI | PASS (plan) | Pending |
| NH-06 | Destructive actions (delete group/automation/device, clear schedules) keep confirmation with scope text; editor leave guard; theme migration non-destructive | PASS (design) | Pending |
| FE-01 | State stays in Vuex; new composables/registries for shared behaviour; no Vuex/router/component-system replacement | PASS (design) | Pending |
| FE-02 | No contract change; AC-14 round-trip test; mocks in `tools/server` reused | PASS (design) | Pending |
| FE-03 | Loading/empty/error/offline/reconnect states specified per screen in [design-system §States](design-system.md#states); no command retries | PASS (design) | Pending |
| FE-04 | Tokens enforce AA (NFR-02); status = icon+text; keyboard + focus ring token; viewports 360/768/1280/1920 | PASS (design) | Pending manual checks |
| FE-05 | Theme storage holds presentation only; activity feed in memory; no `v-html` for assistant (existing `formatTriggerConditions` `v-html` replaced by text renderer) | PASS (design) | Pending |

## Contracts, data, rollout

- HTTP/WS/MQTT/MCP: unchanged. Editor must emit `saveAutomation` payloads identical
  in shape to today (AC-14). Field order/optional fields preserved (`delay`,
  `publishMode`).
- Browser storage: new `nodeherder_theme` (versioned `{v:1,…}`); one-time migration from
  `theme`; unknown/invalid values fall back to defaults. Rollback: old build ignores
  the new key and reads `theme`, which migration leaves in place.
- Rollout: behind a per-browser "new UI" switch during Phases 1–5
  (`nodeherder_theme.design = 'legacy'` keeps old preset), removed in Phase 6.
- Pre-existing defects touched by redesigned screens are fixed in the same phase with
  regression tests: `setDeviceDeConfigfaults` typo, missing `renameDashboardGroup`
  mutation, inverted editor button enablement, `alert()` usage, raw `fetch` in
  `automation-trigger.service.ts` (move to `post()` helper, FE-05).

## Traceability and verification

| Requirement/acceptance IDs | Paths | Test/manual procedure | Actual result |
| --- | --- | --- | --- |
| AC-01, AC-02 | Historical prototypes (removed after implementation) | Recorded gallery review in Chromium across all designs, screens and customiser controls | Round 1 pass 2026-10-09; round 2 pass 2026-10-10 (5 designs × 13 routes + All pages, standalone files opened alone, editor validation, Brief palette, Deck node drawer, 375 px no horizontal scroll). Playwright + Chromium 1194 via `file://`, fonts blocked (offline fallback). 13 routes × 5 designs at 1440×900 rendered with no page errors; scripted checks per design: tile toggle shows pending then confirms; editor shows inline error + disables Save after adding an empty action; leave guard blocks navigation and "Discard and leave" proceeds; mode switch applies; 375 px width shows no horizontal page scroll on Home/Editor/List. Not checked: Safari/Firefox, screen readers, contrast measurement (NFR-02 is a Phase 1 gate). |
| FR-02, AC-18, NFR-02 | `assets/styles/tokens/`, `composables/useThemeSettings.ts` | Jest: migration + validation; contrast script over preset tokens | Pending |
| AC-03, AC-05, AC-06, NFR-01 | `components/dashboards/*`, `components/ui/*` | Manual at 360/768/1280/1920 with `npm run test-server` mocks; keyboard pass | Pending |
| AC-04, FR-04 | `composables/useDeviceCommand.ts` | Jest with fake store: confirm, timeout, offline; manual with mock server | Pending |
| AC-07, AC-08, FR-07, FR-08 | `domain/devices.ts`, `domain/exposes.ts` | Jest: known/unknown protocols, capability resolution for all expose types | Pending |
| AC-09, FR-05, NFR-04 | `selectors/network.ts`, `composables/useActivityFeed.ts` | Jest counts/thresholds; perf script 100 tiles × 10 msg/s | Pending |
| AC-10…AC-15, FR-06 | `components/automations/editor/*`, `selectors/automation.ts` | Jest: describeTrigger, validateAutomation, round-trip equality over `docs/` + test fixtures; manual leave-guard and Run lock | Pending |
| AC-16, AC-17 | shell + all screens | Manual keyboard-only pass; WS stop/start via mock server | Pending |
| NFR-03 | build output | Compare `npm run build` gzip sizes before/after | Pending |
| All phases | `frontend/` | `npm test -- --runInBand`, `npm run lint`, `npm run build` | Pending (not run for Phase 0: no source changes) |

Real household devices MUST NOT be used for routine verification; use
`frontend/tools/server` mocks.

## Debt, risks, exceptions

- Risk: PrimeVue token mapping may not reach every direction's look (e.g. Floorplan walls,
  Deck dials). Mitigation: component-token overrides + `pt` passthrough classes; escalate to
  unstyled mode via ADR-001 amendment only if needed.
- Risk: retro skins hurting legibility. Mitigation: NFR-02 contrast gate; scanline/CRT
  effects optional and off under `prefers-reduced-motion`/`prefers-contrast`.
- Risk: editor rewrite regressions. Mitigation: pure-function core + round-trip fixtures
  before UI work (tasks T020–T022).
- Debt kept: no runtime WS payload validation (FE-02 gap) — recorded, follow-up spec.
- Exceptions: none requested.

## Completion

Phase 0 complete when the owner has reviewed the samples and documents. Each later
phase records actual check output in this plan before merge.

## Implementation design (2026-10-10)

Scope: implement the selected Hearth + Panel prototype in `frontend/` (spec amendment US-06…US-09).
Observed facts that shape it: dashboard groups are a Go map (JSON keys sorted, so no order
survives today); `power_source` values include `"mains (single phase)"`; sign-in is OAuth redirect
only; the mock hub (`frontend/tools/server`) implements HTTP, WS and OAuth.

Note: the `registry/` and `selectors/` folders in the earlier sketch above are implemented as
`src/domain/` modules (protocol registry in `devices.ts`, expose capabilities in `exposes.ts`).

### Layers (single source of truth, FE-01)

| Concern | Owner | Notes |
| --- | --- | --- |
| Device/config/automation truth | existing Vuex modules (`hub`, `automations`, `ws`, …) | unchanged API; bug fixes only (`renameDashboardGroup` mutation, `setDeviceDeConfigfaults` typo) |
| Pure domain logic | `src/domain/` (`devices` (protocol registry), `exposes`, `network`, `dashboard`, `automation`, `activity`, `commands`, `time`, `panel`) | framework-free, Jest-tested; components never re-implement it |
| Presentation preferences | `src/theme/` (`themes`, `theme-settings`) + `useThemeSettings` | `localStorage.nodeherder_theme`, migrates legacy `theme`; no household data (FE-05) |
| Shared behaviour | composables: `useThemeSettings`, `useDeviceCommand`, `useActivityFeed`, `useAutomationDraft`, `useDashboardGroups` | each wraps one domain module + store access |
| Visual primitives | `src/components/ui/` (icon, chip, card, page header, empty, stat) + PrimeVue controls | PrimeVue themed by a token-mapped preset (ADR-001); MDI paths for icons |
| Screens | existing route components rewritten in place; new `overview/`, `panel/`, `automations/editor/` | routes keep their paths; `/overview` and `/panel` added |

Patterns: registry (protocols, expose capabilities), strategy (action editors per action type),
adapter (Vuex ↔ domain in composables), observer (activity feed via `store.subscribe`),
command with pending state (device commands, no retries).

### Contract change: `DashboardGroup.order` (FR-09, NH-04)

- Go `models/settings.DashboardGroup` gains `Order *int \`json:"order,omitempty"\``; pointer keeps
  "absent" distinct from 0. Producers: frontend `saveDashboardGroup`/`importDashboardGroups`.
  Consumers: backend store (persisted in app config JSON), frontend sort, mock hub (passes through).
- Compatibility: old data has no `order` → sorted by name after ordered groups; old frontend
  ignores the field; older backend drops it (session-only order, AC-21). No migration needed.
- Rollback: removing the field leaves stored JSON readable (unknown field ignored by Go).

### Testing strategy (NH-05, [ADR-002](adr-002-browser-tests.md))

1. **Before UI changes**: extend Jest characterization tests for store actions that emit WS
   commands (dashboard groups, device commands, permit join, automations), WS message routing,
   auth service, and add the Playwright harness with flows written against the **current** UI.
   Flows assert wire payloads, so the same specs prove behaviour after the redesign; only the
   page objects change.
2. **During**: Jest tests first for every `src/domain/` and `src/theme/` module.
3. **After**: full e2e at 1280×800 and 390×844; `npm test`, `npm run lint`, `npm run build`,
   `go test ./...` for the backend field.

### Rollout

The owner chose a full redesign, so the legacy-preset switch from the earlier plan is dropped:
`themes/material_blue.js` and `services/theme.service.ts` are removed once nothing imports them.

### Traceability (implementation)

| IDs | Paths | Evidence |
| --- | --- | --- |
| AC-03…06, AC-19…22, FR-09 | `components/dashboards/*`, `domain/dashboard.ts`, Go `settings.DashboardGroup` | Jest `domain/dashboard`; e2e `home.spec`; Go `group_test` |
| AC-04, FR-04 | `domain/commands.ts`, `composables/useDeviceCommand.ts` | Jest; e2e `home.spec`, `panel.spec` |
| AC-07…09, FR-05, FR-07, FR-08 | `domain/devices.ts`, `domain/exposes.ts`, `domain/network.ts` | Jest |
| AC-10…15, FR-06 | `components/automations/**`, `domain/automation.ts`, `useAutomationDraft` | Jest round-trip; e2e `automations.spec` |
| AC-23, AC-24, FR-10 | `components/panel/*` | e2e `panel.spec` |
| AC-25, AC-26, FR-11 | `components/auth/LoginPage.vue` | e2e `auth.spec` |
| AC-27, FR-12 | `domain/activity.ts`, `useActivityFeed`, overview | Jest; e2e `overview.spec` |
| FR-02, AC-02, AC-18 | `theme/*`, `useThemeSettings`, Appearance settings | Jest; e2e `appearance` check |
| NFR-01, NFR-06 | all screens | e2e at two viewports, horizontal-overflow assertion |
| NFR-03 | build | gzip before: JS 641 KB, CSS 66 KB (`4da0245`…`c7284c0` baseline) |

### Evidence (2026-10-10)

Checks run in the dev container against the mock hub; no household devices were contacted.

| Check | Result |
| --- | --- |
| `npm test -- --runInBand` | 21 suites, 136 tests pass |
| `npm run test:e2e` (desktop 1280×800, mobile 390×844) | 26/26 pass: sign-in/out, Home toggle (`deviceSetValue`), area reorder by arrows and drag (`saveDashboardGroup` with `order`), automation list and unchanged round-trip (`saveAutomation` equals the hub payload), device list/page, activity switch, panel mode, no horizontal overflow on 11 screens |
| `npm run lint` | 0 errors (7 warnings in pre-existing `scripts/`) |
| `npm run build` (vue-tsc + vite) | passes |
| `go test ./...` (backend `order` field) | passes |
| NFR-03 bundle (gzip) | JS 560 KB (baseline 641 KB), CSS 62 KB (baseline 66 KB) |

Not run: manual keyboard, screen-reader and offline/reconnect checks at 360/768/1920 (T061),
and tests against a real hub.

## Follow-up plan (2026-10-11)

### Context and design

The main redesign is merged. The owner requests entity rework, a complete history-chart
replacement and measured CRAP remediation/clean-code improvements for everything touched,
and reports delayed illuminance, a troublesome brightness slider and frequent automation
debug dispatch logs. Scope/acceptance: US-10…14, FR-14…19, AC-28…43, NFR-07…08.
Detailed ownership, contracts, UX and investigations: [entities-and-history.md](entities-and-history.md).

Actual baseline: entity detail controls still dispatch directly via `useEntityDialog`;
history has device-only result keys and inconsistent wire types; full and mini charts
duplicate semantic/style logic. Review found offline persistence, duplicate manual runs,
numeric timeout rollback and measured contrast gaps. Existing browser tests cover only
unchanged automation saves and panel navigation, not edits, panel toggles or timeout/recovery.

Use Vuex as confirmed device/query-state owner, pure domain models for entity/history
semantics, narrow transport services and view composables, one command lifecycle, one
metadata formatter/capability policy and one renderer adapter. Retain Vue, Vuex, Router,
PrimeVue and ApexCharts initially. Choose registry/strategy/adapter patterns for concrete
variation; avoid generic frameworks, inheritance or duplicated caches. Responsibilities
are explicitly divided in the design; moving complex logic to unmeasured helpers is not
CRAP remediation.

### Delivery order and dependencies

1. **Preparation and reproductions:** inventory touched functions, establish reproducible
   complexity/coverage/CRAP tooling (including Vue scripts), capture actual HTTP query
   fixtures and build controlled slider/illuminance/automation-log replays. Produce a
   mocked desktop/mobile entity/history prototype using existing tokens before rebuilding
   views. Resolve producer boundaries and diagnosed timing policy before dependent changes.
2. **Reliability and shared entities:** fix diagnosed slider/latency defects, command timeout
   rollback, exact expose targeting and shared confirmation; fix offline persistence and
   manual Run lock, then migrate every entity/detail surface to the shared model/control.
   Preserve the working dial automation with command-payload regression fixtures.
3. **History domain and query layer:** define runtime-checked numeric/state/event unions,
   range/timezone/gap/statistics/decimation policies, keyed Vuex state and bounded HTTP
   request lifecycle. Correct required backend failure handling and mock parity first.
4. **Complete history UX replacement:** new full history toolbar/charts/table/states; numeric,
   proportional state and event strategies; theme-reactive renderer; compact variants reuse
   the same models. Migrate entity details and `DeviceMetrics.vue`, then remove legacy paths
   only after consumer scans and contract tests pass.
5. **Remaining UI and handoff:** Assistant/dialog layouts, editor schedule/validation and
   contrast fixes; keyboard/screen-reader/reconnect/all-width checks, load/bundle comparison,
   per-function CRAP gate and final rule/acceptance review.

CRAP remediation is required in **each** stage for all touched production functions,
including stores, services, composables, Vue callbacks, mock JS and changed Go; it is not
a final optional polish phase. Proposed gates: CRAP ≤ 30, C ≤ 10, ≥ 80% branch coverage
for decision logic and no worsening already-good scores. Formula/mapping/tool versions,
function baselines and successor scores are recorded in this plan during implementation.
Generated reports remain temporary; future build gates should evaluate fresh reports
against thresholds rather than committed snapshots. Tool selection and Vue coverage proof precede numeric baseline claims.

### Contracts, data, rollout and recovery

- Reuse existing `POST /api/metrics/query` through `contracts/api.ts`; use its raw numeric/
  binary response, metadata units and millisecond bounds. This endpoint is currently public
  in `backend/internal/hub.go`; no access expansion is proposed and no claim of existing
  authentication is made. Keep cookies through the existing helper; preserve legacy WS
  consumers. Actual schema and failure fixtures must match Go and the mock.
- `QueryService.Query` currently swallows some collector/read errors. Required fixes
  distinguish legitimate missing-data results from genuine failures; keep HTTP error
  conventions and verify other API/MCP consumers. Bounds are inclusive in the proposed
  client contract; database/tail fixtures must prove compatibility before integration.
- No new persistence schema, retention or household browser storage. Cache at most 10 query
  entries; identity includes device/exposes/UTC range/options, status and generation.
  Request debounce/cancellation/timeout, 10,001-point truncation probe and decimation are
  specified in NFR-08. Query limits do not bound existing backend scans: measure resource
  use; any necessary server cap/pagination requires a specified contract/ADR amendment.
- Display browser timezone/UTC explicitly; do not infer a hub timezone from browser settings.
  The current automation scheduler hard-codes Europe/London; label that zone without
  introducing timezone configuration in this follow-up.
- No change to MQTT commands or automation ordering/event semantics is assumed. Any latency
  fix must identify broadcast vs metrics vs automation timing and a bounded delivery deadline
  before implementation; do not simply remove all debounce or deduplicate events.
- Deploy compatible required backend fixes first, then frontend. Retain legacy metrics
  contracts until consumers are migrated. Rollback to #54 frontend remains possible without
  data migration; backend fixes must remain compatible with that build.

### Constitution Check (follow-up; supported design, execution pending)

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | Amendment US-10…14; explicit assumptions and timing/query preparation gates | PASS (planning); diagnosed timing policy pending before dependent code | Pending |
| NH-02 | One confirmed state/command path; preserve ordered repeated events; latency replay separates sampling/automation/broadcast | PASS (design); timing changes gated on replay/ADR | Pending |
| NH-03 | No new entry point; validate history ingress; actual public endpoint documented; no payload persistence/logging | PASS (design) | Pending |
| NH-04 | Existing Vuex/service/store boundaries; ADR-004, actual producers/consumers and compatibility | PASS (design); dependency/tool choice pending | Pending |
| NH-05 | Traceability below; regressions before fixes; scoped per-function metrics and mocked browser/Go checks | PASS (plan) | Pending |
| NH-06 | Honest missing/partial/gap/unit/timezone semantics; bounded client resources; preserved drafts; rollback | PASS (design); backend scan cost assessment pending | Pending |
| FE-01 | Hub truth, keyed metrics Vuex state, pure domain, narrow service/renderer adapters | PASS (design) | Pending |
| FE-02 | Runtime decoder and Go/mock fixtures; correct boolean and timestamp types; legacy consumer scan | PASS (design) | Pending |
| FE-03 | Matching query completion, errors/timeout/cancel, pending/confirmed/rollback and reconnect | PASS (design) | Pending |
| FE-04 | Token contrast gate, keyboard chart/table alternatives, focus/dialog checks, four widths | PASS (design) | Pending |
| FE-05 | Cookie helpers, text/escaped tooltips, presentation-only preferences, no history persistence | PASS (design) | Pending |
| BE-01 | Required metrics error fixes stay service/domain/repository-owned; handlers adapt existing API | PASS (design, if backend touched) | Pending or N/A with final scope |
| BE-02 | Illuminance/slider/log replay preserves fresh state, repeated events, seed/disabled/availability and automation deltas | PASS (plan, if device pipeline touched) | Pending or N/A with final scope |
| BE-03 | Existing controlled clocks/hubharness; define ownership/shutdown for any timing work; affected race tests | PASS (plan, if backend touched) | Pending or N/A with final scope |
| BE-04 | Query limits/order/empty/failure/boundary fixtures; no schema/retention migration | PASS (design, if backend touched) | Pending or N/A with final scope |
| BE-05 | Existing API exposure/error formats and other query/MCP consumers verified | PASS (design, if backend touched) | Pending or N/A with final scope |

Recheck this table after design and before merge. PASS (design/plan) is not execution evidence,
ratification or maintainer approval. No exceptions are requested or approved.

### Traceability and verification

| IDs | Paths | Proposed verification | Actual result |
| --- | --- | --- | --- |
| AC-28…30, FR-14 | `src/domain/{entities,exposes}.ts`, entity/panel/detail components, `useDeviceCommand` | Capability fixtures; all entry points; zero/missing/unknown/offline; command-confirm/timeout/rollback; dialog keyboard/focus | Pending |
| AC-31…33, FR-15 | `src/domain/history/`, new full/mini history components | Numeric/state/event fixtures, gaps, duplicate timestamps, extrema, empty/partial, unit and DST/range tests; full/mini equality | Pending |
| AC-34…35, FR-16 | `services/history.service.ts`, `contracts/history.ts`, metrics store, `useHistoryQuery`; backend metrics service/API; mock routes | Actual producer fixtures; concurrent/obsolete replies, cancellation/close/logout, loading/error/429, bounded refresh/cache and backend failure-vs-empty | Pending |
| AC-36, NFR-01/02 | chart adapter/tokens/options, UI/chart components | Mock screenshots and keyboard/table inspection at 360/768/1280/1920; screen-reader; all presets/accents contrast | Pending |
| AC-37, FR-17 | automation/dashboards store/composables, trigger service, editor | Offline persistence keeps edits; one POST while pending; refused/error feedback; invalid/schedule times, Europe/London timeline | Pending |
| AC-38, NFR-07 | Every touched TS/Vue/mock/Go function, quality tools/CI | Pinned coverage/complexity reports before/after; source mapping checks; C/CRAP/branch gates and successor mapping | Pending; no scores calculated |
| AC-39 | Assistant/dialogs, E2E and evidence docs | Remaining layouts; actual edited automation save, panel toggle, reconnect, 100 tiles × 10 msg/s, legacy removal scan | Pending |
| AC-40…41, FR-18 | device lifetime/debouncer/controller/WS, hub store/entity views | Controlled illuminance replay timestamps, effective configs, final change then silence, unchanged report, stale feedback; p95 render latency | Pending; backend sampling hypothesis only |
| AC-42…43, FR-19 | `ExposeControl`, command/store path; device/MQTT/controllers/automations | Mouse/touch/keyboard payload vs dial fixture; brightness scale/step/zero/confirmation; dispatch/match/action/echo counts; repeated event/loop tests | Pending; supplied logs do not prove executions |
| NFR-08, NFR-03 | history pipeline, renderer/query store, build | 10k points, 20 warmed timings p95; rendered/cache/request bounds; payload limit caveat; gzip vs #54 | Pending |

Implementation checks: frontend `npm test -- --runInBand`, `npm run lint`, `npm run build`,
`npm run test:e2e`, plus the pinned quality command established in preparation. If Go is
changed, backend `go test ./...`, `go build ./...`, `go vet ./...`, coverage/CRAP and
`go test -race` for affected metrics/API/service/controller/automation packages as applicable.
Use mocks/temp storage/controlled clocks; never routine household-device commands.

Baseline review on 2026-10-11 (#54): 136 Jest tests pass; 26 Playwright tests pass at
1280×800/390×844 after installing missing Chromium; lint has 0 errors/7 existing script
warnings; build passes with large-chunk and duplicate DeviceCard-name warnings. Measured
primary-button contrasts: Fleece light 4.46:1, Daylight light 3.69:1, Teak light 4.40:1
(4.5:1 target). These are baseline findings, not follow-up acceptance results. No backend
checks were rerun in this review. This planning update runs document validation only.

### Debt, risks, exceptions and completion

- Tooling must measure Vue scripts and mock JS; unavailable coverage/scorer support is a
  blocking preparation task for that language. Added dependencies require ADR-004 rationale.
- The raw endpoint can conceal errors and response limits do not bound storage scans;
  capture reproducible failure/cost evidence before integration. Do not invent pagination.
- Sampling can delay the last UI update; fixing broadcast timing must not change automation
  semantics or metrics retention/sampling silently. Logs alone do not diagnose duplicate actions.
- Complete the UX through shared models; do not replace one monolith with a generic framework
  or declare success based on aggregate CRAP/coverage. Unrelated debt stays out of scope.
- Completion requires AC-28…43, remaining original gates, before/after per-function quality,
  truthful acceptance/viewport/keyboard/failure evidence and updated implementation status.
  A human review remains outstanding; no approval or constitutional adoption is invented.



## Device-update repair (2026-10-12)

Scope: T206 backend portion, FR-18/19, AC-40…43. Timing, compatibility, alternatives
and rollback are defined in [ADR-005](adr-005-live-device-updates.md). No household
commands, deployments, dependencies or live-configuration changes were made.

### Investigation findings (2026-10-11)

- `DeviceLifetimeService.Update` accepted fresh state and automation deltas before sampling,
  but metrics and live `OnDeviceUpdated` output shared the same filtered payload. Defaults
  were 60 seconds for numeric measurements and 300 seconds for diagnostics, with device/
  expose overrides and leading-edge sampling. Synthetic reports at 0/2/4 seconds accepted
  100/150/200 while the UI received only 100; a changed report at 61 seconds reached the UI.
  A final suppressed 200 followed by silence, or an identical report at 121 seconds, left
  the UI permanently at 100. This reproduced an application defect independent of sensor lag.
- `useDeviceCommand` waits five seconds for matching reported state and blocks another
  command for that expose while pending. Mouse, touch and keyboard reproductions each sent
  one numeric command per tested gesture. Suppressing the second echo caused a timeout;
  a later command could settle when its echo arrived. `ExposeControl` also retained its
  unconfirmed draft after timeout. The backend dial path reads fresh accepted state and
  bypasses frontend pending state; its fixtures override brightness sampling to zero.
- Automation debug messages were emitted before trigger selection. A synthetic replay of
  100 numeric changes produced 100 debug dispatch lines and zero matching presence triggers;
  disabled automation produced 100 skip lines and no dispatches. Metrics added routine write
  chatter. Dispatch frequency did not prove repeated actions, loops or duplicate subscriptions.
- Supplied approximately 60-second measurement callbacks support the sampling diagnosis
  without identifying changed keys. Repository configuration for `0xa4c13894070052fc`
  listens for presence, with illuminance used as a condition; deployed configuration was
  not verified. Termination signals and WS departure code 1001 do not establish the restart
  cause or explain earlier slider failures. No WebSocket departure handler was changed.

Investigation evidence used synthetic clocks/devices and browser-controlled responses:
three original Go characterization tests (five leaf cases), two command Jest tests and four
browser tests across desktop/mobile passed. Investigation Go/full tests and targeted race
checks passed; Jest passed 22 suites/138 tests with its existing open-handle warning; lint
passed with seven existing warnings. Builds and the whole browser suite were not rerun
at that investigation checkpoint. The Go characterizations are now desired-behavior
regressions in `backend/internal/services/device_lifetime_test.go`. Frontend tests
retain their assertions in `frontend/src/__tests__/composables/useDeviceCommand.test.ts`
and `frontend/e2e/ExposeControl.spec.ts`, matching the source they exercise.

Live effective overrides, device metadata, raw Zigbee2MQTT reports and browser timestamps
remain uncollected. Upstream sensor lag, a lost physical command or an additional deployed
problem cannot be excluded. Browser-controlled echoes are not an integrated MQTT replay.

### Implemented behavior and regression evidence

Live changed exposes and repeated physical events reach the existing update callback in
the same ordered Update call, independently of metrics sampling. Writable same-value reports
can confirm a command without retriggering automation or storing duplicate metrics; read-only
unchanged state stays deduplicated. Metrics intervals/overrides, zero, seed/disabled/unknown/
availability paths, fresh automation state, WS shapes and MQTT targets/payloads remain compatible.
Routine automation/metrics/last-seen logs move to trace, and dispatch wording describes
evaluation; actual matched-action diagnostics and failures remain intact.

Before-fix regressions failed for sampled live values, final-value delivery, writable
same-value/zero reports and debug chatter; after-fix regressions pass. Ordered MQTT replays
assert both stale-then-new and duplicate writable report sequences, with the correct final
state. Repeated physical-event and working dial scenario fixtures still pass.

Intermediate failures exposed old tests expecting sampled UI updates or one writable
broadcast. Unexpected live reports made a test WaitGroup go negative and stalled the
controller suite, exposing a leaked 120-second permit-join callback. The tests now assert
ADR-005's live-report contract; no unrelated permit-join production change was made.

### Backend quality measurements

Inventory: all ten changed production callables below, including every successor extracted
from Update. Unchanged constructors, Seed, availability workers and WS/MQTT producers keep
existing ownership/contracts. No frontend production functions changed at this checkpoint.

| Callable | Before C / CRAP | After C / CRAP | After statement coverage |
| --- | --- | --- | --- |
| `controllers.(*HubController).handleDeviceMeasurementsUpdated` | 2 / 2 | 2 / 2 | 100% |
| `services.(*DeviceLifetimeService).Update` | 16 / 16 | 6 / 6 | 100% |
| `services.(*DeviceLifetimeService).collectUpdate` | New successor of Update | 7 / 7 | 100% |
| `services.acceptExposeReport` | New successor of Update | 3 / 3 | 100% |
| `services.isWritableExpose` | New successor of Update | 2 / 2 | 100% |
| `services.(*DeviceLifetimeService).restoreAvailability` | New successor of Update | 2 / 2 | 100% |
| `services.(*DeviceLifetimeService).attemptToTriggerAutomation` | 3 / 3 | 3 / 3 | 100% |
| `services.(*DeviceLifetimeService).attemptToStoreMetrics` | 6 / 6 | 6 / 6 | 100% |
| `services.getLastSeen` | 3 / 3 | 3 / 3 | 100% |
| `store.(*appStore).StoreMetrics` | 3 / 3.07 | 3 / 3.07 | 80% |


Scorer: `github.com/hbaldwin98/crap/cmd/crap@8eea294df1123778a77ecb826f0b16d95b65768d`,
version `v0.0.0-20260908235400-8eea294df112`, Go grammar v0.23.4. Host: macOS arm64.
The existing executable was `/private/tmp/llm-proxy-crap-bin/crap`; no runtime dependency
or lockfile was added. Baseline coverage used Go 1.26.1; final coverage used pinned Go 1.24.8.
CRAP = C² × (1 − covered fraction)³ + C. Coverage weights executed statements in callable
spans; logical operators contribute to C. Coverage must match the source revision.

All ten callables passed the temporary C ≤ 10, CRAP ≤ 30 and statement coverage ≥ 80%
checker; baseline Update C=16 correctly failed. Existing already-good scores did not worsen.
Raw JSON snapshots and the checker were removed at the owner's request. Future build
threshold gates should generate fresh temporary reports; no such CI gate is implemented yet.

Go coverprofile does **not** measure branch coverage. Tests exercise stopped/active,
known/unknown, event/changed/unchanged, sampled/suppressed, access modes, same-value/zero,
online/offline, enabled/disabled/nil automation, measurement/diagnostic filtering,
last-seen variants and storage success/failure. This is not a numerical branch result.
StoreMetrics retains an unexecuted config-error branch because its concrete cache currently
always returns nil error. Automated branch evidence, Vue script mapping/mounting, TS/Vue/mock
gates and CI integration remain T201 work; no exception or exclusion is approved.

Recorded commands from `backend/` (use the pinned scorer executable as `crap`):

```sh
# Baseline, completed before production edits
go test -count=1 -timeout=5m -coverprofile=/private/tmp/nodeherder-ui-before.cover ./internal/services ./internal/controllers ./store
crap -coverage /private/tmp/nodeherder-ui-before.cover -format json -output /private/tmp/nodeherder-ui-before.json internal/services/device_lifetime.go internal/controllers/hub-controller.go store/store.go
# Final source
GOTOOLCHAIN=go1.24.8 go test -count=1 -timeout=5m -coverprofile=/private/tmp/nodeherder-ui-after.cover ./...
crap -coverage /private/tmp/nodeherder-ui-after.cover -diff-base HEAD -format json -output /private/tmp/nodeherder-ui-after.json internal/services/device_lifetime.go internal/controllers/hub-controller.go store/store.go
```

`HEAD` was the investigation baseline during selection; after committing, use the repair's
base/parent revision instead. Reports and coverage stay in temporary output locations.

### Verification and rule review

| Check (Go 1.24.8 unless stated) | Result |
| --- | --- |
| `go test -count=1 -timeout=5m ./...` with coverage | Passed |
| `go vet ./...` | Passed |
| `go build ./...` | Passed |
| `go test -race -count=1 -timeout=3m ./internal/services ./internal/controllers ./internal/automations/... ./store` | Passed |
| Scoped complexity/CRAP/statement checker | Passed; numerical branch evidence pending |
| `git diff --check` | Passed |
| Frontend full Jest/lint/build/browser suites after backend repair | Unrun; frontend production unchanged |
| Source-named test relocation checks | Passed: Go live-update/log regressions; 2 Jest command tests; 4 desktop/mobile ExposeControl browser tests (Node 24 runners) |
| Browser receipt-to-render timing, full-stack load, manual accessibility, real-device tests | Unrun |

Temporary final logs: `/private/tmp/nodeherder-ui-full.log` and
`/private/tmp/nodeherder-ui-race-final.log`. Original redesign/frontend checks are recorded
in [historical evidence](#evidence-2026-10-10); those are not new repair verification.

NH-01/02/04 and BE-01/02: ADR-005, unchanged metrics semantics, fresh automation-before-output,
repeated-event and ordered MQTT/dial regressions provide scoped evidence. NH-03/BE-05:
existing access/routes/payloads are unchanged. NH-06/BE-04: no schema, units, retention or
migration changes; sampled zero/normalization tests pass. BE-03: no new workers/timers and
affected races pass. NH-05: failing-before/passing-after regressions and final backend
checks pass; frontend/branch/latency/load gates remain incomplete. FE-01…05: browser
implementation is unchanged and live report compatibility awaits integration evidence.
This does not complete AC-38/40…43 or approve merge, governance or ADR ratification.

### Artifact and documentation cleanup

At the owner's request, removed completed-redesign sample pages/assets/generators and
raw quality JSON/checker artifacts. Consolidated investigation/quality evidence here and
progress in tasks; removed the three duplicate documents. Historical design choices and
deferred entity/history drafts remain. Link/reference and diff checks pass; application
suites were not rerun for this documentation-only cleanup.
