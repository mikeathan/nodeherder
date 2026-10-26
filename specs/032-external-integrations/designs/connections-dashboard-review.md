# Connections landing page: design investigation

Status: recommended grouped board approved and implemented 2026-10-26. Alternative designs remain exploratory.
See the [implementation contract](../contracts/connections-board.md) and [actual verification](../implementation-review.md).
Scope: Connections only. Home, Devices, Overview and Panel stay unchanged.

## Recommendation

Make Connections a data dashboard, not the provider/configuration catalog. Show one automatically
created summary widget per configured connection. Its everyday data and manual Refresh are visible
immediately. Clicking/tapping the card opens the existing full connection page, where larger results,
Setup, Configure operations and Help remain together. There is no Details footer or intermediate
result dialog. Add connection opens the
catalog/setup flow separately. Provider definitions, raw origins and Reload providers live there,
not on the landing page. The first release still offers only bin collections. Weather/energy cards
in the examples are clearly labelled future examples, not new implemented capabilities.

Recommended first design: **Connection board**. Bin collections is one widget containing the three
colour-name/date rows; a future weather connection has its own suitable summary. Reuse styling and
base controls across widget types, without pretending every service is a device or bin schedule.

User-directed revision: match Home's rounded neutral tiles, circular coloured icon badges and
compact text hierarchy. Keep Black bin/Blue bin/Brown bin names and give each a clearly visible
colour-matched icon. Remove resizing and the Compact/Wide selector entirely. Reordering and
hide/restore remain layout proposals; neither creates a second place to edit connection settings.

## Three examples

| Design | What is visible | Benefit | Cost / drawback |
| --- | --- | --- | --- |
| Connection board (recommended) | One summary per configured connection; bins grouped in one widget | Useful immediately; scales to more types; one refresh/settings scope | Requires a small renderer for each capability |
| Individual tiles | Each bin stream gets a tile, alongside other service tiles | Familiar Home-like layout; individual streams can be positioned | Multiple widgets per connection, more clutter and more complex remove/restore semantics |
| Compact overview | One result row per connection, with key results inline | Most space-efficient on phones; easiest to scan/manage | Less visual flexibility |

The accompanying interactive design comparison shows these alternatives, Home-inspired icons,
whole-card navigation to a full page, edit mode/move buttons, removing/restoring a widget, adding
a connection, and a distinct connection-removal confirmation in Setup. All actions/data in that
comparison are local examples. Its reduced provider form illustrates the location of editing;
it does not replace the application's existing complete operations/YAML editor.

## Existing code findings (verified by source inspection)

- [ConnectionsPage.vue](../../../frontend/src/components/integrations/ConnectionsPage.vue) currently
  combines a saved-instance directory with the provider catalog. Saved cards show postcode/status
  and an Open connection button, even though the data is already available.
- [useConnectionsCatalog.ts](../../../frontend/src/composables/integrations/useConnectionsCatalog.ts)
  loads the catalog and all instances together. [Instance/resource DTOs](../../../frontend/src/types/integrations.type.ts)
  already include current validated schedules and stable resource/service IDs. The new landing
  widget does not need a second provider lookup or a new provider API.
- [GroupDashboard.vue](../../../frontend/src/components/dashboards/GroupDashboard.vue) uses the already
  installed vuedraggable package for area ordering, with separate arrow controls.
  [dashboard.ts](../../../frontend/src/domain/dashboard.ts) has a reusable immutable moveItem helper.
  There is no existing dashboard resize engine or arbitrary rectangle collision/layout solver.
- [useDashboardGroups.ts](../../../frontend/src/composables/useDashboardGroups.ts) saves device-area
  groups through the hub settings WebSocket path. Those device-specific models/persistence must not
  hold private connection widgets or resource references. Reuse interaction patterns and controls,
  not Home's data transport or area/device projection.
- [Integration routes](../../../backend/internal/api/integrations.go) already cover instance deletion,
  manual refresh, setup and credentials. There is no connection-board preference endpoint today.
  Existing browser PII remains memory-only; layout preferences must not persist snapshots/tokens
  in localStorage as a shortcut.

## Research and alternatives

Home Assistant's sections view separates dashboard cards from configuration and supports arranging
cards on a grid. Its documented dense placement can change order to close gaps, illustrating why
order-preserving layout is a better default here. Use the general card/section distinction, without
copying its whole dashboard system. [Official sections documentation](https://www.home-assistant.io/dashboards/sections/).

GridStack documents dragging, resizing, responsive columns and layout save/restore. It is a credible
option if arbitrary drag-resize becomes a firm requirement. That capability would still require
Vue ownership/cleanup, breakpoint policy, schema validation, collision behavior, minimum sizes,
keyboard/tap alternatives and persistence tests. It is not justified merely to order a few widgets.
[Maintainer documentation](https://gridstackjs.com/).

W3C's dragging guidance requires a simple pointer alternative to dragging; keyboard support alone
does not cover a person using only taps. Keep visible move-earlier/later buttons (also keyboard
operable), even if drag handles are supported.
[W3C WCAG 2.5.7 explanation](https://www.w3.org/WAI/WCAG22/Understanding/dragging-movements.html).

## Suggested interaction and lifecycle

1. **Normal view:** Connections title, Add connection and Edit layout. Results are visible in
   widgets; no setup paragraphs, provider URLs or per-type catalog sections occupy the board.
2. **Add:** choose a supported type/provider, complete its setup. A successful local Save returns
   to the board and inserts/focuses its widget. A pending lookup produces a pending widget with
   Check status; data is never fabricated while the provider finishes.
3. **Daily use:** view next dates directly. Click/tap anywhere on a card's content to open the
   existing full connection page on Results, with a clear return to Connections. Full-page Setup
   and provider editing remain the sole editing surface. A connection-level Refresh uses the
   existing manual admission/backoff mechanism and must not also navigate.
4. **Edit layout:** drag or use move buttons; remove a widget from the page. No size selector.
   These are draft edits until Save layout; Cancel restores the prior layout. No provider call is
   caused by moving/hiding a widget. Cards do not navigate while layout editing is active. Avoid
   an always-visible control strip on normal cards. On mobile, tap arrows remain available so
   dragging is not necessary for either reordering or ordinary page scrolling.
5. **Restore a hidden widget:** Add connection also lists configured-but-hidden connections with
   Show on page. Its address/token remain configured; no repeated postcode setup is needed.
6. **Delete a connection:** full-page Setup/Remove connection explicitly confirms loss
   of saved address, credentials and results. It uses the existing generation-fenced deletion;
   submitted remote lookups cannot be recalled. Adding it again requires setup. Never label a
   harmless layout operation and a destructive data operation with the same ambiguous delete icon.
7. **Empty board:** a concise empty state and Add connection. If connections are merely hidden,
   offer Show existing connections. Do not imply an empty board means no configured instances.

The user approved the recommended board: remove-from-page hides the widget; full-page removal
deletes the connection with confirmation. These remain separate actions in the implementation.

## Card appearance, navigation and responsive layout

Use Home's visual language: rounded neutral surfaces, subtle theme borders, left-aligned labels
and visible circular icon backgrounds. Keep colour mostly on icons rather than filling whole
connection cards. Black/blue/brown bin rows pair colour with names. Production reuses the
existing WasteBinIcon component/palette inside these badges; the design comparison uses supplied
outline icons. Keep short dates and one quiet check/status line, without a bordered footer.
Do not repeat colour names or coverage warnings under every date; retain source/coverage details
on the full connection page. Existing date/quality view models remain authoritative.

Use an order-preserving CSS grid with automatic responsive columns and standard card dimensions.
No resize selector, resize handles, width preference, size enum or resize-library dependency.
At desktop widths a single connection must occupy one normal grid cell rather than stretch across
the whole page. At smaller widths reduce columns; phones use one column in the same saved order.
Heights accommodate real content, long names, unavailable dates and text zoom. Do not dense-pack
widgets into a different visual/reading order. No fixed heights may hide errors or content.

Each normal card has a real named navigation link whose hit area covers the card, preserving
keyboard focus/Enter and ordinary link behavior. In production its target is the existing private
instance route, not a provider URL. Refresh is a sibling button above that hit area; do not nest
buttons in a link or rely on an article click handler. Edit-layout controls replace navigation
while arranging cards. Returning from the full page restores the originating card's focus and
board position; deleted/hidden instances use a safe heading/Add fallback. Refresh results must not
steal focus or change routes. User-visible stale/error/pending indicators stay actionable.

## Implementation boundaries selected after review

- ConnectionsPage becomes a board coordinator. Catalog/setup opens separately (drawer or child
  route), with configured/available states. Keep the existing protected detail/editor/help paths.
  Clicks route directly to ConnectionDetail; do not build a second details dialog or duplicate
  setup forms inside widgets. An individual-bin alternative routes all three tiles to that same
  instance, optionally selecting/focusing the matching existing service by stable ID.
- A typed widget registry maps capability versions to presentational renderers. Waste summary reads
  the existing schedule view model, preserving Black bin/Blue bin/Brown bin names. It does not own
  transport, scheduling, setup or credential management. Future weather/energy need real contracts
  and adapters before their widgets become available.
- One shared feature composable owns all local board reads, cancellation, stale-reply guards and
  optional minute/on-focus local projection reads. Do not add one polling loop per widget. Page
  reads never trigger provider requests; the first release remains manual refresh.
- A layout composable owns a draft, keyboard/tap ordering and save/rollback feedback.
  Shared accepted connection/layout data remains in Vuex. No new global state/UI framework.
- With **one widget per instance**, optional private Instance board preferences (hidden, order,
  preference revision) can be added as backward-compatible JSON fields. Missing preferences
  mean visible and a stable CreatedAt/ID order. This avoids inventing a separate
  database merely for placements, and deletion naturally removes that instance's preferences.
- Proposed protected GET/PUT /api/integrations/board returns only instance references/preferences,
  with an ETag derived from membership and preference revisions (not resource freshness revisions).
  A batched edit saves all placements atomically in the existing repository transaction. Membership
  changes conflict safely. Provider/credential/lifecycle revisions must remain separate from layout
  revisions so refreshing data does not invalidate layout edits or cancel provider work.
- Validate bounded entries, duplicate/unknown IDs and order before saving. Store only
  references/preferences, not snapshots, addresses, provider definitions or arbitrary HTML.
  All signed-in users share this installation's layout under the existing equal-access policy.
- If **multiple widgets per connection** is selected, use stable widget IDs referencing instance +
  resource/service IDs. That requires a separate layout aggregate/migration design. Do not repurpose
  Home's settings buckets. In particular, the existing encryption AAD includes the storage schema:
  changing that schema without an explicit compatible decrypt/migrate plan would break ciphertext.

The approved [board contract](../contracts/connections-board.md) and ADR-002 fix the preference/API
design. The implementation uses optional per-instance preferences and protected board GET/PUT;
no encryption schema, Home widget, resize feature or multiple-widget aggregate was added.

## Delivery comparison and acceptance gates

| Slice | Relative complexity | Evidence required |
| --- | --- | --- |
| Results on the generic board + separate Add catalog | Small/medium | Empty, pending, saved, stale/error, multiple instances and unknown capability; no second click for everyday data |
| Reorder + hide/restore | Medium | Shared persistence, revision conflicts, cancelled edits, membership/deletion races, mobile order and accessible movement |
| Multiple widgets per connection | Medium/high | Stable service references, partial deletion/restore, replacement identity and migration |
| Card navigation and existing full-page editing | Small | Whole-card/touch/keyboard navigation, Refresh isolation, focus/scroll return, pending/error states; no duplicate editor |

User-confirmed directions: Home-like styling with coloured icons, whole-card navigation to the full
connection page, no Details footer and no resizing control. The user then authorized the recommended
Connection board, hide separate from deletion and one shared host-saved layout. Multiple widgets
per connection and per-browser layouts remain alternatives, not implementation requirements.

Verification plan: source-owned pure layout/widget projection tests, API/repository companions if
preferences ship, isolated browser setup→board, direct data/Refresh, save/cancel/reorder, hidden
restore, destructive deletion, stale/error, two-session conflicts, no PII persistence, 390px/1280px
and keyboard/touch alternatives. Add card-content/keyboard navigation, existing full-page tabs/edit,
Refresh isolation, disabled navigation while arranging, Back focus/scroll restoration and no
remaining size/Details controls. Existing public-route and other-page exclusions remain mandatory.

## Investigation verification

Application code and runtime data were not changed by this investigation. The existing unrelated
automation configuration modification was left untouched. No live provider/device calls were made;
application test suites were not run for this planning-only change. Prototype actions use synthetic
in-memory data; optional host state stores layout choices only, without address/credential payloads.

Revised JavaScript syntax and isolated Chromium checks passed for all three alternatives at 1024px,
736px, 390px and 320px, without surface horizontal overflow. Card-content clicks and keyboard Enter
open a full page; Results/Setup/operations navigation, example name editing and return to the board
passed. Refresh remains on the board, layout mode has no navigation links, cancelled reorder restores
the prior order, and save/hide/restore/delete/re-add flows passed. No size selectors or Details
controls remain. Mobile confirmation fit/Escape passed; no browser page errors occurred. Desktop,
mobile and light/dark screenshots were inspected. Drag-by-pointer and browser-history/scroll
restoration are still production acceptance checks, not claimed as prototype passes.

Rule review: NH-01 is covered by the existing planning set and this design-only revision. NH-02 is
unaffected (no device path changes); NH-03/FE-05 retain private data boundaries and synthetic examples.
NH-04/FE-01 keep navigation, presentation and layout ownership separate, reusing the existing full
page rather than a second editor. NH-05/FE-04 have the explicit prototype evidence above; application
suites and production accessibility checks remain unrun. NH-06/FE-03 preserve manual upstream
refresh, truthful state requirements, draft cancellation and destructive confirmation. FE-02 has no
implemented wire change; proposed layout transport/storage contracts still require a pre-code gate.
This is prototype evidence, not production acceptance or verification of the proposed backend
preference contract.
