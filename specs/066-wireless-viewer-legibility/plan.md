# Plan: Wireless viewer legibility

Spec: [spec.md](spec.md) | Status: Draft | Author: agent draft for maintainer review | Updated: 2026-11-20
Constitutions: parent 1.0.0 (Proposed); frontend 1.0.0 (Proposed)
ADR: [adr-001](adr-001-class-groups-and-stable-order.md)

No ratification or reviewer approval is claimed. "Pending"/"Unrun" means not executed.

## 1. Context and design

### Observed current behaviour

- `frontend/src/components/wireless/WirelessViewer.vue` wires the composable to filters,
  radar, cards and grouped list.
- `frontend/src/composables/useObservationViewer.ts` owns filter intent and derives
  `identities`, `visible`, `grouped`, `scoped`.
- `frontend/src/domain/observations/presentation.ts` owns badge/format/grouping helpers.
- `ObservationCards.vue` renders three-line cards; `ObservationGroupList.vue` renders a
  different row anatomy grouped by vendor. Both call `identityTitle`, `estimateDistanceMetres`,
  `trendPresentation`, and `isEphemeralIdentity` directly.
- Order is never decided in the frontend; it is the snapshot array order, produced by the
  backend's `lastSeen` descending sort. Sections are size-sorted.

### Proposed design

#### 2.1 Class groups (`frontend/src/domain/observations/classes.ts`, new)

```ts
export const ObservationClassGroups = {
  Yours: 'yours',
  Someone: 'someone',
  Installed: 'installed',
  Learning: 'learning',
  Outside: 'outside',
  Passing: 'passing',
} as const;

export type ObservationClassGroup = (typeof ObservationClassGroups)[keyof typeof ObservationClassGroups];

/** Ordered. Rank, filter preset order and list section order all come from this array. */
export const CLASS_GROUPS: readonly {
  id: ObservationClassGroup;
  label: string;
  classes: readonly string[];
}[] = [
  { id: 'yours',     label: 'Yours',             classes: ['personal'] },
  { id: 'someone',   label: 'Someone’s device',  classes: ['mobile-unidentifiable'] },
  { id: 'installed', label: 'Installed',         classes: ['infrastructure'] },
  { id: 'learning',  label: 'Learning',          classes: ['learning'] },
  { id: 'outside',   label: 'Outside',           classes: ['neighbour', 'likely-neighbour'] },
  { id: 'passing',   label: 'Passing',           classes: ['transient'] },
];
```

Exports: `classGroupOf(identity)`, `matchesClassGroup(identity, id)`,
`orderIdentities(identities)`, `groupIdentitiesByClass(identities)`, `CLASS_GROUP_ANY`
sentinel used by the filter state. `classGroupOf` looks up the identity's
`classification.class`; absent or unknown values fall back to `learning`, matching
`classPresentation()`.

`orderIdentities` sorts by group rank, then a locale-independent title comparison, then key.
Labels are the existing `CLASS_PRESENTATION` labels. `groupIdentitiesByClass` preserves the
incoming order inside a section (the caller passes an already ordered array) and returns
sections in `CLASS_GROUPS` order with their group id.

#### 2.2 Composable (`useObservationViewer.ts`)

- New state: `classFilter` (`ObservationClassGroup | 'any'`, default `'any'`) and
  `search` (default `''`).
- `orderedIdentities = computed(() => orderIdentities(identities.value))`. Every downstream
  view uses it.
- One predicate, `passesSharedFilters(identity)`, covers triage, distance, facets and search.
- `defaultHidingActive = computed(() => !showDemoted.value && classFilter.value === 'any')`.
  `visible` applies default hiding only when that is true; an explicit class selection
  therefore overrides it (AC-07), and `demotedCount` returns 0 under the same predicate so the
  footer never says "hidden" while the rows are on screen.
- `classGroupCounts = computed(...)`: for each group, a count of ordered identities that pass
  `passesSharedFilters` and belong to the group. Groups with zero counts are omitted (AC-06);
  `Any` is always offered.
- Export `now` so the row can render a reactive, coarse "seen" label.
- `grouped = computed(() => groupIdentitiesByClass(visible.value))`.

The existing `groupIdentities`/`identityGroup` helpers move out of `presentation.ts` (the
vendor grouping is superseded); vendor stays available as a facet and in search.

#### 2.3 Presentation helpers (`presentation.ts`)

- Add `matchesIdentityQuery(identity, query)`: case-insensitive substring over
  `identityTitle(identity)`, `identity.key` and `vendorHex(identity.vendor)`; blank query
  matches everything.
- Add `seenLabel(value, nowMs)`: coarse buckets — "just now" under a minute, "N min",
  "N h", "N d". Invalid timestamps return "—" as today.
- Add `compareIdentityText` (or keep it private to `classes.ts`) that compares lowercased
  titles and then keys with `<`/`>`, not locale collation, so order is deterministic across
  environments.

#### 2.4 Components

- `ObservationRow.vue` (new, presentational): title button (isolate), class badge, distance,
  "rotating" tag when applicable, trend word, seen label, and the existing conditional
  Known/Ignore buttons. One anatomy for both surfaces (AC-17, AC-21). No business logic:
  derived values come from the domain helpers.
- `ObservationList.vue` (new, replaces `ObservationCards.vue`): flat list of `ObservationRow`
  for the radar side panel. Renders the provided empty message.
- `ObservationGroupList.vue`: sections become class groups (from the composable), each with
  count and collapse; rows use `ObservationRow`.
- `ObservationFilters.vue`: add a labelled search input with a clear action, and a
  single-select Class preset group (`Any` + groups with counts) styled like the distance
  presets. Remove the "No source declares filterable facets yet" line, which is only a
  symptom of the missing class dimension; the facet controls remain when useful.
- `WirelessViewer.vue`: wire new state/emits; keep the count line; pass an empty message that
  distinguishes "no matches" from "no observations"; keep the existing toolbar.

#### 2.5 Filter bar layout and theming

- `ObservationFilters.vue` gives every control the same field anatomy — label line, then
  control. The previous bare "Needs triage" button had no label, so it did not share the
  fieldsets' baseline. Fieldset legends do not participate in flex `gap`, so they carry an
  explicit `margin-bottom`; that is what makes the label-to-control distance identical.
- Phones collapse Class/Status/Distance/facets behind a `Filters (N)` disclosure; search stays
  visible. The same fields render inline from 900 px up (`display: contents`), so there is one
  markup tree for both layouts.
- Distance renders as a joined segment on desktop and as wrapping pills on a phone, because the
  joined row was wider than the column and clipped its last bands.
- Class presets are always present, in `CLASS_GROUPS` order, with a `min-width: 3ch`
  tabular-nums count slot; zero counts disable a chip instead of removing it.
- The wireless components are mapped from undefined legacy names (`--surface-raised`,
  `--border-subtle`, `--accent-primary`, …) to the declared theme tokens (`--nh-surface-2`,
  `--nh-border`, `--nh-accent`, …), with fallbacks removed. The light "still gathering evidence"
  box was the fallback winning on the dark theme. `tokens.test.ts` now scans this folder.

#### 2.6 Reading space

- List is the default view; Radar is a deliberate alternative rather than the left half of a
  split. The old rail was capped at 24rem, which truncated every name regardless of row design.
- The Radar view renders the radar alone, adding the Focus panel only when an identity is
  selected (`is-split`). Clearing a selection returns to the view it was opened from, tracked
  in the composable as `lastNonFocusView`.
- Wide rows carry the receiving source names (mapped from `snapshot.sources`) and a signal
  sparkline. `sparklinePoints()` is a pure domain function; `SignalSparkline.vue` renders it and
  draws nothing for fewer than two finite readings.
- `ObservationList.vue` (the flat rail) is deleted; `ObservationGroupList.vue` is the only row
  list surface, so the empty message is passed to it.

#### 2.7 Alternatives considered

- Sort in the backend: rejected. The snapshot order is a documented consumer contract for
  other surfaces; changing it would move the problem rather than solve it, and the viewer's
  presentation intent belongs to the viewer (ADR-001).
- Keep two row anatomies: rejected. The same identity should look the same in both surfaces,
  and duplication is what let them drift (FE-01).
- Filter by presentation tone: rejected because "Yours" would contain "Someone's device"
  (spec Q1).

## 3. Constitution Check

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | spec.md precedes edits; Q1–Q4 resolved, Q5 recorded | PASS (design) | PASS |
| NH-02 | No device state or command; triage write path unchanged | PASS (design) | PASS |
| NH-03 | No new entry point; no new stored data; search runs in memory | PASS (design) | PASS |
| NH-04 | New domain module with ADR-001; no contract change; one-way dependency (`classes` → `presentation` helpers) | PASS (design) | PASS |
| NH-05 | Pure-function unit tests + composable tests + e2e; traceability §5 | PASS (design) | PASS |
| NH-06 | No persistence, no background work; missing classification handled as Learning | PASS (design) | PASS |
| FE-01 | Filter intent stays in the composable; components remain presentational | PASS (design) | PASS |
| FE-02 | No payload change; optional `classification` already validated; no mocks change required beyond e2e assertions | PASS (design) | PASS |
| FE-03 | Empty state distinguishes no match from no data; no invented values | PASS (design) | PASS |
| FE-04 | Search has a label, presets are buttons with `aria-pressed`, focus visible, class stated in words, 1280/390 checks planned | PASS (design) | PASS |
| FE-05 | No secrets; nothing written to browser persistence or logs | PASS (design) | PASS |

## 4. Contracts, data, rollout

- No HTTP/WS/MQTT/MCP schema, route, configuration or persistence change. `contracts/` is
  therefore N/A for this spec.
- Compatibility: an absent `classification` is grouped as Learning (NFR-05); an unknown future
  class value is also grouped as Learning rather than dropped.
- Rollout is a single frontend change. Rollback is reverting the commit; no data migration.
- The backend array order remains untouched (assumption A1).

## 5. Traceability and verification

| Requirement/acceptance IDs | Paths | Test/manual procedure | Actual result |
| --- | --- | --- | --- |
| FR-01, AC-01…06 | `domain/observations/classes.ts`, `components/wireless/ObservationFilters.vue` | `classes.test.ts`: group mapping, counts/omission | PASS — 13 tests; e2e presets 4/4 desktop+mobile |
| FR-02, AC-07 | `composables/useObservationViewer.ts` | `useObservationViewer.test.ts`: explicit class overrides default hiding; "Any" restores | PASS |
| FR-03, AC-08…11 | `domain/observations/classes.ts` | `classes.test.ts`: volatile-only snapshot diff keeps order; new identity inserts; paused order | PASS |
| FR-04, AC-12 | `domain/observations/presentation.ts` | `presentation.test.ts`: seen buckets | PASS — 3 tests |
| FR-05, AC-13…16 | `classes.ts`, `ObservationGroupList.vue` | `classes.test.ts`: fixed section order under count changes; e2e collapse | PASS |
| FR-06, AC-17, AC-21 | `ObservationRow.vue`, `ObservationList.vue`, `ObservationGroupList.vue` | e2e: both surfaces expose the same row roles and triage rules | PASS — 80/80 browser tests |
| FR-07, AC-18…20 | `presentation.ts`, `useObservationViewer.ts`, `ObservationFilters.vue` | `presentation.test.ts` + `useObservationViewer.test.ts` + e2e search | PASS |
| FR-08, AC-22…23 | all changed UI | `npm run lint`, `npm run build`; Playwright desktop + mobile at 1280/390 | PASS — lint 0 errors, build and e2e type-check clean, 80/80 browser tests |
| FR-09, AC-24 | `components/wireless/ObservationFilters.vue` | e2e measures the label-to-control gap for search, Class, Status and Distance | PASS |
| FR-10, AC-25, AC-29 | `ObservationFilters.vue`, `WirelessViewer.vue` | e2e: fields absent from the a11y tree when collapsed, present after opening; full-width view switch | PASS |
| FR-11, AC-27 | `useObservationViewer.ts`, `ObservationFilters.vue` | e2e: every chip keeps its `data-group` order and geometry while counts change | PASS |
| FR-12, AC-28, NFR-06 | all `components/wireless/*.vue`, `theme/tokens.test.ts` | token guard over the wireless folder; no non-`--nh-*` reads remain | PASS |
| FR-13, AC-30, AC-31, AC-34 | `WirelessViewer.vue`, `useObservationViewer.ts` | e2e: list width spans the page; Radar is a separate view with no rail; no overflow at 1280/390 | PASS |
| FR-14, AC-32 | `useObservationViewer.ts` | unit test: selecting from Radar and clearing returns to Radar | PASS |
| FR-15, AC-33, NFR-07 | `presentation.ts`, `SignalSparkline.vue`, `ObservationRow.vue` | `presentation.test.ts` sparkline cases; e2e sees `via <source>` in rows | PASS |
| NFR-01/02 | domain + composable | unit tests; review that no sort runs in a template | PASS — review found no template-side sorting |
| NFR-03/04/05 | whole diff | dependency/bundle review; no storage writes; absent-class test | PASS — no new dependency, no persistence write, Learning fallback tested |

Recorded 2026-11-20: full unit suite 154 suites / 1038 tests passed; lint 23 pre-existing
warnings / 0 errors; build passed; e2e type-check passed; wireless browser suite 80/80 across
desktop and mobile. Backend suites were not run because no backend file changed.

Commands (`frontend/`): `npm test -- --runInBand`, `npm run lint`, `npm run build`,
`npm run type-check:e2e`, `npx playwright test e2e/wireless.spec.ts` (desktop; add
`--project=mobile` for the viewport work). No live device is used.

## 6. Debt, risks, exceptions

- **e2e locator churn.** Renaming the side-panel container and row classes touches many
  locators in `e2e/wireless.spec.ts`. All existing assertions must be updated deliberately,
  not loosened; the suite must still fail if ordering or filtering regresses.
- **Count cost.** Per-group counts scan the snapshot once per group (6 × 256). Acceptable, but
  counts must be computed once in a `computed`, not per render.
- **Collation determinism.** Locale-sensitive comparison would make ordering differ between
  machines; the comparator is lowercased code-unit comparison, tested.
- **Default-hiding interaction.** The footer, the reveal toggle and the class override share
  one predicate; a test must assert they never contradict each other.
- **Undefined theme tokens.** The wireless components had drifted onto names the theme never
  defines, so light fallbacks won on dark themes. The mapping removed the fallbacks and the
  token guard now covers this folder; any new component style must use `--nh-*` names.
- **Chip stability.** Every class group is rendered from a fixed array and the count has a
  fixed slot. A future group must be added to `CLASS_GROUPS`, not conditionally rendered.
- **List/radar trade-off.** Making the list full width means the radar and the list are no
  longer visible together. This is deliberate: the rail could not show a name. Reintroducing a
  split would need a different interaction (for example selection-driven), not a wider cap.
- No exceptions requested. No unresolved blockers.

## 7. Completion

Implement, run the checks, record actual results, update this plan's traceability and
`tasks.md` with evidence. Reviewer approval is not claimed.
