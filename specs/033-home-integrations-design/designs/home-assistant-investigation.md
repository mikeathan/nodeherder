# Home Assistant: cards, sections and integration results

Research: 2026-10-26. The user's single-Household-group objection supersedes the earlier B
recommendation. This document records evidence for the agreed replacement. The revised
[card-first HTML](home-cards.html) demonstrates individual placements; the original comparison
remains historical. This is a synthetic preview, not a production implementation.

## What the actual demo shows

Inspected the supplied [live demo](https://demo.home-assistant.io/#/lovelace/home) in Firefox,
including its desktop screenshot, accessibility hierarchy and edit mode:

- Living room contains individual light, blinds and media-player cards, with temperature and
  humidity beside its heading. Kitchen similarly combines different device functions.
- Energy contains separate EV, Last charge, Home power, Voltage, Fossil fuel and CO2 readings.
  Climate combines Sun, Rain and two thermostat cards. Grouping follows meaning/context.
- Outdoor mixes normal tiles with a larger temperature graph. Individual cards need not all have
  the same height or presentation.
- Editing exposes separate Edit card controls, Add card within each section and Create section.
  There is no required visual container holding every integration together.

These are direct observations, not an inference from a static documentation image. No demo
device controls, card data or configuration were changed. Its Header settings were read without
saving. Browser resizing and cross-section dragging were not manually exercised.

## How the model works

Integration configuration supplies data. Entities represent individual values/functions, including
outputs of information services; they are not restricted to physical devices. Home Assistant's
own explanation explicitly uses municipal waste services with separate waste-stream entities.
[Official concepts](https://www.home-assistant.io/getting-started/concepts-terminology/).

Dashboard cards present those values. A tile normally references one entity; a specialised card
can offer richer content. Grouping cards is itself a separate dashboard choice. Therefore one
integration can feed several cards, while one section can contain cards from different sources.
[Cards](https://www.home-assistant.io/dashboards/cards/),
[Tile card](https://www.home-assistant.io/dashboards/tile/).

The demo's section/card editing model matches the documented Sections view: a grid of sections,
each containing cards, with movement at both levels. Sections need not have a visible title.
Dense placement is optional and can affect ordering; filling every hole is a deliberate tradeoff,
not what makes data reactive. [Sections](https://www.home-assistant.io/dashboards/sections/).

Masonry is a different layout: cards are placed into columns using their sizes, with the next card
below the shortest column. It packs variable-height cards, but grouping requires stack/grid cards.
Do not conflate Masonry with the section grid seen here.
[Masonry](https://www.home-assistant.io/dashboards/masonry/).

The terminology also matters for the wall display: Home Assistant's named Panel view is exactly
one full-width card. NodeHerder's existing wall-panel mode is a different product concept; copying
HA's Panel view would not solve our room navigation requirements.
[Panel](https://www.home-assistant.io/dashboards/panel/).

## What was wrong with our Household proposal

The earlier prototype separated all external-service results into a provider-driven bucket.
Reordering the bucket did not let its contents move among useful contexts. A growing set of
services would make that one column taller and prevent independent placement.

There are two separate requirements: responsive layout must reflow cards/sections at narrow
widths, while reactive data must update their accepted values/status without replacing identity,
order or focus. Splitting cards fixes placement granularity; shared state handles live updates.

## Proposed NodeHerder direction

**Cards are individually placeable; groups are optional context.** Keep the current room groups
and their inline drag/up/down/rename/delete controls, tile-remove buttons and Add tiles flow.
Allow those groups to contain integration-result cards as well as device values. Also allow an
independent integration card to participate beside the room groups when that suits its content.
There is no compulsory Household page or group.

| Result | Suggested card | Possible placement |
| --- | --- | --- |
| Black bin next collection | Standard tile with date/freshness | Kitchen, Garden or an optional Bin collections group |
| Blue bin next collection | Another independently placeable tile | Same group or elsewhere; not tied to Black bin placement |
| Brown bin next collection | Another standard tile | Garden next to outdoor sensor values |
| Complete bin schedule | Optional composite card backed by the same connection | Standalone or inside a user-selected group |
| Weather (future capability) | Compact or wider forecast card | Beside room groups or within Garden |
| Energy price (future capability) | Standard value tile | Beside power/energy device readings |

These placements are examples, not a final selected default. Do not split arbitrary JSON fields
into tiles: each supported capability defines meaningful card choices. The Integrations page can
retain its existing grouped connection summary and configuration flow.

Use shared card sizes and an order-preserving responsive grid first. Wider summaries can have
fixed semantic spans without adding arbitrary resizing. Sections remain intact when the viewport
changes; individual cards can be moved between sections in editing. A standalone card can reflow
among room-group blocks, avoiding a single giant integration column. Dense packing, a free-form
canvas and a resize engine are separate choices and are not justified by this investigation alone.

## Repository implications and panel mode

Current [dashboard items](../../../frontend/src/domain/dashboard.ts) contain only device ID/expose;
[dashboard groups](../../../frontend/src/types/settings.type.ts) group them by device. Add a typed
presentation reference for an integration result: instance + resource + service identity, separate
from its display name/date. A changing next collection must update the same card, not create a
new one. Do not manufacture Zigbee devices, use a service's list index as identity, or place private
integration references in the existing device-settings WebSocket/export path.

The future placement model needs one visual order for device and integration cards, optional group
membership, standalone placements, and independent Home/panel visibility. One hidden flag or one
order per integration instance is insufficient for three independently placed bin-service cards.
The private placement contract must reconcile imports, group renames/deletion and missing services
before production code. Accepted results stay in the existing integrations state; cards must not
own provider refresh loops. Missing/stale/unavailable remain distinct.

[Panel projection](../../../frontend/src/domain/panel.ts) currently converts device-only rooms into
climate and button presentations. Extend its card projection to render read-only schedule tiles
beside device buttons, preserving the selected group and clock/status/exit. A bin tile placed in
Garden should appear on Garden's panel page; a standalone card needs an explicit overview or
page-placement policy. No forced Household panel tab. Its exact policy remains a design decision.

## Recommendation and evidence limits

The agreed card-first revision of the preview demonstrates separate bins mixed into existing
groups, an optional composite schedule, and standalone integration cards reflowing beside rooms.
Preserve the existing edit appearance and heading-count removal. This addresses the actual layout
restriction without requiring a Home Assistant entity-registry rewrite.

The initial research changed documentation only. The subsequent agreement produced the linked
card-first preview, with its own placement, identity/update, responsive, keyboard, panel and
instrumented lifetime checks recorded in plan.md. No production code changed. Actual touch,
assistive-technology, backend persistence/migration and production performance remain unverified.
