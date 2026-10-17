# Plan: Compact panel toolbar

Spec: [spec.md](spec.md) | Status: Complete | Updated: 2026-10-17
Constitutions: parent/frontend v1.0.0 proposed; backend N/A.

## Design

PanelMode.vue owns dock sizing: inline icon/label layout, smaller gap/padding/radii,
44px control height, floating centered mobile dock, bottom safe-area margin. Keep dock
in its existing grid row so content never hides beneath it. DefaultViewButton.vue hides
only dock label on mobile; accessible name/title and pressed state remain. Active view
uses filled accent; pin preference uses a quieter outlined/tinted state. Reuse theme
variables; no JS state, timer, listener or dependency added. No contract/migration/ADR.

## Constitution Check

| Rule | Evidence/design | Before implementation | Final |
| --- | --- | --- | --- |
| NH-01 | AC-01–03 before edits | PASS | PASS |
| NH-02 | Presentation only | PASS | PASS |
| NH-03 | No data/trust changes | PASS | PASS |
| NH-04 | Existing dock/shared button ownership | PASS | PASS |
| NH-05 | Existing panel companion browser tests | PASS | PASS |
| NH-06 | No persistence/resources changed | PASS | PASS |
| FE-01 | Existing pin/navigation state | PASS | PASS |
| FE-02 | No wire changes | PASS | PASS |
| FE-03 | Existing confirmed pressed state | PASS | PASS |
| FE-04 | 44px targets, keyboard names, 320/390/1280px | PASS | PASS |
| FE-05 | No sensitive data changes | PASS | PASS |

## Verification

Update panel.spec.ts intentional visual-state assertion to distinguish pin from active
view; preserve all navigation/persistence checks. Add bounding-box/tap-size, mobile
label and status/keyboard/exit assertions. Run Jest/lint/build, panel/layout browser
checks and inspect screenshots on desktop/mobile in light/dark themes. Results (2026-10-17):

- `npm test -- --runInBand --silent`: PASS, 32 suites / 247 tests.
- `npm run build`: PASS; existing DeviceCard naming/large-chunk warnings remain.
- `npm run lint`: PASS, seven existing script warnings.
- `npm run test:e2e -- e2e/panel.spec.ts e2e/layout.spec.ts`: final isolated run
  10/10 PASS. Verified 320/390/1280px, dock height <=60px and every control >=44px,
  no horizontal overflow, keyboard Rooms/Status/pin, exit and pin persistence.
- Inspected light desktop and dark pinned 320px mobile screenshots. Dock remains in
  its own grid row, with safe-area margin and content scrolling above it. Quieter pin
  outline differs from the filled active view, including hover.
- `git diff --check`: PASS. No runtime JS state, subscription, timer or dependency added.

Initial default-port browser run reused an existing development server and landed on
its local sign-in screen: three assertions failed and one test was interrupted. Stopped
that test process; did not stop/change the user's servers. Reran in a copied temporary
workspace on mock-only ports 4200/4210. Temporary config adjusted only test origins/ports
and Vite's font serving allow-list for symlinked dependencies; production source/tests
were unchanged. First copied run passed but blocked fonts; final run loaded real fonts
and passed all ten checks. No mock-port/config changes are included in the branch.

Existing selected-color equality assertion was deliberately replaced with pin-vs-view
background separation, visible pin outline and active Rooms assertion; all persistence,
direct-link, fallback and exit checks remain. Backend/full browser suites were not
rerun for this scoped presentation-only change. No household commands were tested.
Follow-up spec 022 fixes review findings and records final combined verification and publication on fix/compact-panel-toolbar.
No household tests. Leave changes uncommitted unless requested.
