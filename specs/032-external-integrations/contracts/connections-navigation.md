# Connections navigation and page boundaries

Status: M1 implementation contract; later-milestone provisions remain deferred. See [implementation review](../implementation-review.md).
Spec: [../spec.md](../spec.md) | [Plan](../plan.md)

## Contained first release

Connections is the user-approved top-level section, immediately after Devices in the existing
sidebar/mobile navigation. All connection discovery, setup, results, health, provider editing
and Help live inside this section. Existing authenticated direct routes remain available.

Home, Devices, Overview and Panel keep their existing content, layouts, counts and controls.
Do not add connection tiles/cards, room placement, pin controls, summaries, banners, explanatory
copy, cross-promotion links/buttons or a connection-specific global footer. The shared main
navigation entry is sufficient. The earlier Home/Garden/Panel placements were exploratory samples;
the user narrowed scope to Connections only. They are not implementation requirements.

No connectionTiles dashboard field, dashboard import/export migration, settings-cache writer
change, room projection or PanelConnectionTile renderer is needed. Keep integration resources and
private state out of AppConfig, dashboard settings and device snapshots. M1 uses protected REST
reads; no integration payload or reference is added to the global settings WS stream.

## Product vocabulary and setup

One installation is one shared home; no new household selector, identity heading or breadcrumb.
Every signed-in user has equal access and shared sessions, with revision conflicts handled safely.

| Concept | First example | Purpose |
| --- | --- | --- |
| Connection type | Waste collection | Groups a useful feature and compatible providers |
| Provider | WhenIsBins | Shared host YAML describing supported requests/mappings |
| Configured connection | Bin collections — selected address | Saved household settings/private binding |
| Resource | One waste schedule containing several services | Validated data for Results and later consumers |

The landing page shows configured connections and supported types/providers. Add connection starts
provider selection, postcode entry, explicit address choice and save. Only waste collection ships
initially. Weather/energy placeholders must be labelled unavailable or omitted. No marketplace is
needed. Normal setup does not require opening Configure operations.

A configured connection has Results, Setup, Configure operations and Help. Provider edits affect
all instances using that definition; show their scope/affected count before activation. Address
changes affect only that configured connection. Internal APIs, packages and YAML directories may
remain named integrations; the visible product label is Connections.

## Authenticated browser routes

```text
/connections                                      catalog and configured connections
/connections/setup/:id                            new/edit shared setup session
/connections/:id                                  results and connection settings
/connections/providers/:type/:provider             shared Form/YAML editor
/connections/providers/:type/:provider/help        bundled guide and contextual anchors
/connections/:pathMatch(.*)*                       feature-owned not-found view
```

The existing authentication guard applies to every route. Unknown/mismatched type or instance
routes return a useful not-found page. The backend keeps /api/integrations. Help/Back preserves
draft state, focus and scroll. Direct Help/editor links work without an installed instance.

## Existing source ownership and scope

- `frontend/src/components/shell/navigation.ts` owns menu order. Extend it, `AppShell.vue`,
  `frontend/src/types/router.ts` and `frontend/src/router/index.ts` for the one new section,
  active-section membership and lazy protected pages. Preserve every existing menu item.
- `frontend/src/components/dashboards/GroupDashboard.vue` / `AreaCard.vue` remain device-area/tile
  views. Their models, helpers, cache writers, imports/exports and Panel entry are not changed.
- `frontend/src/components/overview/OverviewPage.vue` retains its Devices online, Need a look,
  Automations on, Home areas and MCP server stats, ActivityFeed, attention and Network panels.
  Provider health belongs in Connections, not existing device/network counts or activity.
- `frontend/src/domain/panel.ts` / `components/panel/PanelMode.vue` keep their climate, room,
  device-button, status and command behavior. No new data or links appear in a room.
- `frontend/src/components/devices/` receives no feature promotion, notice or shortcut. Exact
  device page owners remain unchanged; only the shared shell exposes the new menu item.

Reuse Vuex, authenticated services, existing controls/theme and ingress validation in Connections.
No new UI/state framework, device impersonation or dashboard refactoring is needed.

## Verification and future presentation

AC-23 covers the new menu/catalog/route behavior. AC-26/27 cover unchanged pages and absence of
cross-promotion. T028 owns those boundaries; T029/T030 and AC-24/25 are retired rather than reused.
Check protected routes, menu highlighting, mobile drawer behavior and normal setup; use negative
browser assertions on Home/Panel/Devices/Overview content and review their production diffs.
Isolated application browser fixtures must verify existing Home/Devices/Panel content contains
no new connection data, controls or prose; those pages are not redesigned by this feature.

A future dashboard/panel consumer may be designed separately if requested. No room association,
widget schema, renderer, implementation task or promised milestone is reserved in this release.
Stable resource contracts still support the already planned later assistant/automation work.
