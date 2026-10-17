# Spec: Compact panel toolbar

ID: 021 | Status: Complete | Updated: 2026-10-17
Constitutions: parent/frontend v1.0.0 proposed.

- AC-01: Panel dock uses a compact single row, at most 60px high excluding bottom
  safe-area spacing. Controls retain at least 44px touch targets on 320/390px phones.
- AC-02: Mobile pin is icon-only with full accessible name/title and pressed state.
  Rooms, Status and Exit stay labelled. Pinned preference is visually distinct from
  the active view. Desktop keeps text; shared header pin behavior is unchanged.
- AC-03: Pin persistence, room/status navigation and exit remain functional with
  keyboard focus. No overflow; content scrolls independently, dock never covers it.

No navigation, storage, automation or device semantics change. Preserve safe-area inset.
