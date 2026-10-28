# ADR-001: One private Home layout with a legacy device projection

Status: Technical design selected for the user-authorized implementation; no constitutional
ratification or maintainer review claimed. Date: 2026-10-26.

## Decision

Store one optional versioned `homeLayout` aggregate in the existing settings database's AppConfig
record. It contains references, labels and presentation choices only. `AppConfigCache` is the
single mutation owner and uses one existing Bolt record transaction to persist both canonical
layout and its derived legacy device-area projection. Publish in-memory state only after success.
Do not depend on the optional integrations repository/keyring to edit/render device areas.

Expose Home only through protected `/api/home/layout`. Strip it from EVERY ordinary AppConfig
snapshot/save response before hub/API/WS encoding. Legacy `hub.dashboardGroups` contains only
device values projected from canonical groups. It remains stored for old-client/rollback use,
but cannot be mutated independently after a canonical layout is accepted. Legacy mutation methods
adapt edits to the canonical model in the same cache lock/record transaction, preserving integration
cards not explicitly removed with their group. Whole legacy import replaces device groups/cards;
retain existing integration cards in matching groups, and move displaced integration cards to
standalone blocks so a device-only import cannot silently delete their placement.

If `homeLayout` is absent, derive deterministic group/card references from sorted legacy areas,
preserving group order and expose order. GET does not write/migrate on disk. First accepted Home
save persists version 1 and the device projection. Legacy writes before that point retain their
existing representation; the content-derived ETag still detects concurrent changes.

## Alternatives

- Add integration refs to DashboardGroup/WS: rejected; broadens legacy payload exposure and
  mixes provider data with settings consumers.
- Separate overlay with independent device membership: rejected; two authorities diverge on
  cross-group moves, import and rename.
- Store all Home state in optional integrations DB: rejected; breaks independent device Home
  availability when integration storage/keys are unavailable and requires cross-store transactions.
- New database: unnecessary; existing settings Bolt record provides atomicity and backup scope.

## Consequences and recovery

Existing settings storage now also contains private layout references; no postcodes, credentials,
provider responses or resource snapshots are stored in this field. Public config projections must
be regression tested, including settings-save responses. Unknown schema/corrupt layout fails
visibly; never overwrite it with defaults. Back up settings.db before upgrading. An older binary
can use the device-only projection, but its subsequent settings writes may discard the unknown
Home field; restore the matching backup before re-upgrading to recover integration placements.
No in-place encryption/AAD/schema change to the integrations database.
