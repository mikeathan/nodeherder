# Plan: Compact connection status instead of stacked notices

Spec: [spec.md](spec.md) | Status: Implemented (unreviewed) | Author: agent draft | Updated: 2026-11-06
Constitutions: parent 1.0.0 (proposed); frontend 1.0.0 (proposed)

## Context and design

Facts: `StatusBanners.vue` owns a fixed bottom-right `nh-connection-notices` panel built
from `connectionNotices()` and a `dismissed` flag that is reset by a `watch(notices)`.
`ConnectionStatus.vue` already renders the header pill with `is-connected`,
`is-reconnecting`, `is-disconnected` state classes.

Design:

- `ConnectionStatus.vue` becomes the single surface: a button pill with the existing status
  dot/text, an attention badge when `connectionNotices()` is non-empty, an accessible label
  naming the status and issue count, and a `Connection` `UiDialog`.
- The dialog reuses `connectionNotices()` and the existing retry actions
  (`ws/connect`, `startHubConnection`) and shows one card per notice.
- `StatusBanners.vue` keeps only the permit-join banner; all connection/session/snapshot
  banners and the fixed panel are removed.
- No auto-opening and no toast: the badge/dot is the signal; the user opens the dialog.

Files:

| Path | Change |
| --- | --- |
| `src/components/shell/ConnectionStatus.vue` | pill button, attention badge, Connection dialog |
| `src/components/shell/StatusBanners.vue` | remove connection notices; keep permit join |
| `e2e/StatusBanners.spec.ts` | open the chip dialog instead of banners; content stability |

No new components, dependencies, stores, API or persistence.

Alternatives rejected:

- Toast-only: still interrupts on every transition and is harder to revisit.
- Auto-open the dialog: reintroduces the interruption.
- Shrink the existing banner: still duplicates the header pill and still needs Dismiss.
- Persisting dismiss in storage: unnecessary; the pill remains the state surface.

## Constitution Check

| Rule | Evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | Spec 058 with AC-01…AC-05 | PASS | PASS |
| NH-02 | No device commands; retries reuse existing connection actions | PASS | PASS |
| NH-03 | No new entry points, secrets or persistence | PASS | PASS |
| NH-04 | Reuses `connectionNotices`, `useHub`, `useSession`, store actions | PASS | PASS |
| NH-05 | StatusBanners e2e + full suites | PASS | PASS |
| NH-06 | No data mutation; notices derive from live state | PASS | PASS |
| FE-01 | Components/composable and store reuse only | PASS | PASS |
| FE-02 | No API/WS payload change | PASS | PASS |
| FE-03 | State remains observable in the pill; retry remains available | PASS | PASS |
| FE-04 | Button label, dialog name, Escape/Close, 44 px pill on phones | PASS | PASS |
| FE-05 | No secrets or household data added | PASS | PASS |

## Traceability and verification

| IDs | Paths | Procedure | Result |
| --- | --- | --- | --- |
| AC-01, AC-03 | `ConnectionStatus.vue`, `StatusBanners.vue` | startup and loss tests assert no `.nh-banner` and the pill badge/state | pass |
| AC-02 | `ConnectionStatus.vue` | open dialog, assert notice text and retry button, Escape closes | pass |
| AC-04 | `ConnectionStatus.vue` | loss/recovery test at `/groupdashboard` and `/panel` | pass |
| AC-05 | `ConnectionStatus.vue` | 2.75 rem pill at ≤760 px; dialog uses `UiDialog` bottom sheet | pass (CSS + viewport suites) |
| FR-04 | `StatusBanners.vue` | permit countdown test | pass |
| SC-01 | `e2e/StatusBanners.spec.ts` | 7 tests | pass |
| SC-02 | all | unit/build/lint/type-check/full e2e | recorded below |

## Verification (recorded)

- `npm run test:e2e -- --project=desktop e2e/StatusBanners.spec.ts` → **7 passed**.
- Full suites after the change: **142 passed/9 skipped desktop**, **151 passed mobile**.
- Unit tests: **748 passed**; build and `type-check:e2e` pass; lint 0 errors
  (28 pre-existing warnings).

## Debt, risks, exclusions

- No toast: a user who does not notice the badge will not see the notice text. The badge is
  persistent and the pill is the existing affordance; accepted.
- The permit-join banner remains a content banner; out of scope.
- No persistence of read/dismissed notices; the dialog is simply closed.

## Completion

Implemented and verified; no review approval is claimed.
