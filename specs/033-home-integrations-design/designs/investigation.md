# Bringing integrations into Home

Design exploration, 2026-10-26. Open the current [individual-card preview](home-cards.html).
All data is synthetic; the clock/calendar are fixed to Monday 26 October 2026. No production
changes or external services are needed. Edit layout moves cards between groups or directly onto
Home; each card's Move dialog also controls panel visibility. Integration data demonstrates updates
and missing/failure states. The [earlier A/B/C comparison](home-integrations.html) is historical.

## Current direction

The user rejected the single Household container after reviewing its responsive limitations.
The [Home Assistant investigation](home-assistant-investigation.md) supersedes that grouping
recommendation with individually placeable result cards and optional visual groups. The user
agreed to that model and required engineering/clean-code/lifecycle discipline. The current card
preview demonstrates Black/Blue as standalone cards and Brown inside Garden, with Move/drag
between contexts, optional composite summary and shared panel placement. Source readings are
separate from layout; updated dates patch existing card nodes. See the plan for lifecycle checks
and production ownership/migration gates. The material below records previous design decisions.

## Previous direction (superseded)

The user previously selected **B: Household area**, with every group movable and no numbers beside group
headings. The revised preview opens in B. Household is a logical service group alongside the
physical rooms. Its position is editable just like the rooms, and its saved position determines
where it appears in the panel pager. It is not permanently fixed as the first page.

The user's supplied edit-mode screenshot fixes the layout: preserve the existing AreaCard
controls. Drag handle comes before the title; move up/down, rename and delete are ghost icons
on that same row. Keep dashed area outlines, remove-tile badges, Add tiles and the existing
New area / Export / Import / Done toolbar. Remove only the small group-name count suffixes.
Household uses the same header and an Add integrations tile; its remove actions affect placements.

Use **Edit layout** on Home to drag each group's handle or use the earlier/later arrows. The
first/last movement is disabled at the boundary. Changes apply as you go in memory, matching the
current application's interaction; Done or Escape leaves edit mode with changes kept. This
supersedes the previous sample's separate control row and outer Save/Cancel draft editor.
Individual rename/delete/pin dialogs retain their own Cancel action. The pin dialog saves
Home and panel choices independently.

Rename and group deletion are local interactive examples; deletion confirms that actual devices
and connections are kept. Removing a device tile only removes its placement; Add tiles can
restore the sample values. New area creates an empty sample group. Export/Import exchange only
the sample's group references and labels using a versioned preview-only format. This does not
replace or migrate the application's existing dashboard-groups JSON format. No browser
persistence, actual data changes or provider requests are introduced.

Use one integration instance as the unit of pinning. Black/Blue/Brown bin services stay together,
with names and coloured icons; do not multiply them into fake devices. Home/panel should reuse the
same projection and link to the existing integration Results page. Setup, provider operations and
Help stay in that existing flow. B is selected; production contract/storage design remains separate
from this HTML refinement.

## What the code supports today

| Concern | Actual owner and observed behaviour | Consequence |
| --- | --- | --- |
| Home | [GroupDashboard.vue](../../../frontend/src/components/dashboards/GroupDashboard.vue) counts and renders device-only areas, imports/exports them, and links to panel mode | There is no current slot or pin action for integrations |
| Area contents | [settings.type.ts](../../../frontend/src/types/settings.type.ts), [group.go](../../../backend/models/settings/group.go): name, deviceGroup of device IDs/exposes, optional order | Integration references cannot be passed to the existing area renderer as devices |
| Area writes | [useDashboardGroups.ts](../../../frontend/src/composables/useDashboardGroups.ts) saves through existing hub WebSocket settings actions | Do not put private integration data into this transport or its device-area JSON export |
| Panel | [PanelMode.vue](../../../frontend/src/components/panel/PanelMode.vue) shows one area at a time plus a device/network status page; [panel.ts](../../../frontend/src/domain/panel.ts) selects temperature/humidity and device buttons | Integrations need a separate panel presentation/page type; the current projection cannot render a schedule |
| Saved integration data | [integrations.type.ts](../../../frontend/src/types/integrations.type.ts) exposes instances, resources, operation state and errors | Existing saved results can power initial summaries without a new provider call |
| Summary | [connection-board.ts](../../../frontend/src/domain/connection-board.ts) groups up to three chronological service rows, tracks missing dates and lifecycle status; [waste-schedule.ts](../../../frontend/src/domain/waste-schedule.ts) owns schedule presentation | Reuse its projection and bin title/date logic rather than independently calculating collection dates |
| Current widget | [ConnectionWidget.vue](../../../frontend/src/components/integrations/ConnectionWidget.vue) presents accepted data and emits navigation/refresh/layout intents; its current header also shows postcode | Reuse presentation with an explicit Home/panel context; omit postcode and management actions from glance surfaces |
| Reads/lifetime | [useConnectionsBoard.ts](../../../frontend/src/composables/integrations/useConnectionsBoard.ts), [useConnections.ts](../../../frontend/src/composables/useConnections.ts), [integrations store](../../../frontend/src/store/modules/integrations/index.ts) own auth, cancellation, logout generation, bounded operation observation and local projection reads | Do not mount the whole board coordinator on each card; share a page-owned read scope and accepted Vuex data |
| Preferences | [ConnectionBoard](../../../frontend/src/types/integrations.type.ts) has ordered instance IDs and hidden flags for the Integrations board; [integrations.service.ts](../../../frontend/src/services/integrations.service.ts) saves them with an ETag | Board visibility currently means that page only. Home and panel need separate, explicit preferences |
| Access/navigation | [router](../../../frontend/src/router/index.ts) protects `/panel` and `/integrations/:id`; [integration API](../../../backend/internal/api/integrations.go) verifies authenticated context | Keep protected reads; do not add a public panel schedule endpoint. Return to the originating Home/panel page after detail navigation |

The backend runtime was made capability-neutral in spec 032, but the current browser resource
type is still `WasteSchedule[]` and the widget mapper still understands waste resources. Weather
and tariff cards in the prototype are labelled future examples, not newly available integrations.
They require actual browser contracts, validation and presentation adapters before release.

## Three alternatives, including the panel

| Option | Home | Wall panel | Advantage | Cost / limitation |
| --- | --- | --- | --- | --- |
| **A · Household glance** | A wide, quiet household summary before the existing areas; future cards share the household section | Existing room pager plus a collection reminder. Household in the dock opens full grouped summaries | Smallest visual change; useful information visible immediately; existing area editor remains recognisable | Adds a little vertical space. Multiple household capabilities need a compact card row rather than many full-width banners |
| **B · Household area (selected)** | A service area alongside device areas; all groups can be rearranged; names have no count suffix | The pager follows saved group order; Rooms/Status/Exit remain in the dock | Fits the current mental model of navigating areas, with a natural large panel page | Household is a logical section, not a physical area. Needs a distinct section/page model and explicit ordering; must never be saved as a fake dashboard device group |
| **C · One home board** | Grouped integration cards and room cards share one grid | Overview combines service summaries with room shortcuts; Rooms preserves the pager and Status remains separate | Strongest overview when more capabilities arrive | Broadest layout/editor change; must reconcile existing area order, mixed-card order and mobile reading order. A separate mixed-layout model may be needed |

The B prototype demonstrates whole-group ordering, existing edit controls and independent integration pins.
Its order follows normal grid reading order at each width; all five groups participate, including
Household. Room data/identity remain stable across moves. Panel filters out an unpinned Household
page while preserving the remaining order. There is no arbitrary positioning, resizing, tile-level
reordering or browser persistence in this sample. A/C remain comparison references with clean headings.
No drag/resize library is needed; production should reuse the existing vuedraggable pattern and
explicit arrow alternatives, rather than copy the standalone native-drag implementation directly.

## Proposed production approach for B

B requires one explicit ordered sequence of logical groups (Household plus existing areas),
with stable references and a defined rename/delete/import reconciliation policy. Keep legacy
device-group data and integration references in their existing owners. Preserve old device-area
order during migration; define where the opt-in Household group joins it. Home rendering and panel
navigation must consume the same accepted sequence. The small heading suffixes currently rendered
by AreaCard should be removed in production as well; indices may remain internal for movement.

1. **Separate data from placement.** Keep device areas and integration results in their existing
   owners. Add explicit private placement preferences for Home and panel. An integration hidden on
   Integrations can still be pinned to Home/panel, and unpinning must not disable or delete it.
   Proposed default: opt-in pins for existing instances; upgrading should not silently display
   a configured service on a wall tablet. The prototype begins pinned so the designs are visible.
2. **One summary projection.** Reuse `connectionWidget`/`presentWasteSchedule` and `WasteBinIcon`.
   Extract a focused presentational summary if needed, with Home/panel styles and optional metadata.
   Avoid copying `ConnectionWidget` and its postcode/navigation/refresh assumptions wholesale.
   Only accepted capability versions get richer renderers; unsupported capability remains an
   explicitly unsupported summary, not a waste parser applied to arbitrary JSON.
3. **One owner per active page.** Share accepted integrations state and a single page-scoped read
   lifecycle. Use the existing authenticated service/ingress parsers, cancellation, session guards,
   logout clearing and on-focus/minute local projections. Returning to Home or changing rooms must
   not admit provider work. Keep current manual Refresh/Check status semantics on the detail page;
   repeated mounts must not start repeated provider requests or per-card polling loops.
4. **Protected, revisioned preferences.** For A, independent per-instance Home/panel pin fields in
   the existing private integration repository are a plausible small extension. Keep Integrations
   board order/visibility and its ETag contract intact. Define a separate placement revision and
   atomic update semantics before implementation; provider-refresh revisions must not conflict with
   layout edits. Validate instance references and bound membership; reject stale saves rather than
   dropping a concurrent pin. No endpoint/schema name is selected here. B/C need a wider section or
   mixed-card ordering contract, not just additional booleans.
5. **Safe absence and recovery.** A deleted instance leaves no active placement; a stale placement
   read must not recreate it. An unavailable feature or failed integration read must leave device
   areas operational. Keep last accepted results visibly stale/offline; without results show a
   pending/unavailable placeholder. An empty placement list is distinct from no configured services.
   No snapshots, postcodes, credentials or provider definitions go into localStorage or area exports.
6. **Panel navigation and privacy.** Preserve the current clock, connection state, Rooms, Status,
   exit and room keyboard navigation. Use at least 44px touch targets for panel actions. Show names
   and collection dates without an address/postcode. A detail link should retain a typed, validated
   internal return destination and originating page/room; the current Integrations-board return
   helper is scoped to that board and must be extended deliberately, not reused blindly. Panel mode
   remains authenticated; this visual design does not create a read-only security role or guest mode.

No new library, state system or background provider scheduler is proposed. Layout changes should
have backward-compatible defaults, preserve existing area imports/exports, and roll back to the
current Home/panel with existing results/preferences intact. Any actual storage extension needs
compatibility fixtures and repository transaction tests before release.

## Meaningful data states

- **Saved schedule:** chronological dated rows, explicit check time; a check is not proof of a
  collection taking place. “Tomorrow” is based on the resource's Europe/London calendar.
- **Stale:** keep last dates with a prominent needs-refreshing status and old check time.
- **Offline:** keep accepted dates labelled offline and last checked; do not imply a fresh fetch.
- **Pending:** show lookup in progress/awaiting dates, not a guessed next collection.
- **Confirmed empty:** say None scheduled only when the provider confirms it.
- **Unavailable:** say Dates unavailable when evidence is missing; do not substitute None scheduled.

The demo picks one state for all three bin services to make comparison easy. Production must retain
per-service distinctions and the existing precedence for disabled/error/operation/freshness status.
Actual date-only/DST logic remains in the existing waste-calendar domain functions. The HTML does
not implement a new calendar algorithm. Demo room values remain static across integration failures.

## Verification and remaining implementation gates

Prototype results are recorded in [plan.md](../plan.md). Validate each Home and panel at 390/1024/1280px,
both themes and every data state; independent pin save/cancel/Escape; room keyboard navigation;
status/exit; synthetic detail/back/refresh; absence of HTTP requests/browser persistence; and inspect
screenshots. App suites/build/lint are not relevant to a standalone design artifact and were not run.

Before production: specify B's placement defaults/order/export compatibility and
private storage/API semantics; map all requirements and failure paths to source-owned tests; verify
deletion/membership conflicts, logout/cancellation and service-unavailable isolation; run required
frontend checks, isolated browser checks and affected backend checks. Check at 390/1280px plus wall
tablet landscape/portrait, text zoom, keyboard and touch. No routine checks may command devices.
