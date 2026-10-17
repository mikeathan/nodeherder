# Plan: Target distance measurement classification

Spec: [spec.md](spec.md) | Status: Complete | Updated: 2026-10-16
Constitutions: parent/backend v1.0.0 proposed; frontend N/A.

## Context and design

Use the existing measurementWhitelist in models/devices/device.go as the classification
source of truth. Add target_distance there. Entity JSON decoding upgrades diagnostic
entries found in that same measurement whitelist; this handles stored snapshots before
service construction without per-report work or destructive database migration.
Keep explicitly configured reporting/history overrides and all other entity fields.
New bridge construction already uses that whitelist. No automatic measurement filter.

Alternative: fixing discovery alone leaves persisted devices stale; filtering policy
special cases would leave UI/history categories inconsistent. Normalize at decoding.
JSON shape is unchanged. Old binaries retain stored categories until bridge rediscovery;
rollback changes no readings/settings. No live device changes or storage deletion.

## Constitution Check

| Rule | Evidence/design | Before implementation | Final |
| --- | --- | --- | --- |
| NH-01 | AC-01–03 and tasks recorded first | PASS | PASS |
| NH-02 | Category correction; automation precedes live filters | PASS | PASS |
| NH-03 | No access or payload trust changes | PASS | PASS |
| NH-04 | Domain-owned classification and decode compatibility | PASS | PASS |
| NH-05 | Source companion tests; isolated checks | PASS | PASS |
| NH-06 | Retain readings/config; no background resources | PASS | PASS |
| BE-01 | Domain model owns category | PASS | PASS |
| BE-02 | Verify immediate distance/dedup/automation | PASS | PASS |
| BE-03 | No new shared state/workers; affected race checks | PASS | PASS |
| BE-04 | Distance now uses existing measurement history policy | PASS | PASS |
| BE-05 | No routes/assistant changes | PASS | PASS |

## Traceability and verification

AC-01: device_test.go bridge and persisted JSON cases.
AC-02: reporting_policy_test.go defaults; device_lifetime_test.go immediate changes.
AC-03: lifetime raw/automation checks and existing explicit-filter regressions.
Run full Go tests/build/vet, affected race tests, diff/complexity review. No household
checks. Frontend unchanged; its checks are outside this change. Results pending.

## Risks and completion

Live distance becomes more frequent by design; user overrides can reduce UI noise.
Measurements are not given automatic thresholds. Existing historical data is retained;
new samples follow enabled metrics sampling policy. Work remains uncommitted.
