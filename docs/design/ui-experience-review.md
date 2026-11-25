# NodeHerder UI experience review

**Status:** exploration only — no production code changed. Mocks are static HTML.
**Companion mocks:** [`docs/design/mocks/ui-improvements.html`](mocks/ui-improvements.html)
(screenshots in [`docs/design/mocks/shots/`](mocks/shots/)) and
[`docs/design/mocks/ring2.html`](mocks/ring2.html)
(screenshots in [`docs/design/mocks/shots-ring2/`](mocks/shots-ring2/)).
**Scope:** Home cards, Console/logs, alerts and new-session state, Settings, automation actions.
**Reviewed against:** current `frontend/src` (Vue 3 + Vuex + PrimeVue) and the relevant backend
routes; project and component constitutions apply to any follow-up implementation.

---

## 1. The short version

Four problems explain most of the friction:

1. **The UI cannot tell “online” from “still reporting”.** Availability drives the status badge;
   `last_seen` exists but is not used on cards, and the store has no snapshot timestamp. A device
   that goes quiet — or a browser whose socket dropped — keeps looking live. That is the exact
   failure the user hit during the WebSocket incident.
2. **Anything transient is lost.** Alerts are three-second toasts with no history. A failed
   automation, a dropped socket or a command error is gone if you were not looking at the right
   corner of the screen. There is no inbox, no badge, no “while you were away” summary.
3. **Settings is read-only where it matters.** Storage reports 28 MB but cannot purge; logs fill
   the disk but cannot be downloaded or pruned from the UI; Home and automation exports exist in
   code but have no entry point. Any index must also stay additive: all six existing sections,
   including MCP and Assistant, keep every control.
4. **Mobile is “desktop, narrower”.** Tables, edge menus, side-by-side destructive buttons and
   three-column chips do not survive 390 px. The same features need sheets, stacked rows and
   full-width controls — designed now, not retrofitted.

Recommended kit (see mocks):

| Area | Pick | Why |
| --- | --- | --- |
| Home cards | **1A** age line per tile, plus the **1C** paused banner | fixes trust without a new mental model |
| Logs | **2C** Settings home + **2B** global drawer | quick access stays one click; retention/export get a real home |
| Alerts | **3A** inbox + **3B** new-session digest | nothing actionable is lost; no toast replay |
| Settings | **4B** dashboard index + current rail for detail | discoverability with no section or control removed |
| Cleanup | **4D** Data & cleanup | answers “how do I clear X?” |
| Automations | **5A** row menu + export all | small change, high utility |
| Overview | **Ring 2 O-A** calm verdict + three actions | removes nine competing blocks |
| Panel mode | **Ring 2 P1–P3** day / night+offline / status | 2 m glance surface with its own rules |
| Mobile | **M1–M8** | every area redesigned for 390 px, not squeezed |

Do **1A/1C + the inbox producers** first: they fix the reported blind spot. The rest is
sequenced in §5. Mobile is not a later phase: every recommendation below has a mobile shape in
the gallery (§7, M1–M8), and the acceptance criteria in §3.11 apply from the first implementation.

---

## 2. Findings

Severity is about user impact, not code size.

### F1 — A tile’s status ignores reporting silence · **High**

`EntityTile.vue` derives `status` from `statusOf(device)`; `domain/devices/index.ts:66` returns
`offline` only when `availability === 'offline'`, `online` when availability says online, and
`unknown` otherwise. The tile prints “Offline” only for that flag. `device.last_seen` is available
and is already shown in the Devices list and Overview attention list, but Home tiles ignore it.

Consequence: a sensor whose availability has not flipped (common with Zigbee2MQTT availability
disabled, sleepy end devices, or a bridge that is up while the device is not) renders a stale value
as if it were live. The user cannot see it from the card.

### F2 — The store has no notion of snapshot age · **High**

`store/modules/hub-state/state.ts` holds `initialized` and `loadStatus` (`idle | loading | ready |
failed`), but no `lastSyncedAt`. `useHubConnection` publishes the HTTP snapshot once, then WebSocket
messages merge silently (`hubSnapshotReplay.ts`). When the socket is down, the last values stay on
screen with no timestamp and no dimming; the only signal is the small header pill.

Consequence: the header says “Reconnecting”, but the page still looks authoritative. Even a
correct connection pill does not tell the user *which* data is 23 minutes old.

### F3 — Alerts are ephemeral and have no inbox · **High**

`store/modules/alerts/index.ts` uses `TIMEOUT = 3000` for every severity and stores only in Vuex.
`ws/index.ts` maps `operationFailed` straight to `alerts/showError`. There is no persisted read
state, no unread count, no grouping, and no way to revisit an error after the toast disappears.

The only durable attention surface today is Overview → “Needs a look”, which is limited to
offline / disabled / low battery / weak link (`domain/network.ts:40`) and is not visible from other
pages.

### F4 — Console is a thin page in the wrong place · **Medium–High**

Observed in `components/hub/console/ConsoleViewer.vue` and `store/modules/console/index.ts`:

- Nav item at `components/shell/navigation.ts:28` competes with the things people use daily, while
  the same two controls (remote logging toggle, level) also live in Settings → Logger.
- No search, source filter, “problems only”, pause-follow, copy, or download. Only level and clear.
- The level dropdown writes **hub** settings while sitting in a **browser** view; the toggle is
  “send logs to this browser”.
- Buffer is 1,000 messages, memory-only, 24 h expiry, and cleanup starts only when the page is
  visited (`services/console-cleanup.service.ts`).
- Every message triggers a deep watch (`watch(messages, { deep: true })`) and a scroll to the
  bottom; rows are keyed by `message.timestamp`, which is not guaranteed unique.
- The backend already exposes `GET /api/listlogs` and `POST /api/logfile` (protected routes,
  `internal/hub.go:91-94`), returning file names and file bytes, but the frontend has no client
  for either. Archived logs are unreachable from the UI.

### F5 — Settings cannot clean anything up · **High for operators**

`StorageSettings.vue` has a single action: Refresh. Retention is configurable, but there is no
purge, no “delete history before…”, no log pruning, no cache reset, no export/backup. Sections are
fixed, one at a time, with no search and no summary of current values. Scope is implied by the
group (“This browser” / “Hub”) but not repeated on individual rows, so it is easy to misread what a
toggle affects.

### F6 — Attention signals are computed but not routed · **Medium**

`needsAttention()` produces a good ranked list, used only by Overview. Devices going quiet are not
in it (see F1). There is no global badge, no sidebar count, and no shared selector that other
surfaces could reuse.

### F7 — No new-session state · **Medium**

Reloading the app resets alerts and nothing records “what changed since last visit”. The user has
to reconstruct the story from the log buffer or memory. A digest computed from unread inbox items
and current attention would solve this without replaying a queue of toasts.

### F8 — Card and layout polish · **Low–Medium**

- The tile root carries `:title="entity?.description"` (`EntityTile.vue:85`), which produces the
  native tooltip seen overlapping the Presence card in the screenshot. Native `title` is slow,
  unstyled and inaccessible on touch; use an explicit info affordance instead.
- Grid auto-flow leaves large holes when a row has fewer tiles (Garden/Kitchen in the screenshot);
  tiles also vary in height. A consistent span/row policy or a denser packing mode would help.
- Room headers carry no health summary; there is no way to scan “which room has a problem”.
- Presence and observation cards claim a full row even when their content is one line.

### F9 — Automation actions are minimal · **Low–Medium**

Rows offer Run, enable toggle and Delete. There is no export, duplicate, copy ID, “last run”, or
“why did it fail?”. Export helpers already exist for Home layouts (`services/home/export.ts`) and
CSV history, so the pattern is proven.

### F10 — Log viewer performance · **Medium**

Even at 1,000 rows the current watch + key + scroll pattern does unnecessary work per message:
deep traversal of the whole array, a full list re-render path, and a forced scroll that fights the
user when they scroll up to read. A global drawer makes this worse because rendering would stay
mounted. Fix with a ring buffer, batched commits, stable IDs, a non-deep watch and scroll pinning.

### F11 — Mobile is mostly “desktop, narrower” · **High**

There are good foundations: the sidebar becomes a drawer below 960 px, log entries stack below
768 px, and the settings rail becomes a horizontal scroller below 860 px. But the patterns the
proposal adds make mobile worse if copied literally:

- A top-level drawer with a 5-row log table at 390 px leaves ~60 px for the message; the header
  would hold four icon buttons plus title.
- A row kebab menu opened near the right edge is easy to miss and can clip.
- Settings cards with three-column health chips, sliders with values in tooltips, and destructive
  buttons sitting beside explanatory text are all hostile on touch.
- Hover-only affordances (tile layout tools, card hover borders) have no mobile equivalent.

Mobile needs different shapes for the same information — not smaller desktop frames.

### F12 — Settings information must not be collapsed · **High (design constraint)**

The first dashboard draft showed eight cards but omitted **MCP server**, implied detail pages were
just summaries, and lost the visible inventory (six registered section ids). Any index must be
*additive*: all existing sections stay reachable with every control, and Maintenance is a new group,
not a replacement. Mocks 4B (index), 4C (MCP depth) and 4D (cleanup) now demonstrate this.

### F13 — Overview has no hierarchy · **High**

The page stacks four stat cards, a connection group, three attention groups, an activity feed and a
network table — nine competing blocks, several repeating the same counts. Everything is equally
prominent, so nothing is. The four stat cards alone consume the first screen without answering “what
should I do first?”.

### F14 — Panel mode was not covered by this review · **High**

Panel mode is a different device: a wall tablet read at ~2 m with a thumb. Its current status page
uses four dials, the room page mixes a climate dial with tile buttons, and there is no night
simplification or explicit “paused” treatment for an offline hub. Responsive rules from the web app do
not transfer; it needs its own density and its own mock.

---

## 3. Recommendations

Effort is rough: **S** ≤ 1 day, **M** 1–3 days, **L** 3+ days, for one experienced contributor
including tests.

### R1 — Separate “availability” from “reporting freshness” · **P0 · M**

Keep `deviceStatus()` exactly as it is (automation routing and command gating depend on it; do not
widen spec 068). Add an orthogonal, display-only concept:

```ts
// domain/devices/freshness.ts (new)
export type ReportFreshness = 'fresh' | 'quiet' | 'stale' | 'unknown';
export function reportFreshness(
  device: Device,
  now: number,
  policy: { quietAfterMs: number; staleAfterMs: number }
): ReportFreshness;
```

- `quiet` = no report for more than the device’s expected interval × 3 (minimum 15 min).
- `stale` = beyond × 12 (minimum 6 h), or when the socket is down beyond the snapshot grace period.
- Default derived from `appConfig.hub.devices.defaults`, with per-device overrides already supported
  by Device defaults; battery/change-only sensors need a longer window.
- One shared 60-second ticker (`useNow`) drives relative ages; never one timer per card.
- Surface it in `EntityTile`, `HomeCardItem`, `DeviceCard`, `DeviceList`, Overview attention and
  panel mode. Copy: `Updated 4 h ago · not reporting`, `Quiet · 4 h`, `Offline · last seen 2 h`.

Important nuance: **silence is not offline.** A contact sensor may legitimately not report for a
day. Treat `quiet` as an amber hint with an explanation, not as a failure, and never let it change
automation semantics. The precise thresholds are an open question (§6).

### R2 — Make data age a first-class global state · **P0 · S**

- Add `lastSyncedAt` to hub state; update it when a snapshot is published and (throttled) when
  live reports arrive.
- Show a thin banner in `StatusBanners` when `connection !== 'connected'` or the snapshot is older
  than the grace period: *“Live updates paused. Showing the snapshot from 16:32 (23 min old).”*
  with a Retry action.
- Dim or mark tiles while the snapshot is stale; keep commands disabled (already handled by
  `canControl`).
- Add “Snapshot 16:32” to the Home/Overview/Devices subtitle when not live.

This is objective — it needs no device thresholds — and alone fixes the reported WebSocket
blindness.

### R3 — Add a notification inbox and keep toasts for action results · **P0/P1 · L**

New `notifications` module (separate from `alerts`, which stays the toast bus):

```ts
type Notification = {
  id: string; at: number; read: boolean;
  severity: 'info' | 'success' | 'warn' | 'error';
  category: 'connection' | 'device' | 'automation' | 'integration' | 'storage';
  title: string; body?: string;
  dedupeKey?: string;                 // “attic-light-quiet”, “ws-lost”
  action?: { label: string; to: RouteLocationRaw } | { label: string; event: string };
};
```

- Producers: connection lost/restored, `operationFailed`, device entering `quiet`/`stale`
  (debounced, one per transition), automation failures (derive from log events until the hub emits
  a structured event), storage above threshold, integration refresh errors.
- UI: bell + unread badge in `AppHeader`, slide-over grouped by day, filters (All / Problems /
  System), “Mark all read”, per-item action links. Mobile: full-height sheet.
- Sidebar “Needs attention” entry with count, reusing `needsAttention` + freshness.
- Persist in `localStorage`, capped at ~200 entries / 7 days, scoped to hub + user.
- Keep toasts for direct results (saved, command sent/failed) but give error toasts an **Open
  logs** action and also write them to the inbox.
- Raise actionable toast lifetime to ~8 s; keep 2–3 s for plain success.

This is the single change that answers “I saw a toast and then it was gone”.

### R4 — New-session digest · **P1 · S (after R3)**

- Store `lastVisitedAt` per browser.
- On the first Overview render of a session, show one card: *“While you were away — 3 things”*
  with at most three rows (unread errors first, then quiet/offline devices) and **Review all** /
  **Dismiss**.
- Dismiss persists until something new arrives. No toasts are replayed.
- Badge stays visible in the header so the digest can be dismissed without losing the signal.

### R5 — Reposition logs as a utility · **P1 · M**

Information architecture (mock 2C + 2B):

- Canonical page: **Settings → Logs & diagnostics** (tabs: Live tail, Archived files, Retention).
- Global access: terminal icon in `AppHeader` with an error-count badge opens a bottom drawer
  (mock 2B) over any page. Drawer holds search, level chips, problems-only, pause-follow, download,
  clear and “All logs”.
- Remove Console from the main nav; keep the `/consoleviewer` route as a redirect or full-page
  “expand” target so existing deep links still work.
- Merge the duplicated controls: the Logger settings section becomes part of Logs & diagnostics.

Viewer changes:

- Ring buffer (2,000 messages), batched commits on a microtask/`requestAnimationFrame`, stable IDs
  (`timestamp + sequence`), `shallowRef` + non-deep watch.
- Virtualise or cap rendered rows (e.g. 300) with “load older”.
- Auto-scroll only while pinned within ~40 px of the bottom; otherwise show “N new messages”.
- Filters: level chips with counts, source (device/hub/automation), free-text, problems-only.
- Row actions: click to copy, expand for structured details, timestamp in the user’s locale.
- Downloads: current buffer as `.log`/`.json`; archived file via `POST /api/logfile`
  (`{ file, action: "load" }`, expects `Content-Type: application/json`) with a blob download; list
  via `GET /api/listlogs`. Both routes already exist and are protected.
- Optional: keep the last N messages in `sessionStorage` so an accidental reload does not lose the
  trail that led to an error.

### R6 — Settings: additive index, full detail pages, search, scope · **P1 · M**

Two levels, no information removed (mocks 4B index, 4C detail, 4D cleanup):

**Index (`/settings`)**

- Hub health strip: connection, version, data size, free disk, uptime, log errors, MCP state.
- Three groups, every existing section represented:
  - *This browser*: Appearance.
  - *This hub*: Device defaults, Storage & retention, Logs & diagnostics, **MCP server**,
    **Assistant**.
  - *Maintenance* (new): Data & cleanup, Backup & export.
- Each card shows a live summary and warning chips (e.g. `78 % used`, `2 errors`, `Running ·
  2 clients`). Cards are links to the detail page, not inline replacements.

**Detail (existing rail + section, deep-linkable)**

- Keep `?section=` and the current rail; add search, per-item summaries, and a “Maintenance” group.
- Every control that exists today stays: Appearance (mode, palette, accent, density, corner
  rounding, text size, texture, reset), Device defaults (reporting interval, sensor selection,
  overrides), Storage (cleanup interval, keep history, usage, disk), Logs (remote stream, level,
  buffer, archives, retention), MCP (status, name, version, clients, Start/Stop/Restart), Assistant
  (service URL, save/validation).
- MCP and Assistant were missing from the first draft; the corrected mock includes both, and 4C
  proves the depth is preserved.
- Repeat scope per row (`This browser` / `This hub`), not only in group headings.
- Add “changed from default” indicators, per-row reset where it makes sense, and a breadcrumb back
  to the index.

**Inventory guard:** `SETTINGS_SECTIONS` remains the source of truth. A test should assert that the
index renders a card for every non-alias section id, so a future section cannot silently disappear.

### R7 — Data & cleanup · **P1 · M–L (backend work required)**

New Settings section (mock 4D), grouped by blast radius:

- **Storage:** usage bars; purge history before a chosen date; purge spatial/position history;
  prune logs by retention now; show freed space after each action.
- **Clear:** browser cache/layout cache, reset appearance, clear automation run history.
- **Danger zone:** delete all history, remove all automations — typed confirmation naming the
  action and impact counts.
- **Export & backup:** Home layout (helper exists), all automations as JSON, full config backup.
- Every destructive action writes an audit line to the log and shows a result toast plus an inbox
  entry.

Backend gaps to confirm before implementation: purge endpoints for metrics/spatial/log files and a
full-backup payload. The log file list/read endpoints exist; history deletion likely needs new
routes. Do not fake these in the client.

### R8 — Automation row actions and export · **P2 · S**

- Kebab per row: Run now, Edit, Duplicate, Export JSON, Copy ID, View last run / failure, Delete.
- Header: Export all, Import (validate with the same bounded pattern as Home exports).
- Failure state: replace the generic toggle-only row with a visible `Failed 3×` chip linking to the
  relevant log lines and last run time.
- Export shape: a small versioned envelope (`{ version, exportedAt, automation(s) }`) so import can
  reject unknown formats; reference devices by ID, never by index.

### R9 — Card and layout polish · **P2 · S**

- Remove the root `title` attribute from `EntityTile`; add an explicit info button/popover for the
  description (keyboard and touch accessible).
- Add a room header health summary (“1 quiet · 1 offline”) and a per-area “needs a look” count.
- Define tile spans per kind so sparse rooms do not leave holes; consider a compact density that
  drops the device line (keep it in the tooltip/popover).
- Make stale/offline states distinguishable without colour alone (icon + text, as in mock 1A/1C).
- Presence/observation cards: collapse to a summary row unless live activity is present.

### R10 — Accessibility and motion · **P2 · S**

- Badge changes announced via `aria-live="polite"`; bell has an accessible count label.
- Drawer and inbox use the existing dialog/focus-trap pattern, `Esc` to close, focus restored.
- Respect `prefers-reduced-motion` for the alert pulse and drawer transitions.
- Ensure the log stream does not steal focus or announce every line.

### R11 — Mobile-first shapes for every surface · **P0 for this work · M**

Mocks M1–M8. Treat 390 × 812 as the reference viewport and keep the desktop layout as the wide
variant.

**Shared rules**

- 44 px minimum targets; no hover-only affordance; text actions have a visible pressed state.
- Header keeps at most two utility icons: **notifications** and **logs**. Theme and user move into
  the nav drawer. The title truncates before the icons shrink.
- Sheets replace popovers and row menus: grabber, focus trap, `env(safe-area-inset-bottom)`
  padding, swipe/`Esc` to close.
- Destructive actions open a sheet with impact counts and a typed confirmation; never a 34 px
  button next to explanatory text.
- Respect safe areas and dynamic viewport height (`100dvh`); never rely on `100vh`.

**Per area**

| Area | Mobile shape |
| --- | --- |
| Home (M1) | Banner + digest, then one area at a time; single-column tiles with the age line; room chips stay right-aligned; no grid holes |
| Logs (M2) | Full-height sheet/full page; horizontally scrolling level chips; two-line rows (meta, then message); footer shows follow state and buffer |
| Logs settings (M3) | Equal-width segmented tabs; archive rows stack with a 44 px download icon; “Download all” as a full-width button |
| Alerts (M4) | Full-screen inbox; actions wrap under the body; “Mark all read” as a text button in the header |
| Settings index (M5) | Grouped list cards, one per section, all six present; health chips wrap to two rows |
| Settings detail (M6) | Scrollable section picker; label + help above full-width control; sliders show the value in the label |
| Cleanup (M7) | Stacked destructive rows; confirmation sheet with impact text and typed `DELETE` |
| Automations (M8) | Bottom-sheet actions with 48 px rows, name in the sheet header, delete separated |

**Verification**

- Browser specs for every changed page at 390 × 844 and 360 × 740, plus landscape 844 × 390 for
  sheets and the log viewer.
- No horizontal page scroll; only designated chip/tab rails scroll.
- Tap targets measured at ≥ 44 × 44 CSS px.
- Keyboard and screen-reader checks on desktop; touch checks on mobile.
- With a 200 % text scale, no action is clipped or overlapped.

Optional, not required for this pass: a bottom tab bar (Home / Devices / Automations / More) would
suit a wall-panel install, but it changes navigation for every screen and should be its own spec.

### R12 — Overview: one verdict, three actions, one line of facts · **P1 · M**

Mocks `07-overview-calm-desktop.png` (recommended) and `08-overview-domains-desktop.png`
(alternative), plus `09-overview-calm-mobile.png`.

- Keep the four numbers, remove the four cards: one inline health line with links.
- Replace the attention groups with “Start here”: at most three rows, ranked by consequence, each
  with one action. A “Show N more” link reveals the rest.
- Delete the network table from Overview; the same numbers already exist on Devices and the panel status page.
- Move the new-session digest here (status surface), not to Home (control surface).
- Provide a real all-clear state: one sentence and a timestamp, not four empty cards.
- Keep “What changed” as three rows; the full feed is a link.

Success: the first viewport answers “is anything wrong, and what first?” without scrolling.

### R13 — Panel mode: day, night/offline, status · **P1 · M**

Mocks `12-panel-day.png`, `13-panel-night-offline.png`, `14-panel-status.png`.

- **Day:** clock and weather are glanceable, one climate instrument, 6–9 large controls with state
  and value. Minimum target 64 px; adjacent gap ≥ 8 px.
- **Night:** drop to a large clock and four configurable essentials. Warm, dim palette; no glows,
  no network detail. Decide the trigger (schedule vs ambient light) before implementation.
- **Offline:** one banner with snapshot age, every control shows “Paused” and keeps last-known state;
  never leave a control looking live while the hub is unreachable.
- **Status:** four instrument readings plus a short attention list. No tables, no scrolling for the
  common case.
- Keep the existing dock (Rooms / Status / Exit) and the room pager; they already match the device.

### R14 — Home stays the control surface · **P2 · S**

Mock `01-home-desktop.png` / `02-home-mobile.png`.

- Add one status strip under the header: count of issues, offline/quiet split, snapshot time, link
  to Overview. Nothing else moves.
- Room headings show a health chip only when something is wrong.
- Move the new-session digest to Overview; Home should get to the rooms immediately.
- Do not add network tables, attention groups or extra stat cards here.

---

## 4. What not to do

- **Do not treat “quiet” as offline** in `deviceStatus`, automations or `canControl`. Availability
  and silence are different facts; conflating them risks the automation guard and creates alert
  fatigue. Keep freshness display-only until a spec says otherwise.
- **Do not keep both** a top-level Console page and a separate log section with duplicated toggles.
- **Do not replay toasts** on a new session; show a digest instead.
- **Do not add per-card timers** or polling; one ticker plus a sync timestamp is enough.
- **Do not add client-side purge behaviour** before the backend endpoints exist; destructive UI
  must reflect real server capability.
- **Do not drop settings sections** when introducing the index. The index is additive; MCP and
  Assistant are not optional, and detail pages keep every control.
- **Do not shrink desktop patterns** onto mobile — no 5-column log tables, edge menus or
  side-by-side destructive buttons. Use sheets, stacked rows and full-width controls.
- Do not widen the automation feedback-loop rules to make “notifications on device update” work;
  that needs its own spec (spec 068 constraint).

---

## 5. Suggested sequence

| Step | Work | Outcome |
| --- | --- | --- |
| 1 | R2 (snapshot age + paused banner) | objective “data is old” signal, small diff |
| 2 | R1 (freshness model + 1A card treatment) | per-card honesty without changing status |
| 3 | R3 minimal (connection + operation failures + inbox shell + error-toast action) | nothing actionable disappears |
| 4 | R4 (digest) | new-session flag |
| 5 | R5 (logs IA + viewer + archive download) | one clear home, real tooling |
| 6 | R6 (settings index + details, no information lost) + R7 (cleanup) | discoverability and maintenance |
| 7 | R12 + R14 (Overview hierarchy, Home strip) | trust summary without noise |
| 8 | R13 (Panel mode day/night/offline) | wall surface that tells the truth |
| 9 | R8 + R9 + R10 | polish, export, accessibility |

R11 (mobile) is implemented **with** each step, not at the end: every frame that ships gets its
390 px browser spec and touch-target check in the same PR. A separate mobile-only retrofit would
mean building the wrong shapes twice.

Each step is a behaviour change: per the SDD workflow it needs a spec → plan → tasks before
implementation, with tests colocated beside the owning source file (AGENTS test ownership). R1 and
R2 are small enough to share one spec (“data trust”); R3/R4 one (“notifications”); R5–R7 are
separate specs because R7 crosses into backend routes; R12–R14 form one “surface hierarchy” spec.

---

## 6. Open questions

1. **Freshness thresholds.** Should “quiet” be derived from the configured reporting interval
   (default 60 s), or a fixed window (e.g. 6 h) for every device? Battery and change-only sensors
   need a longer window; do we add an explicit “expected reporting interval” to device settings?
2. **Should quiet devices appear in Overview “Needs a look”** and count toward the badge, or stay a
   softer Home-only hint?
3. **Notification scope.** Browser-local only (recommended first) or hub-side so every client sees
   the same inbox? The latter needs persistence and read-state sync.
4. **Console nav.** Remove “Console” from the sidebar in favour of Settings + header drawer, or keep
   the nav item and have Settings link to it? The mock recommends removal with a drawer.
5. **Cleanup permissions.** Should any authenticated user be able to purge history/automations, or
   only an admin role? Current auth model needs checking before the danger zone ships.
6. **Automation export contract.** Reuse the Home export envelope and versioning, or define a
   dedicated import/export contract? Import must validate device references.
7. **Log retention defaults.** The UI shows 30 days in the mock; the hub currently has no visible
   retention setting for files. Is pruning a new setting or derived from the history window?
8. **Settings navigation.** Confirm the additive model: index page plus the existing rail on detail
   pages. The alternative — index-only with expanding cards — would sacrifice depth, so it is not
   recommended.
9. **Mobile navigation.** Keep the hamburger + two header icons, or introduce bottom tabs? The mocks
   assume the former to limit scope; bottom tabs suit a wall panel but need their own spec.

---

## 7. Mock index

Open `mocks/ui-improvements.html` and use the left rail; each frame is labelled with its trade-off.

| Mock | File | Idea | Verdict |
| --- | --- | --- | --- |
| 1A | `shots/01-home-a-age-line.png` | age line on every tile | **recommended baseline** |
| 1B | `shots/02-home-b-dim-chip.png` | corner age chip, dim, `~` values | compact alternative |
| 1C | `shots/03-home-c-outage.png` | status verbs, room rollups, paused banner | **use with 1A for outages** |
| 2A | `shots/04-console-a-page.png` | improved full-page log viewer | deep-work view |
| 2B | `shots/05-console-b-drawer.png` | global drawer + header shortcut | **recommended quick access** |
| 2C | `shots/06-console-c-settings.png` | Settings → Logs & diagnostics | **recommended home** |
| 3A | `shots/07-alerts-a-inbox.png` | bell + slide-over inbox | **recommended** |
| 3B | `shots/08-alerts-b-digest.png` | “While you were away” digest | **recommended for sessions** |
| 4A | `shots/09-settings-a-rail.png` | refined rail with summaries/search | least disruption |
| 4B | `shots/10-settings-b-index.png` | settings index, all six sections + maintenance | **recommended index** |
| 4C | `shots/11-settings-c-mcp-detail.png` | detail page keeps full depth (MCP shown) | **proof nothing is lost** |
| 4D | `shots/12-settings-d-cleanup.png` | Data & cleanup | **fills the gap** |
| 5A | `shots/13-automation-a-menu.png` | row kebab + export all | **recommended** |

**Mobile**

| Mock | File | Idea |
| --- | --- | --- |
| M1 | `shots/14-mobile-home.png` | Home: banner + digest, single-column tiles, age line |
| M2 | `shots/15-mobile-logs.png` | Logs: full-screen sheet, scrolling chips, two-line rows |
| M3 | `shots/16-mobile-log-files.png` | Logs settings: segmented tabs, archive download |
| M4 | `shots/17-mobile-notifications.png` | Notifications: full-screen inbox, wrapped actions |
| M5 | `shots/18-mobile-settings-index.png` | Settings index: all sections, grouped list |
| M6 | `shots/19-mobile-settings-detail.png` | Settings detail: section picker, full-width controls |
| M7 | `shots/20-mobile-cleanup-confirm.png` | Cleanup + typed confirmation sheet |
| M8 | `shots/21-mobile-automation-sheet.png` | Automation actions as a bottom sheet |

Home Assistant is used only as a pattern reference: logs and diagnostics under Settings, a
persistent notifications panel, and a Repairs-style attention list with per-issue actions. The
proposed structure is not a copy — it keeps NodeHerder’s rail, tokens and existing route/query
conventions. Mobile follows the same system: no new visual language, only different shapes.

---

## 8. Wider layout pass: system first, not a big-bang redesign

**Verdict: do a system pass; do not redesign everything at once.**

The changes above introduce shared semantics — freshness, status, attention, sheets, scope — that
other pages will visibly contradict if they are left alone. But “touch everything” is the wrong
shape for this codebase: it turns one reviewable change set into several specs, breaks muscle
memory, and multiplies the testing surface without solving a user problem that those pages have
reported.

Treat it as rings. Inner rings cascade outward cheaply; outer rings need their own evidence.

### Method

Reviewed with the local `ui-ux-pro-max` knowledge base. Targeted checks that changed the design:

- **Colour never alone** → every status carries an icon and text, including panel controls.
- **Touch target and spacing** → 44 px primary targets, 8 px gaps, no hover-only actions; panel
  controls are larger by design.
- **Table handling** → desktop keeps the dense table; mobile uses card rows, not horizontal scroll.
- **Destructive confirmation** → name the action and the impact; typed phrase for irreversible data
  loss, undo for routine removal.
- **Focus and live regions** → visible focus rings, sheets trap and restore focus, badge changes
  announce a phrase (“3 items need attention”), not a bare number.
- **Contrast** → muted text kept above 4.5:1 on the dark surfaces; no reliance on elevation alone.

Project tokens take precedence; no new palette or font was introduced.

### Rings (updated)

| Ring | Scope | Why / when | Risk |
| --- | --- | --- | --- |
| **0 — Primitives** | `<StatusChip>` + freshness language, page header, card/row, setting row, empty/loading/error, sheet/dialog, header actions, form field | These are copied by every page. Fix once, get consistency everywhere without page rewrites. | Low; mostly additive |
| **1 — Surfaces in this review** | Home cards, logs, notifications, settings, automation actions, mobile shapes | Already mocked and decided; mobile included. | Medium; shipped incrementally |
| **2 — Daily surfaces** | Home strip, Devices, Device detail, Overview redesign, Integrations board, Panel mode | Same status/attention problems, one click from Home. Overview is now a dedicated redesign, and Panel mode is a different device that cannot wait for a generic responsive pass. | Medium |
| **3 — Editors and deep flows** | Automation editor/creator, integration setup/provider forms, Assistant chat | High value but large and specialised. Do one at a time, with their own mocks and specs. | High |
| **4 — Low-traffic / static** | Login, permission dialogs, about/help | Only touch if Ring 0 changes force it. No mock needed now. | Low value |

### What is actually worth changing outside this review

1. **Header density.** Collapse theme/palette/user into one account menu; keep connection,
   notifications, logs. Matches the mobile rule and gives the title room.
2. **One status vocabulary.** Availability, integration freshness and observation stability must
   render through one grammar: icon + text + tone, `fresh / quiet / stale / offline / disabled /
   unknown`. Highest leverage after freshness itself.
3. **Page header and states.** One header pattern and four list states (loading, empty, error,
   no-results). Some pages currently hand-roll headings or show bare spinners.
4. **Home ↔ Overview boundary.** Home controls, Overview judges. No duplicated attention lists or
   network tables on either; the digest moves to Overview.
5. **Devices, device detail, integrations.** They already compute the truth; align the presentation
   and make freshness actionable. No new data model.
6. **Panel mode.** Its own density, night rule, offline honesty and larger targets. Do not derive it
   from desktop cards.
7. **Native browser tooltips.** Remove remaining `title`-attribute explanations in favour of the
   Ring 0 popover/description pattern.

### What not to change

- **Navigation IA again.** Removing Console is enough. Merging Overview into Home is a separate
  experiment; it changes deep links and needs its own evidence.
- **Theme token names or palettes.** They are the only stable layer; page work consumes them.
- **Component library.** Do not swap PrimeVue mid-stream. Consolidate `Ui*` primitives and stop
  adding new PrimeVue usages where an existing primitive fits.
- **Route structure.** Keep `?section=`, aliases and deep links.
- **Panel mode’s wall-device identity.** Do not shrink desktop cards into it or make it look like the
  mobile app; it is a 2 m glance surface.

### Ring 2 mocks delivered

Gallery: [`docs/design/mocks/ring2.html`](mocks/ring2.html) · screenshots in
[`docs/design/mocks/shots-ring2/`](mocks/shots-ring2/).

| Mock | File | Idea |
| --- | --- | --- |
| P0 | `shots-ring2/p0-pattern-sheet.png` | status/freshness matrix, actions, header, row anatomy, four states, typed confirm, tokens, interaction contract |
| H-D | `shots-ring2/01-home-desktop.png` | Home status strip, rooms, health chips; no new card design |
| H-M | `shots-ring2/02-home-mobile.png` | same strip; digest moved to Overview |
| D-D | `shots-ring2/03-devices-desktop.png` | freshness column, attention banner, quiet/stale filter |
| D-M | `shots-ring2/04-devices-mobile.png` | grouped rows, touch-sized overflow |
| V-D | `shots-ring2/05-device-detail-desktop.png` | trust header, “last known” values, diagnostics + maintenance |
| V-M | `shots-ring2/06-device-detail-mobile.png` | warning first, stacked controls, sticky Settings/History |
| O-A | `shots-ring2/07-overview-calm-desktop.png` | one verdict, three actions, one health line, three events (recommended) |
| O-B | `shots-ring2/08-overview-domains-desktop.png` | domain board alternative; four subsystem rows + change feed |
| O-M | `shots-ring2/09-overview-calm-mobile.png` | calm version on mobile; digest lives here |
| I-D | `shots-ring2/10-integrations-desktop.png` | summary, attention filter, explicit stale/error reasons |
| I-M | `shots-ring2/11-integrations-mobile.png` | grouped rows, inline retry, no dead refresh |
| P1 | `shots-ring2/12-panel-day.png` | wall panel: clock, climate dial, 9 large controls |
| P2 | `shots-ring2/13-panel-night-offline.png` | night simplification + offline “Paused” treatment |
| P3 | `shots-ring2/14-panel-status.png` | four instrument readings + short attention list |

The pattern sheet is the contract: chips, header, rows, states, confirmation and interaction are
shared with Ring 1. Ring 2 excludes the automation editor, Assistant and Login.

### Decisions to confirm

1. Ring 2 scope: Home, Devices, Device detail, Overview, Integrations, Panel mode.
2. Overview direction: **A calm** (recommended) or **B domain board** — pick one, do not ship both.
3. Home ↔ Overview boundary: digest moves to Overview; Home keeps one status strip.
4. Panel mode: night trigger (schedule vs ambient light), the four configurable essentials, and
   whether “Paused” controls stay tappable for a retry.
5. Devices: add a `quiet` filter **and** a freshness column.
6. Integrations: per-row refresh + one “Refresh all”; no refresh-on-open.
7. Device actions: one overflow menu on both platforms, opening the Ring 1 sheet.
8. Extract Ring 0 primitives during Ring 1, not as a separate project.

### Recommended order

1. Finish Ring 1 spec → implementation (the current kit).
2. Build Ring 0 while implementing Ring 1; extract the chip/header/row patterns from real usage,
   not from a separate design project.
3. Spec Ring 2 after Ring 1 is in review, reusing the shipped primitives; the mocks are ready now.
4. Ring 3 one flow at a time, starting with Panel mode or the automation editor — whichever is
   causing more real support pain.

Doing every mock at once would produce a large design document that drifts before implementation.
The rings keep each decision reviewable and let the system settle between passes.
