# Spec: Compact mobile history and navigation

ID: 030 | Status: Implemented; automated checks passed | Owner: NodeHerder contributors | Updated: 2026-10-23
Constitution: proposed 1.0.0 | Components: frontend

## Problem, scope, exclusions

The supplied 384×854 phone screenshot shows oversized entity/period pills and a wrapping export
row pushing the chart below most of the screen. `HistoryToolbar.vue` currently renders two
horizontal chip strips plus wrapping mode/actions. Improve phone density while preserving the
history/statistics and data semantics delivered by merged spec 029. Desktop keeps labelled chips.
The user also requests removing the redundant phone bottom bar in favour of the existing hamburger
drawer. Restore every destination in that drawer and release the bar’s reserved content space.
This supersedes spec 029 AC-10’s bottom-bar presentation; route semantics remain unchanged.
No transport, device command, storage, query, aggregation or dependency change.

## Scenarios and acceptance

### US-01 — Reach the chart quickly on a phone (P1)

- AC-01: Below 720px, entity and period use two labelled native pickers on one row. All entities
  and eight presets remain available; selecting them updates history, including Custom bounds,
  and restores the selection after reload using existing preferences.
- AC-02: On phones, metric, period and a named History options button share a single 44px row.
  Options opens the existing dialog/sheet with a labelled History/Statistics picker (numeric,
  non-energy only), refresh, export, native/display unit disclosure and full selected bounds/zone.
  Closing or changing mode returns focus to the opener; Escape closes the sheet. Actions retain
  44px targets and disabled/error explanations; current mode remains described on the opener.
- AC-03: With loaded numeric 24h history/export available, the phone toolbar is at most 48px tall,
  and the canvas begins within 12px below it. No dedicated date/help/button rows precede the plot.
  Full numeric and statistics canvases are at least 280px and use roughly 48% viewport height
  (capped at 420px); verify 320/384/390 widths, 568/854 heights, both themes and no overflow.
- AC-04: At 768/1280px labelled entity/period buttons remain visible and selected state is accurate.
  Resizing preserves the current choice; statistics, export and empty/partial/error behavior remain.

### US-02 — Use the hamburger drawer without a bottom strip (P1)

- AC-05: No bottom navigation is rendered. At 320/390/768px the header hamburger opens a drawer
  containing Home, Devices, Automations, Assistant and all existing secondary links/actions.
  Navigation closes the drawer; Escape closes it; route and includes-based `aria-current` remain.
  The hamburger and drawer controls retain at least 44px touch targets. At 1280px the existing
  sidebar remains visible.
- AC-06: Phone content reserves no bottom-bar height. With zero inset, bottom padding is at most
  32px; a synthetic 32px safe inset protects content and the drawer’s last action. No page overflow;
  manifest/viewport metadata and manifest-failure boot still work. Panel mode keeps its own dock.

### US-03 — Give metrics the main phone surface (P1)

- AC-07: Pan/zoom/reset remain available via a keyboard-operable Chart tools disclosure below the
  numeric plot. It starts collapsed, closes with Escape, and does not cover or precede the canvas.
  Gestures, keyboard shortcuts, range clamping and zero extra requests/commands remain unchanged.
- AC-08: Phone Statistics places its legend/chart ahead of the detailed summary/sample count/table;
  missing/partial data and every numeric statistic stay available. The compact entity-dialog chart
  and detailed desktop controls/layout retain their behavior.

## Requirements and success

FR-01: Compact phone controls; AC-01…03. FR-02: Preserve desktop and history behavior; AC-04.
FR-03: Remove redundant bottom navigation and restore drawer destinations; AC-05/06.
FR-04: Chart-first phone presentation with secondary controls on demand; AC-07/08.
NFR-01: Use existing tokens, scoped styles and dependencies. SC-01: Existing history browser flows
pass on both projects and the phone geometry/keyboard regression passes against isolated fixtures.

## Entities and external interfaces

Presentation only: existing selection events and props remain the source of truth. No network,
auth, MQTT, unit/timezone, retention, persistence schema or rollout ordering change. Frontend rollback
restores the old layout without migration. No household-device testing.

## Questions and assumptions

No blocking questions. Native pickers trade a tap for substantially less phone height; desktop
retains direct buttons. This follows the user's complaint about the supplied phone layout.
