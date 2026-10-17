# Plan: Recent activity device search

Spec: [spec.md](spec.md) | Status: Complete | Updated: 2026-10-17
Constitutions: parent/frontend v1.0.0 proposed; backend N/A.

## Context and design

ActivityFeed.vue owns a local search ref and derives filtered entries from the existing
useActivityFeed bounded store. Add a framework-free filterActivityEntries helper to
 domain/activity.ts, tested in its existing companion. Match current friendly names
with device ID fallback for removed devices. Filter before slicing/formatting, preserving
order and a maximum of 60 rendered rows. At most 200 retained events: no debounce,
network requests, extra subscriptions or retained indexes needed.

Use an inline toolbar below the card header with the existing search icon/input style,
explicit clear-search button, filtered count and a distinct no-match message. Leave
recording/Clear controls intact; Clear clears all activity even while filtered, with
an accessible name explaining that scope. Search stays available while empty/off.
No wire contracts/persistence changes. No ADR needed for local presentation filtering.

## Constitution Check

| Rule | Evidence/design | Before implementation | Final |
| --- | --- | --- | --- |
| NH-01 | AC-01–03 and tasks first | PASS | PASS |
| NH-02 | View-only; device/automation processing unchanged | PASS | PASS |
| NH-03 | No new boundaries or external input execution | PASS | PASS |
| NH-04 | Pure domain filter; component presentation | PASS | PASS |
| NH-05 | Companion unit plus browser regressions | PASS | PASS |
| NH-06 | Existing bound; no storage/workers | PASS | PASS |
| FE-01 | Existing feed state remains authoritative | PASS | PASS |
| FE-02 | No contracts changed | PASS | PASS |
| FE-03 | Local reactive view, no requests | PASS | PASS |
| FE-04 | Search label, reset, focus, theme/390–1280px | PASS | PASS |
| FE-05 | Search not persisted; no raw HTML | PASS | PASS |

## Traceability and verification

AC-01/02: activity.test.ts matching/order/limit/immutability; ActivityFeed.spec.ts mock
WS updates, live matching, reset, links and clear-all while filtered. AC-03 browser
empty/off/keyboard and mobile/desktop screenshots. Run Jest/lint/build and affected
browser suite including overview regression. No household-device checks. Results (2026-10-17):

- `npm test -- --runInBand --silent`: PASS, 32 suites / 247 tests. New companion
  cases verify literal partial matching, case/whitespace, removed-device fallback,
  empty results, order, immutability and matches beyond the visible window.
- `npm run build`: PASS. Existing DeviceCard naming and large-chunk warnings remain.
- `npm run lint`: PASS, seven existing script warnings. Initial concurrent lint run
  raced Playwright deleting test-results and returned ENOENT; reran after browser
  completion, successfully. No lint/source issue was suppressed.
- `npm run test:e2e -- e2e/ActivityFeed.spec.ts e2e/overview.spec.ts`: 4/4 PASS across
  desktop and mobile, using controlled incoming mock frames (no device commands).
- Reviewed screenshots at 1280px/390px: search, result count and rows fit; visible
  keyboard focus and no horizontal overflow. Escape/reset retain focus; off/waiting/
  no-match states remain distinct. New matching reports appear without losing hidden
  device events, and clear-all removes every retained event while filtered.
- `git diff --check`: PASS. Final review confirms filtering before formatting/cap,
  constant-time device-map lookup, bounded 200-event scan only for non-empty queries,
  no new watcher/subscription/timer or persistence. Recording logic unchanged.

Updated docs/setup.md. Full browser/backend suites not rerun for this isolated view
change; affected browser checks plus all frontend units/lint/build passed.

## Completion

No backend check required: no backend source changed. Work remains uncommitted.
