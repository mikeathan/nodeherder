# Spec: Wireless viewer legibility — class filter, stable order, unified list

ID: 066 | Status: Draft | Owner: agent draft for maintainer review | Updated: 2026-11-20
Constitution: 1.0.0 (Proposed) | Components: frontend
Follows: [spec 065](../065-observation-classification/spec.md), [spec 064](../064-wireless-observations/spec.md).
This spec states outcomes. It does not claim ratification, reviewer approval, or
implementation authorization.

## Problem, scope, exclusions

### Problem

Spec 065 made behavioural classification real: every identity carries an advisory class,
and the default view hides installed equipment and probable neighbours. Against a real
install (about 62 identities heard for a day and a half) three things still make the viewer
hard to use.

**The classification cannot be filtered on.** The filter bar offers triage, distance bands
and source-declared facets. Class is not a source facet, so it is never offered. The only
class control is the binary "Show all kinds". An operator who wants to see "just the things
it thinks are mine" or "just the ones still learning" cannot ask.

**The list reorders constantly.** The frontend never decides order: it renders
`snapshot.identities` in the order received, and the backend sorts by `lastSeen` descending
(`backend/internal/observations/service.go`). `lastSeen` advances on every advertisement and
the snapshot is republished on a 250 ms tick, so each device jumps to the top as it emits.
The list view adds a second shuffle by sorting sections by member count
(`frontend/src/domain/observations/presentation.ts`). Rows moving under the pointer makes
triage and reading practically impossible.

**The list is tall and inconsistent.** The radar side panel renders three-line cards with
always-visible triage buttons (`ObservationCards.vue`); the List view renders a different,
denser row anatomy grouped by vendor (`ObservationGroupList.vue`, `identityGroup`). Vendor
grouping is opaque — "Vendor 0x004C" does not answer a question the operator is asking — and
the two anatomies mean the same device looks different depending on which view is open.

There is also no way to find a known identifier among dozens of mostly opaque keys.

### Outcome

- The viewer filters by behavioural class, with one-click presets and honest counts.
- Row order is stable between snapshots: nothing moves unless the operator changes a filter
  or a deliberately rare event occurs (a new identity appears, or a class changes).
- The List view groups by class in a fixed order, and the radar side panel and List view
  share one compact row presentation.
- A text search finds a device by name, key or vendor.

### Scope

Frontend only: the wireless viewer, its composable and the observations presentation domain.
Classification is already present in the snapshot; no payload or contract changes.

### Exclusions

- No backend ordering change. The snapshot's array order remains "most recently heard
  first" for other consumers; the viewer orders for display at its own presentation layer.
- No Tracked/Census two-surface split or Timeline pane from 065 ADR-003. This spec keeps one
  surface and adds filters; those are separate future work.
- No persistence of filter/search state across reloads. Existing viewer intent is
  component-local by design.
- No change to identity resolution, triage semantics, distance model, or class rules.

### Affected users and data

Operator on an authenticated browser session. No new data is collected or stored; filtering
and search run over the snapshot already in memory.

## Scenarios and acceptance

### US-01 — Filter by what the classifier says (priority P1)

Independent verification: unit tests over the presentation/composable, plus a browser check
with the mock snapshot.

- AC-01: Given identities in several classes, when the viewer loads, then a Class preset group
  offers one option per class group present plus "Any", each showing its matching count.
- AC-02: Given a class preset is selected, when the viewer renders, then the list and radar
  contain only identities in that class group, and the "X of Y shown" count agrees.
- AC-03: Given a class preset selected, when "Any" is selected, then every identity that the
  other filters allow is shown again.
- AC-04: Given a class filter, when a distance band, triage filter, facet or search is also
  applied, then all predicates combine (logical AND).
- AC-05: Given an identity with no classification, then it belongs to the Learning group.
- AC-06: Given a class group has no identity under the other active filters, then its preset
  remains in place showing `0` and is disabled (unless it is the active preset).
- AC-07: Given a class group whose identities are hidden by default (Installed, Outside), when
  that class preset is explicitly selected, then its identities are shown without first
  operating the separate "Show all kinds" control; selecting "Any" restores the default hiding.

### US-02 — The list stops moving (priority P1)

Independent verification: unit tests with two snapshots differing only in volatile fields;
a browser check is not sufficient because it needs controlled snapshots.

- AC-08: Given two snapshots whose identities differ only in `lastSeen`, `packets`, `rssi`,
  `rssiSmoothed`, `trend` or `signal`, then the rendered row order is identical.
- AC-09: Given unsorted input, then order is class-group rank first, then a stable comparison
  of the display title, then key as the final tiebreak.
- AC-10: Given a new identity, then it appears in its sorted position and no other pair of
  existing rows changes relative order.
- AC-11: Given the viewer is paused, then the frozen order is the order captured at pause and
  does not change until resumed.
- AC-12: Given a "seen" label, then it is expressed in coarse buckets (under a minute, whole
  minutes, hours, days) and does not change every snapshot or every second.

### US-03 — The List view is grouped by class, in a fixed order (priority P1)

- AC-13: Given the List view, then sections are class groups in the same fixed order as the
  filter presets, each with a count and a collapse control.
- AC-14: Given section counts change, then sections do not reorder.
- AC-15: Given a section is collapsed, then its members are hidden and the count remains.
- AC-16: Given vendor grouping is removed, then vendor remains available as a facet filter and
  the vendor is still discoverable through search and Focus.

### US-04 — One compact row and a search (priority P2)

- AC-17: Given the radar view, then the side panel renders the same row anatomy as the List
  view (title, class badge, distance, trend, seen, triage), just without section headers.
- AC-18: Given a search query, then identities match case-insensitively on display title, raw
  key or vendor hex; matching identities are shown and non-matching ones are excluded.
- AC-19: Given a search query with no matches, then the list states that no identities match
  and offers a clear action; the radar and counts reflect the empty set.
- AC-20: Given the search is cleared, then the previous filters apply again unchanged.
- AC-21: Given a row already known or ignored, then its redundant triage button is not shown
  (preserving the existing rule that a no-op control is not offered).

### US-05 — Works on desktop and mobile (priority P1)

- AC-22: Given 1280 px and 390 px viewports, then the page has no horizontal overflow and the
  class presets, search and rows remain operable and keyboard reachable.
- AC-23: Given a filter state, then class is conveyed by words as well as colour (FE-04).

### US-06 — The filter bar is clean and stable (priority P1)

Reported after the first implementation: "Needs triage" did not line up with the labelled
fields; the phone view was a noisy stack of unrelated groups; the joined distance control
clipped its last bands; and class chips moved as their counts changed.

- AC-24: Given the desktop filter bar, then every field places its control the same distance
  below its label, whether the fields share a row or wrap.
- AC-25: Given a phone viewport, then search stays visible and Class/Status/Distance/facets sit
  behind a "Filters" disclosure that shows the number of active filters; collapsing it removes
  those controls from the accessibility tree.
- AC-26: Given a phone viewport, then every distance band is reachable and the bands wrap
  instead of overflowing or clipping.
- AC-27: Given any change to a count, then every class preset remains present, in the same
  order and at the same position and width; a zero-count group is shown disabled rather than
  removed.
- AC-28: Given the wireless components under any theme, then every custom property they read
  is defined by the theme layer; a hardcoded light fallback must not paint a light surface on
  a dark page.
- AC-29: Given a phone viewport, then the Radar/List switch occupies its own full-width row
  instead of sitting stranded beside the count.

### US-07 — Read the list without truncation (priority P1)

Reported with a screenshot of the Radar view: the side rail was capped at 24rem, so every name
truncated ("Dinin…", "fp:13…"). The reference design is a full-width list with room per row.

- AC-30: Given the viewer opens, then List is the selected view and its list spans the page
  width, not a fixed side rail.
- AC-31: Given the Radar view, then the radar is the only main pane unless an identity is
  selected; no truncated row list is rendered beside it.
- AC-32: Given an identity selected from List or Radar, when the selection is cleared, then the
  view it was selected from is restored.
- AC-33: Given a row whose identity has source readings and a signal history, then the row names
  the receiving source(s) and draws a sparkline; missing data is omitted rather than drawn as an
  empty line or a fabricated value.
- AC-34: Given 1280 px and 390 px viewports, then the wider rows do not overflow and the phone
  layout still stacks.

## Requirements and success

- FR-01: MUST offer a single-select Class preset group derived from the identities present,
  with "Any" and per-group counts. Acceptance: AC-01…AC-06.
- FR-02: MUST treat an explicit selection of a default-hidden class as an override of default
  hiding for that selection only. Acceptance: AC-07.
- FR-03: MUST order rows deterministically by class-group rank, then title, then key, and MUST
  NOT use volatile fields for ordering. Acceptance: AC-08…AC-11.
- FR-04: MUST render relative "seen" time in coarse buckets. Acceptance: AC-12.
- FR-05: MUST group the List view by class group in fixed order. Acceptance: AC-13…AC-16.
- FR-06: MUST use one presentational row component for the radar side panel and the List view.
  Acceptance: AC-17, AC-21.
- FR-07: MUST filter by a free-text query over title, key and vendor, with an empty state and
  a clear action. Acceptance: AC-18…AC-20.
- FR-08: MUST preserve keyboard access, accessible names, visible focus and non-colour status
  cues on every added control. Acceptance: AC-22, AC-23.
- FR-09: MUST give every filter one labelled field anatomy, so controls align by construction.
  Acceptance: AC-24.
- FR-10: MUST collapse the filter fields behind a labelled disclosure on phones, with a count
  of active filters. Acceptance: AC-25, AC-29.
- FR-11: MUST render every class preset at all times in fixed order with a fixed-width count
  slot, disabling zero-count groups. Acceptance: AC-27.
- FR-12: MUST style wireless components with declared theme tokens only, never undefined names
  with hardcoded fallbacks. Acceptance: AC-28.
- FR-13: MUST make List the default, full-width surface and give Radar its own view rather than
  a shared split. Acceptance: AC-30, AC-31, AC-34.
- FR-14: MUST restore the view a detail was opened from when the detail is cleared.
  Acceptance: AC-32.
- FR-15: MUST name the receiving source and draw a signal sparkline in a wide row, omitting
  absent data. Acceptance: AC-33.
- NFR-01: Ordering, grouping, class mapping and search matching MUST be pure functions in the
  observations domain, independently unit-tested. No sorting inside a Vue template.
- NFR-02: The added filtering, grouping and sorting MUST be computed from memoised `computed`
  state; at the snapshot ceiling (256 identities) it must not add noticeable work per tick.
- NFR-03: No new runtime dependency. No backend, contract, persistence or storage change.
- NFR-04: Filter and search state MUST remain component-local; nothing is written to browser
  persistence or logs.
- NFR-05: A missing `classification` MUST be handled as Learning, never as an error or a
  hidden row beyond the existing default rule.
- NFR-06: The theme-token guard MUST cover the wireless components, so an undefined property
  cannot reintroduce a fallback-coloured surface. Acceptance: AC-28.
- NFR-07: Sparkline geometry MUST be a pure function in the observations domain, unit-tested;
  a component only renders its output. Acceptance: AC-33.

## Entities and external interfaces

| Entity | Owner | Meaning |
| --- | --- | --- |
| Class group | frontend presentation domain (new) | Ordered bucket of one or more advisory classes used by filter, order and sections |
| Filter state | `useObservationViewer` | Operator intent: class preset, search, distance, triage, facets |
| Row presentation | new presentational component | One identity rendered identically in both list surfaces |

No external interface changes: no HTTP/WS/MQTT schema, route, storage or configuration change.

## Questions and assumptions

| ID | Question | Consequence | Owner | Resolution |
| --- | --- | --- | --- | --- |
| Q1 | Filter granularity: exact class labels or coarse tones? | Determines preset count and wording | Maintainer | **Resolved 2026-11-20: six exact labels** — Any, Yours, Someone's device, Installed, Learning, Outside, Passing. Only the two neighbour classes merge under Outside. |
| Q2 | Sort key: class then name, name only, or distance? | Determines whether the list can reorder | Maintainer | **Resolved 2026-11-20: class group, then title then key.** Distance sorting was rejected because RSSI jitter reintroduces movement. |
| Q3 | One compact row everywhere, or keep cards beside the radar? | Determines row anatomy and shared component | Maintainer | **Resolved 2026-11-20: one compact row everywhere**; sections only in the List view. |
| Q4 | Add text search? | Adds one control and an empty state | Maintainer | **Resolved 2026-11-20: yes**, over title, key and vendor. |
| Q5 | How do class presets interact with the existing default hiding? | A preset that filters to an empty list would look broken | Agent proposal for review | **Chosen: an explicit class selection overrides default hiding while active; "Any" restores it.** Recorded in [plan.md](plan.md) §2.3. |
| Q6 | How should the filter bar be arranged, especially on a phone? | Determines whether fields collapse and how they align | Maintainer | **Resolved 2026-11-20: two labelled rows on desktop; collapsed behind a "Filters" disclosure on phones.** Class chips stay permanently present with a fixed-width count. |

### Assumptions

- A1: The backend's array order stays as-is; only the viewer re-orders. Recorded so a future
  consumer does not assume the snapshot order is stable.
- A2: The operator's classification is advisory and may be absent or provisional; grouping
  must tolerate both.
- A3: "Outside" legitimately covers neighbour and likely-neighbour, because both answer
  "probably not in my home" and both are demoted by default.
