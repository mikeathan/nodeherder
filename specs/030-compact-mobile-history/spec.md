# Spec: Compact mobile history toolbar

ID: 030 | Status: Implemented; automated checks passed | Owner: NodeHerder contributors | Updated: 2026-10-23
Constitution: proposed 1.0.0 | Components: frontend

## Problem, scope, exclusions

The supplied 384×854 phone screenshot shows oversized entity/period pills and a wrapping export
row pushing the chart below most of the screen. `HistoryToolbar.vue` currently renders two
horizontal chip strips plus wrapping mode/actions. Improve phone density while preserving the
history/statistics and data semantics delivered by merged spec 029. Desktop keeps labelled chips.
No transport, device command, storage, query, aggregation, dependency or global control-style change.

## Scenarios and acceptance

### US-01 — Reach the chart quickly on a phone (P1)

- AC-01: Below 720px, entity and period use two labelled native pickers on one row. All entities
  and eight presets remain available; selecting them updates history, including Custom bounds,
  and restores the selection after reload using existing preferences.
- AC-02: History/Statistics, refresh and CSV export fit on one subsequent row at 320/384/390px.
  Each control is at least 44×44px, named, keyboard operable and visibly focused. Export's native
  versus displayed-unit disclosure and disabled/error explanations remain available.
- AC-03: For a loaded numeric 24h history with export available, the toolbar is at most 144px
  tall at those phone widths, including the unit disclosure; no horizontal page overflow or clipped
  action. Date/zone remains readable. Verify both light and dark appearance with screenshots.
- AC-04: At 768/1280px labelled entity/period buttons remain visible and selected state is accurate.
  Resizing preserves the current choice; statistics, export and empty/partial/error behavior remain.

## Requirements and success

FR-01: Compact phone controls; AC-01…03. FR-02: Preserve desktop and history behavior; AC-04.
NFR-01: Use existing tokens, scoped styles and dependencies. SC-01: Existing history browser flows
pass on both projects and the phone geometry/keyboard regression passes against isolated fixtures.

## Entities and external interfaces

Presentation only: existing selection events and props remain the source of truth. No network,
auth, MQTT, unit/timezone, retention, persistence schema or rollout ordering change. Frontend rollback
restores the old layout without migration. No household-device testing.

## Questions and assumptions

No blocking questions. Native pickers trade a tap for substantially less phone height; desktop
retains direct buttons. This follows the user's complaint about the supplied phone layout.
