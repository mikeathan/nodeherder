# Spec: Scoped CI repair

ID: 003 | Updated: 2026-10-02 | Related: [CI](../002-ci-verification/spec.md)

User narrowed runtime changes to mapper synchronization and closing the key/value
DB when initialization fails. Automation/scheduler edits, driver replacement,
transaction-value copying and legacy fixture work were removed. Existing worker-test
repairs remain under spec 002. No deployment or household-data access.

- FR-01: Concurrent mapper configure/update/lookup preserves IEEE mapping,
  stale-entry removal and hashed fallback without map races.
- FR-02: Failed key/value initialization closes the opened DB, releasing its lock.
- AC-01: Concurrent mapper regression passes repeatedly under race detection.
- AC-02: Invalid bucket initialization followed by reopening the same temporary file succeeds.

Verify vet/build/ordinary backend tests and targeted races; do not claim full race
CI is repaired. Further runtime changes need separate user review/authorization.
