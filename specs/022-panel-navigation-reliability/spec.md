# Spec: Panel navigation reliability

ID: 022 | Status: Complete | Updated: 2026-10-17
Constitutions: parent/frontend v1.0.0 proposed.

Review of compact panel work (021) reproduced three existing defects: disconnect
notices cover the dock, invalid stored room indexes crash rendering, and arrow keys
change rooms behind the Appearance modal. A 12-hour clock also wraps at 320px.

- AC-01: Panel notices sit above the dock without moving content or obscuring navigation;
  dismissal/retry/permit-join remain available. Other app notices retain current placement.
- AC-02: Invalid/non-finite/fractional/unsafe saved indexes fall back to the first room;
  valid integers wrap safely, including negative/out-of-range indexes and zero rooms.
- AC-03: Room tabs implement roving focus, arrows/Home/End and tab-panel association.
  Global room shortcuts ignore modal/editable/interactive targets and modifier keys.
  Existing ordinary left/right shortcuts remain; selected tabs stay visible on navigation.
- AC-04: Clock stays on one line at 320/390px in 12-/24-hour layouts. Dock remains
  compact/touch accessible, content scrolls independently with notice present/absent.
- AC-05: Test/review, then commit/push the combined 021/022 work and create a PR with
  the behavior changes, verification and limitations described. No device commands.
