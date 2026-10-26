# ADR-002: Private per-instance board preferences

Date: 2026-10-26 | Status: Implemented technical decision; maintainer review pending
Scope: AC-34–37. User authorized the recommended Connections board and subsequent icon/navigation
changes. No constitutional ratification or independent technical approval is implied.

## Decision

Keep exactly one results widget per saved instance. Persist optional BoardPreference on that
instance in the existing encrypted repository. Array order and hidden state are exposed through
the existing protected integration API namespace. A preference-only revision and canonical
membership ETag allow atomic shared layout saves without modifying lifecycle revisions/generations.
No encrypted schema/AAD/bucket change is required. See contracts/connections-board.md.

Presentation, immutable projections/layout edits, page-owned reads and provider discovery each have
separate source owners. Use existing vuedraggable and shared UI/SVG components. Details and editing
continue in the established full connection page. No weather feature, resize engine or device/WS
projection is introduced.

## Alternatives and consequences

- Browser-only layout: less backend work, but loses shared host persistence and risks mixing private
  snapshots with browser persistence. Rejected for this installation-wide feature.
- Home settings/WS storage: already supports area ordering, but device-specific models and public
  device boundaries are unsuitable for private connection references. Reuse interactions only.
- New layout bucket/table: useful for independently identified multiple widgets per connection,
  but unnecessary now; changing the encrypted schema requires careful AAD migration. Defer until
  a real multi-widget requirement exists.
- Arbitrary rectangles/resizing: higher dependency and interaction cost; the user removed resizing.

Old records remain visible in creation-time order. New instances append automatically. Deletion
removes preferences naturally. Concurrent membership/layout changes reject stale saves; resource
refreshes do not. Shared layouts are intentionally not per-user. Older binaries may discard these
optional preferences when rewriting records; schedules remain compatible. Failed saves retain the
browser's in-memory draft. No provider requests occur when arranging cards or reading the board.
