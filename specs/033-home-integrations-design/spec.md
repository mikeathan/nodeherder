# Spec: Individual Home cards and shared panel layout

ID: 033 | Status: Implemented and verified | Owner: NodeHerder contributors | Updated: 2026-10-27
Constitution: proposed 1.0.0 | Components: Go settings/API; Vue/TypeScript Home and panel

## Problem, scope, exclusions

Home and panel mode currently project device-only areas. Saved integrations already expose
useful household information on a separate board. Compare ways to bring that information
into Home and a wall panel without pretending an integration is a Zigbee device.

Deliver source-grounded investigation and one standalone interactive HTML containing three
designs. This is new work after merged spec 032; preserve its completed implementation history.
The user explicitly authorized implementation in NodeHerder. Production Home/panel, protected
layout API and compatible persistence are now in scope; no live provider/device testing, provider
configuration changes, automatic provider refresh or additional integration capabilities.

Previous direction: **B, Household area**. The user subsequently rejected the mandatory single
integration group because its contents cannot flow/place naturally with other groups. Investigate
the supplied Home Assistant demo and distinguish cards, sections and integration configuration.
All existing groups remain reorderable. Remove the small numbers beside group names.
The selected card-first model is now implemented in the production app; earlier HTML remains historical.
The supplied current edit-mode screenshot supersedes the previous draft editor: preserve the
single-row handle/title/move/rename/delete header, dashed area outlines, remove-tile badges,
Add tiles and New area/Export/Import/Done toolbar. Retain existing save-as-you-go interaction.
The original HTML is historical evidence of the previous design; AC-01–10 record that prototype.
The user agreed to individual cards and required readable code, single state ownership, useful
comments, engineering patterns, performance verification and lifecycle cleanup. The current
artifact is [home-cards.html](designs/home-cards.html), an offline synthetic preview. Production
Vue/Go implementation was separately authorized by the user’s explicit implementation choice.

## Scenarios and acceptance

### US-01 — Choose a Home and panel design (P1)

Independent verification: open the HTML locally, switch designs/modes, exercise its controls.

- AC-01: Three visibly distinct Home layouts each have a corresponding panel layout using
  the current dark, rounded-card, coloured-icon visual language.
- AC-02: Existing rooms and sample integration results coexist. Panel mode retains room
  navigation, clock, status and exit, with readable integration summaries and touch targets.
- AC-03: A local editor can independently pin/unpin the sample integration on Home/panel;
  Save commits the in-memory draft, Cancel/Escape discards it. No action deletes a connection.
- AC-04: Fresh, stale, offline, pending, confirmed-empty and unavailable examples remain
  distinguishable. Missing data never becomes a fabricated date/zero; room samples remain usable.
- AC-05: The prototype works at 390px and 1280px, with keyboard focus, named controls,
  no horizontal overflow and a light-theme comparison. Future capability examples are labelled.
- AC-06: Investigation cites actual owners, explains layout/data boundaries and compares
  effort/tradeoffs. Recommendations and future contracts are explicitly proposals.
- AC-07: B opens by default. In Edit layout, every group can move earlier/later with named
  keyboard/tap controls or be dragged to another group's position. Changes apply as you go;
  Done/Escape leaves edit mode. First/last movement is bounded.
- AC-08: Saved group order determines B's panel pager and arrow/Home/End navigation;
  hidden Household pages are omitted without changing the relative order of other groups.
- AC-09: Group headings show names only, with no index/count suffix in any alternative.
- AC-10: Match existing AreaCard's inline icon positions and ghost styling: drag before title,
  up/down/rename/delete after title, no separate control row or boxed movement toolbar.
  Keep dashed outlines, tile remove badges and Add tiles; Household has analogous placement-only
  controls. Rename/delete are synthetic local actions; deleting a group never deletes devices
  or an integration connection. Preserve the existing New area/Export/Import/Done toolbar.
- AC-11: Inspect the actual Home Assistant demo and official layout/entity documentation;
  distinguish direct observations from documented behaviour and NodeHerder proposals. Explain
  independent card placement, optional grouping, responsive flow and panel implications, without
  claiming the current Household prototype already implements them.
- AC-12: Black/Blue/Brown bin results are individual cards, each movable into any group or
  directly onto Home without moving the other services. Device cards use the same placement flow.
  Removing placements/groups keeps the source. An optional summary uses the same accepted data.
- AC-13: Changing accepted sample results changes the existing card's value/status, preserving
  its DOM identity, focus, source reference and placement. Missing states never invent dates.
- AC-14: Preserve the current inline group controls, toolbar, save-as-you-go interaction and
  count removal; provide named keyboard/tap alternatives to drag. No block/card overlap or
  horizontal overflow at 390/1024/1280/1440px in either theme, including edit/panel modes.
- AC-15: Panel pages follow accepted block order. Group pages show their cards; standalone
  cards have their own page. Per-card panel visibility does not move/remove its Home placement.
  Clock, status, keyboard navigation and Exit remain available; no forced Household page.
- AC-16: One shared layout observer, zero card timers/polling loops, bounded download resources,
  no retained detached observer targets after repeated page/edit changes. Real page disposal
  disconnects observations, cancels scheduled work and releases listeners/URLs; bfcache is safe.
- AC-17: Imports are bounded and validated before mutation; unknown/duplicate references and
  malformed layouts preserve accepted state. Names are rendered as text. Sample layout exports
  contain references/preferences only, never source readings or dates.
- AC-18: Existing device-only areas migrate without losing names, tile membership or visible
  order. No integration cards are automatically pinned. One canonical private Home layout owns
  accepted placements; legacy device-area payloads are projections, never integration snapshots.
- AC-19: Protected GET/PUT layout requires established authentication and cookie-origin/request
  checks on mutation. If-Match fences changes; stale requests conflict and storage failures leave
  accepted layout/revision untouched. Bounded malformed/duplicate refs cannot partially publish.
- AC-20: Legacy device-area save/rename/delete/import remains functional, adapting to the
  canonical model after migration and preserving unrelated integration placements. Legacy config
  responses/exports never include the private Home layout or integration references.
- AC-21: Integration/provider failure leaves device Home/panel controls usable. New unknown source
  refs are rejected; previously accepted missing refs remain explicit unavailable placeholders.
  Local reads never trigger upstream work. Login change/unmount cancels and fences all replies.
- AC-22: Production uses existing Vuex device/integration sources and one Home layout module;
  reusable domain projections and shared renderers power Home/panel/picker without copied dates.
  Existing device command acknowledgement/offline semantics remain unchanged.
- AC-23: Real frontend and Go checks, meaningful companion unit/transport/transaction tests and
  isolated browser UI cover success/failure/migration/concurrency/lifecycle/keyboard/responsive
  behaviour. Record actual results and limits; no guarantee of zero undiscovered defects.

- AC-24: In panel mode a standalone schedule summary stays compact on wide displays (maximum
  24rem), with readable bin/date rows placed together. It fills the available width on phones
  without horizontal clipping; Home card sizing, navigation and source references remain intact.
  Reproduction: open a standalone Bin collections panel page at 1680px; previously the summary
  stretched edge-to-edge and separated each bin label from its date by most of the viewport.

- AC-25: Refactor the feature's frontend changes against the repo rules: Home renders and wires
  named actions; editor interactions, file IO/ingress, immutable placement commands, legacy
  compatibility, panel presentation and picker choices have cohesive owners. Templates contain
  no persistence/mutation expressions. Public APIs document ownership/side effects, lifecycle
  cleanup remains explicit and accepted state is read-only to consumers. Preserve all agreed
  behavior, including compact panel summaries; verify source-owned tests and isolated UI flows.

- AC-26: Standalone integration cards use the same horizontal and vertical spacing as cards in
  named areas. When a standalone card follows a shorter area, its top aligns with the next tile
  row within 2px. Preserve ordering, drag/edit controls, responsive fit and no-overlap behavior.

- AC-27: In edit mode a standalone card exposes a clear drag handle next to its reorder controls;
  dragging reorders the card among Home blocks and persists the accepted layout. Keep moving
  cards between areas, keyboard alternatives, and busy-state behavior available.

- AC-28: Restore the original compact area edit layout and area drag controls. Standalone
  cards use matching area-style drag and reorder controls; placement controls do not add rows
  of whitespace. Preserve card transfer, keyboard controls, and saved ordering.

- AC-29: In edit mode standalone cards and areas have matching header and outline colors, and
  standalone headers provide a delete action with confirmation that removes only the placement.

- AC-30: In normal Home view standalone cards on the first grid row align with area tiles below
  their headings, while cards below shorter areas retain their tile-row alignment. Recalculate
  when layout or viewport changes and preserve editing, ordering and panel behavior.

## Requirements and success

- FR-01: Supply an offline, dependency-free HTML comparison; all data/actions are synthetic.
- FR-02: Revised Home proposal treats integration outputs as independently placeable cards;
  no mandatory Household container or one-widget-per-integration restriction on Home. Preserve
  bin names, colours and truthful date/check state. The existing Integrations board remains
  one grouped widget per instance. No arbitrary resizing is selected by this research.
- FR-03: Keep connection setup/operations on the existing integration page in any future
  implementation; Home/panel manage placements only. The prototype detail page is illustrative.
- FR-04: Keep source data, layout, projection, mutations, DOM rendering and lifetime ownership
  explicit. Use cohesive named functions, guard clauses, and useful contract/control-flow comments.
  Production must reuse Vuex, authenticated services, domain presentation and existing controls;
  no new state framework or copy of source truth inside cards. See the plan's production gates.
- NFR-01 (offline preview): No network calls, browser persistence, real commands, credentials or personal addresses. Production uses authenticated local APIs and isolated fixtures for verification.
- NFR-02: In the isolated Chromium preview with 18 cards, each of 100 consecutive value-only
  updates completes below 100ms. This is a measured preview target, not a production benchmark.
- SC-01: Historical AC-01–10 retain their evidence; AC-11 has source-linked research; current
  AC-12–17 map to isolated preview checks. Production AC-18–23 require the application evidence recorded in the plan.

## Entities and external interfaces

Observed: dashboard areas contain device IDs/exposes; integration board entries contain instance
references/visibility. These are separate owned models. The private Home layout contract now owns placements while preserving device-only legacy payloads.
Proposed Home/panel placements reference integration instances, never resource snapshots or secrets.
Collection dates belong to Europe/London; fixed demo today is Monday 26 October 2026.

## Questions and assumptions

Assume the request means surfacing the newly implemented external integrations on NodeHerder's
Home page, rather than integrating a separate product named Homepage. Bins are the supported
example. Weather/energy remain future capabilities. B's mandatory Household container was rejected;
card-first placement and app implementation are explicitly authorized. The
[layout contract](contracts/home-layout.md) and [ownership ADR](adr-001-home-layout.md) resolve
the previously open ownership/privacy/migration decisions before production code.

### CI drag and navigation regressions

- AC-31: In Home edit mode, dropping a card over an unavailable/disabled card control completes one atomic placement save, removes the source placement, and survives reload. A failed write retains the accepted layout and feedback.
- AC-32: The blocked-browser-storage navigation scenario distinguishes the panel and Home pin controls and verifies navigation before changing the Home preference.
