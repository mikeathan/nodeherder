# Current frontend: code paths, logic and contracts

Snapshot of `frontend/` at `4da0245` (2026-10-09) for [spec 007](spec.md). Descriptive
only; verify against code before implementing. Paths are relative to `frontend/`.

## 1. Stack and bootstrap

| Concern | Current implementation |
| --- | --- |
| Framework | Vue 3.5 (`<script setup>` mostly), Vue Router 4 (HTML5 history), Vuex 4 |
| Components | PrimeVue 4 (auto-imported via `unplugin-vue-components`), PrimeFlex 3 utility classes, PrimeIcons, `@mdi/js` paths for sensor icons |
| Theme | `src/themes/material_blue.js` PrimeVue preset (660 lines), `darkModeSelector: '.dark'`; 7 global CSS variables in `src/assets/styles/variables.css`; many hard-coded colours inside components (e.g. `#363636`, `#ccc`, `#ff5c5c`, `rgba(0,0,0,.38)`) |
| Fonts | `@fontsource/roboto` 300/400/500/700 |
| Charts | ApexCharts 4 via `vue3-apexcharts` (global `ApexChart`) |
| Events | `mitt` emitter provided as `emitter`; separate dialog/panel event buses in `src/mixins/` |
| Persistence | `vuex-persistedstate`: `auth` → `nodeherder_auth`; `assistant.activeConversation.id` → `nodeherder_assistant`; `localStorage.theme` (`dark`/`light`); `localStorage.remainingTime` (permit join) |

`src/main.ts` creates the app, installs PrimeVue (MaterialBlue preset), ToastService,
store, router, `v-click-outside`, ApexCharts, and the `mitt` emitter. `App.vue` calls
`useTheme().initTheme()` (saved preference, else `prefers-color-scheme`) and renders
`AuthWrapper`, which shows `LoginPage` or `MainLayout`.

### MainLayout startup sequence (`components/layout/MainLayout.vue`)

1. If persisted auth says authenticated: `ws/connect` (if not connected/connecting),
   `fetchHubState()` → `GET /api/hubstate` → `hub/init` (devices + appConfig) unless
   already initialised, `hub/loadMCPStatus`, then `auth/restoreSession` in background.
2. Else: `await auth/restoreSession`; if authenticated do the same; otherwise mark WS
   `disconnected`.
3. Layout = `NavigationDrawer` (side, collapsible, icon rail on desktop, overlay on
   mobile ≤ 768px) + `NavigationBar` (logo, connection dot, theme toggle, version,
   sign-out) + `<main>` containing `PermitJoinTimer`, `Notifications`, `DialogHost`,
   `RouterView`.

## 2. Routes and screens

| Path | Name | Component | Purpose |
| --- | --- | --- | --- |
| `/` | login | `auth/LoginPage.vue` | Username/password + OAuth (`OAuthDialog`, `GoogleIcon`) |
| `/groupdashboard` | groupdashboard | `dashboards/GroupDashboard.vue` | Default post-login page: user-defined groups of `EntityCard` tiles |
| `/devicedashboard` | devices | `dashboards/DeviceDashboard.vue` | Grid of `DeviceCard`s (measurement exposes, footer with last seen/LQI/power) |
| `/devicelist` | devicelist | `device-list/DeviceList.vue` | Paginated DataTable: name, IEEE, last seen, power, actions |
| `/devicepage/:id` | devicepage | `device/DevicePage.vue` | Tabs: About, Exposes, Settings, Metrics |
| `/deviceview/:id` | deviceview | `device/DeviceView.vue` | Single `DeviceCard` (embeddable view) |
| `/viewer` | viewer | `automations/Viewer.vue` | Automation list + status + delete |
| `/creator` | creator | `automations/Creator.vue` | Pick a device to create an automation |
| `/editor/:id` | editor | `automations/Editor.vue` → `DeviceAutomation.vue` | Automation editor (panel stack) |
| `/assistant` | assistant | `assistant/AssistantView.vue` | Chat + history sidebar |
| `/consoleviewer` | consoleviewer | `hub/console/ConsoleViewer.vue` | Live log stream, level filter, remote-logger toggle |
| `/settings` | settings | `hub/settings/Settings.vue` | Tabs: Device Defaults, History, Logger, MCP Server, Assistant |

Guard (`router/index.ts`): `requiresAuth` without auth → `login`; authenticated user on
login → `groupdashboard`. `router/navigation.ts` resolves post-login redirect from
`?redirect=`. Side navigation items come from `mixins/composables/useNavigationItems.ts`
(groups, devices, device list, automations, assistant, console, settings, permit join).
"Permit join" is a nav item that opens a confirmation dialog and then starts the timer.

## 3. State ownership (Vuex modules, `src/store/`)

| Module | Owns | Key getters / actions |
| --- | --- | --- |
| `hub` | `deviceMap`, `appConfig` (hub settings, bridge, dashboard groups, device defaults/overrides), `initialized`, `mcpStatus` | `listAllDevices`, `findDevice`, `findDeviceSetting` (override ?? defaults), `dashboardGroups`, `bridge`; actions emit WS commands (see §4) |
| `ws` | socket, `connectionStatus` (`connected`/`connecting`/`disconnected`) | `connect` (exponential back-off 1s→30s, max 10 attempts), `emit` |
| `automations` | `automationsMap`, `initialized` | `listAll`, `find`; `save`/`delete` commit optimistically then emit |
| `metrics` | per-device query results | `query` → `POST /api/metrics/query` |
| `console` | log messages | `addMessage`; `console-cleanup.service` trims periodically |
| `alerts` | toast queue | `showSuccess`, `showError` (fed by WS `operationSuccess/Failed`) |
| `auth` | user/session | `login`, `restoreSession`, `logout` (`auth.service.ts`) |
| `assistant` | conversations, active conversation | `assistant.service.ts` → `/api/assistant/*` |

`hub/updateDevice` merges a `DeviceUpdate` (`{id,last_seen,availability?,data}`) into a
shallow clone, only touching known exposes, and only replaces the device if something
changed (Vue reactivity). Unknown exposes in updates are ignored.

## 4. Transport contracts used by the browser

### HTTP (`src/contracts/api.ts` `fetchWithAuth`, cookies via `credentials: 'include'`)

| Endpoint | Caller |
| --- | --- |
| `GET /api/hubstate` | `services/hubstate.service.ts` (devices + appConfig) |
| `POST /api/metrics/query` | metrics store (rate limited 4/s server-side) |
| `POST /api/automation/trigger` | `services/automation-trigger.service.ts` (manual "Run"); note: uses raw `fetch`, not `fetchWithAuth` |
| `POST /api/assistant/message`, `GET /api/assistant/conversations`, `GET /api/assistant/history/:id` | assistant service |
| `POST /api/logfile`, `GET /api/listlogs` | console/log files |
| `/api/auth/*` | auth service |
| `POST /api/collect` | **device ingress** for HTTP devices (not called by the UI) |

### WebSocket (`VITE_WS_BASE_URL`, handled in `store/modules/ws/index.ts`)

Browser → backend (`{type, payload}`): `saveAutomation`, `deleteAutomation`,
`deviceSetValue {id,name,value}`, `deviceRename {from,to}`, `deviceRemove {id,force,block}`,
`deviceInterview {id}`, `bridgePermitJoin {permitJoin,maxTimeAllowed}`,
`saveLoggerConfig`, `saveHistoryConfig`, `saveAssistantConfig`,
`saveDeviceConfigOverride`, `deleteDeviceConfigOverride {id}`, `saveDeviceConfigDefaults`,
`saveDashboardGroup`, `renameDashboardGroup {oldName,newName}`,
`deleteDashboardGroup {groupName}`, `importDashboardGroups`, `loadMCPStatus`,
`restartMCP`, `stopMCP`, `startMCP` (backend also defines `loadAutomations`,
`loadDevice`, `loadDeviceList`, `loadDashboardGroups`, `loadAppConfig`, `loadMetrics`,
`deleteAutomationTrigger`).

Backend → browser: `automations`, `automationUpdated`, `deviceList`, `deviceAdded`,
`deviceUpdated`, `bridgeConfig`, `appConfig`, `dashboardGroups`, `metrics`, `logger`,
`mcpStatus`, `operationSuccess`, `operationFailed` (`device` is defined server-side but
unhandled in the browser; unknown types are logged).

Observed gaps relevant to a redesign (not fixed by this spec):

- Payloads are `JSON.parse`d without runtime validation (FE-02).
- Several mutations are optimistic without reconciliation (`GroupDashboard` rename/delete
  has `TODO: this is wrong` comments; `automations/save` commits before ack).
- `saveDeviceConfigDefaults` commits `setDeviceDeConfigfaults`, a mutation that does not
  exist (typo) — the local defaults update silently fails.
- `renameDashboardGroup` commits a non-existent `renameDashboardGroup` mutation.
- Device commands (`deviceSetValue`) have no pending state; UI waits for `deviceUpdated`.
- `GroupDashboard` uses `alert()` for validation errors.
- Reconnect does not refresh device state (noted in spec 006).
- Group dashboard shows raw brightness with a `%` unit (screenshot: "Brightness 254%"); the value is 0–254, not a percentage.
- Backend blocks device-triggered runs of triggers without conditions (`backend/internal/automations/trigger.go`), but the current editor gives no warning.

## 5. Device model and rendering logic

```ts
Device { id, friendly_name, description, connection_type, power_source, last_seen,
         availability: 'online'|'offline'|'unknown', exposes: {[name]: Expose}, properties }
Expose { name, description, unit, data, type: ''|'binary'|'enum'|'numeric',
         category: 'measurement'|'diagnostic'|'config',
         access_mode: 'read'|'write'|'readwrite', attributes, values }
```

- `connection_type` is set by the ingress path: `"mqtt"` for Zigbee2MQTT topics
  (`HubController` MQTT handler) and `"http"` for `POST /api/collect`
  (`backend/internal/api/routes.go` `DataCollectorHandler`). HTTP is how Wi-Fi devices
  (ESP32/ESPHome/Tasmota/Shelly scripts) are integrated today. `ConnectionType.vue`
  maps `mqtt`/`http` to logo images and prints anything else as text.
- `power_source`: backend sets `battery` when the payload has `battery`, else `mains`.
  `getPowerSourceValue` reads `battery`/`battpercentage` or `voltage`.
- Categories are assigned in `backend/models/devices/device.go` (whitelists + access
  bits). Dashboards show `measurement`; automations exclude `config` for triggers.
- Binary on/off uses `expose.values.on/off`; `toggleExposeBinaryProperty` flips it.
- A "state expose" = binary, writable, measurement (`configs/automation/device.config.ts`).
  `EntityCard` icon click toggles the expose if writable binary, else the device's state
  expose; offline devices ignore clicks; numeric `0` renders greyed as "disabled".
- Icons/colours/units/labels come from `modules/formatters/sensor-formatter.ts`
  (`getSensorIcon`, `getSensorName`, `getFormattedSensorValue`) with ~60 MDI mappings.
- Device config (`DeviceConfig`): `disabled`, `metricsEnabled`,
  `defaultDebounceByCategory`, `debounceOverrides`; per-device override or defaults.
  Disabled/offline states render `DeviceStatusOverlay`.
- Controls on Exposes tab (`DeviceExposes.vue`): `Toggle` (binary), `Range`/sliders
  (numeric with min/max attributes), `ButtonGroup`/`Selection` (enum/preset values).
- Metrics tab: period presets (1h…30d), `ChartComponents` map → Numeric/Binary/Area
  charts; mini charts on cards via `useMiniChartData`.

## 6. Dashboard groups

`appConfig.hub.dashboardGroups: {[name]: {name, deviceGroup: {[deviceId]: {deviceId, exposes[]}}}}`.
`GroupDashboard` flattens each group into `EntityCard`s in a 2-column masonry
(`column-count: 2`, max 400px per group). Edit mode (cog) reveals New Group, Export
(JSON download), Import (file → `importDashboardGroups`), per-group rename/delete/add
entities (`DeviceGroupSelectionDialog`), per-card delete. Card click opens
`EntityViewDialog` (history chart for that expose).

## 7. Automations (most complex area)

Data shape (`types/automation.type.ts`): an automation is keyed by **source device id**
(`type: 'device'`), with `enabled`, `schedules[] {startAt 'HH:mm', type enable|disable}`
and `triggers[]`. Each trigger:

- `name` = the source expose that fires it (e.g. `action`, `occupancy`, `contact`),
  `type` = `deviceTrigger` | `manualTrigger`.
- `conditions[]`: `expose {name, equality '='|'<='|'>='|'>'|'<', value}` (binary/enum
  only allow `=`) or `time {timeRange {startAt,endAt}}`.
- `actions[]` (UI allows one per trigger today):
  - `trigger`: target device `id`, `exposes[] {name,data}`, optional `delay`
    (`TimeInterval`), `publishMode` `batch|single`.
  - `step`: numeric increments: `property`, `steps[] {id, property, operator + - *}`, `data`.
  - `preset`: cycle through an expose's preset `values` (`property`).
- Validity: trigger `name` set, ≥1 action, all actions have target id (`isValid`).
- A trigger can be run manually if it has no expose conditions (`canTriggerManually`)
  → `POST /api/automation/trigger {automationId, triggerName}`.

Editor flow: `Editor.vue` → `DeviceAutomation.vue` (id/name/description, status,
ButtonPanel Save/Delete/Schedule, triggers DataTable). Editing uses a **panel stack**
(`controls/Panel.vue` + `usePanelComponents` + `useAutomationsEventBus`): opening a
trigger pushes `Trigger.vue`; adding an action pushes `ActionEditor.vue`, which renders
`TriggerAction`/`StepAction`/`PresetRotationAction`; the scheduler pushes `Scheduler.vue`.
Close pops the last panel. Edits are deep-cloned and written back via callbacks; the
automation is persisted only on Save (`automations/save` → WS `saveAutomation`).
`AutomationStatus.vue` shows Enabled/Disabled/Scheduled and toggles `enabled`.

Usability problems observed (screenshots + code): no overview of "when → if → then";
trigger name is a bare expose dropdown; nested panels lose context; one action per
trigger; Save/Delete disabled logic inverted in `buttonPanelItems`
(`triggers.length == 0 && …`); no validation messages; conditions keyed by `name` in
DataTable (duplicates collide); no undo/discard warning; no test/preview.

## 8. Other screens

- **Assistant**: `useAssistant` composable, `HistorySidebar` (new/select/delete),
  `ChatMessageList` (User/Agent bubbles), `ChatInput`, `EmptyState` (suggestions),
  `NotConfiguredState` (assistant URL or MCP off). Assistant text is rendered as text.
- **Console**: live `logger` WS messages, level `Select`, `Tag` severity, auto-scroll,
  remote logger toggle, cleanup timer.
- **Settings tabs**: Device defaults (disabled, metrics, debounce by category/expose),
  History (sleep timeout, expireAt), Logger (remote, level), MCP (status, start/stop/
  restart, clients), Assistant (URL).
- **Dialogs** (`DialogHost`, one at a time): Confirm, Input, RenameDevice,
  RemoveDevice (force/block), DeviceSelection, DeviceGroupSelection,
  ExposeSelection, Selection, EntityView, OAuth.
- **Permit join**: `PermitJoinTimer` banner with countdown (default 120s), sync with
  `bridgeConfig` WS, persisted `remainingTime`.
- **Notifications**: PrimeVue Toast driven by `alerts` store.

## 9. Shared controls inventory

`components/input`: ActionButton, BaseSlider, ButtonGroup, FillSlider, InputBox,
MultiSelection, PickSlider, Range, Selection, TimePicker, Toggle.
`components/controls`: ButtonPanel, DeviceSelector, Dropdown, ExposeDataInput,
ExposeSelector, Icon (MDI path in circle), Logo, MenuDropdown, NavigationBar,
NavigationDrawer, NavigationMenuBar, Panel, PermitJoinTimer, ReorderableList, Status,
TimeInterval. `components/device`: ConnectionType, LastSeen, LinkQuality, PowerSource,
Sensor, DeviceStatusOverlay. Charts: Base/Binary/DynamicBinary/Numeric (+header),
DateRangeDisplay, AreaChart, Mini{Energy,Percent,Realtime,Numeric,DynamicBinary}.

## 10. Tests

`src/__tests__/` (Jest + ts-jest): store modules (device/hub, automation), types,
contracts, formatters, utils. No component mounting tests; no visual/e2e tests.
