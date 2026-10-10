# Plan: Frontend redesign and customisable design system

Spec: [spec.md](spec.md) | Tasks: [tasks.md](tasks.md) | Status: Draft for review
Author: Claude Code | Updated: 2026-10-09 | Reviewed code: `4da0245`
Constitutions: parent/frontend 1.0.0, proposed; none claimed ratified.

Supporting documents: [current state](current-state.md) ·
[design system](design-system.md) · [device & protocol support](device-support.md) ·
[automation editor](automation-editor.md) · [ADR-001 theming](adr-001-theming-tokens.md) ·
[design samples](samples/index.html).

Progress: Phase 0 (documentation + static samples) is complete in this branch.
No production code was changed. Phases 1–6 are **blocked on Q-01** (owner picks a design).

## Context and design

### Observed facts

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

`samples/` holds 5 static, dependency-free designs sharing one mock-data set and one
renderer, each with its own skin and layout. Open `samples/index.html`.

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

Each design has its own layout module (`samples/designs/<id>.js`) over a shared renderer core
(`samples/shared/app.js`: state, contracts, actions, editor field builders `edParts()`), so every
editor emits the same `saveAutomation` payload. In production the equivalent split is shared
composables/selectors plus per-layout route components (FE-01).
Every sample has a live customiser (mode, accent, density, radius, font scale, preset).
Samples are review aids, not production code; they are not built, linted or shipped.

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
| AC-01, AC-02 | `specs/007-frontend-redesign/samples/` | Open `index.html` via `file://` in Chromium; visit every screen in all 5 designs; change each customiser control; reload | Round 1 pass 2026-10-09; round 2 pass 2026-10-10 (5 designs × 13 routes + All pages, standalone files opened alone, editor validation, Brief palette, Deck node drawer, 375 px no horizontal scroll). Playwright + Chromium 1194 via `file://`, fonts blocked (offline fallback). 13 routes × 5 designs at 1440×900 rendered with no page errors; scripted checks per design: tile toggle shows pending then confirms; editor shows inline error + disables Save after adding an empty action; leave guard blocks navigation and "Discard and leave" proceeds; mode switch applies; 375 px width shows no horizontal page scroll on Home/Editor/List. Not checked: Safari/Firefox, screen readers, contrast measurement (NFR-02 is a Phase 1 gate). |
| FR-02, AC-18, NFR-02 | `assets/styles/tokens/`, `composables/useThemeSettings.ts` | Jest: migration + validation; contrast script over preset tokens | Pending |
| AC-03, AC-05, AC-06, NFR-01 | `components/dashboards/*`, `components/ui/*` | Manual at 360/768/1280/1920 with `npm run test-server` mocks; keyboard pass | Pending |
| AC-04, FR-04 | `composables/useDeviceCommand.ts` | Jest with fake store: confirm, timeout, offline; manual with mock server | Pending |
| AC-07, AC-08, FR-07, FR-08 | `registry/*` | Jest: known/unknown protocols, capability resolution for all expose types | Pending |
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
