# Spec: Meaningful sensor reporting and automation routing

ID: 012 | Status: Complete | Owner: Codex | Updated: 2026-10-16
Constitution: 1.0.0 proposed | Components: backend/frontend

## Problem, scope, exclusions

Read-only Z2M observations show a presence sensor's link quality oscillating 134↔138,
three changed reports in one second, and illuminance increments of 1–3 lx. See
[observations](observations.md). Current raw changes all broadcast and dispatch through
an enabled automation; category debounce samples metrics only. Target distance is
classified diagnostic and excluded from history. Raw state must remain authoritative.

Scope: deterministic numeric observation thresholds, per-device/global overrides,
metric eligibility of distance, fast routing of automation inputs, accurate docs.
Excludes live household configuration/device commands, adaptive statistical learning,
blanket packet throttling, changes to automation thresholds, timer-based batching,
and redesigning the existing metrics sampler. Pin wording remains uncommitted.

## Scenarios and acceptance

### US-01 — Useful observations without jitter (P1)

Verify with synthetic/anonymized replays, never live commands.

- AC-01: Read-only numeric jitter below a configured absolute threshold produces no
  live/UI/activity updates; raw entity state remains fresh. Compare against last
  published value, so gradual drift eventually passes. Equal-threshold changes pass.
- AC-02: Presence/contact/enums, repeated physical events, requested writable command echoes,
  unknown fields and missing numeric units retain existing semantics.
- AC-03: First seeded observation establishes baseline; significant final changes are
  delivered immediately even if the source becomes silent. No delayed queues/timers.
- AC-04: Threshold precedence is device field → global field → no filter. No property names or operational thresholds are hardcoded in reporting logic.
  Zero disables numeric significance filtering. Reject negative/nonfinite overrides.
- AC-05: Any field, including older diagnostic distance metadata, can be included/excluded
  from metric history through a configured per-field history override when device
  metrics are enabled. New read-only distance exposes follow generic measurement classification. No historical backfill.

### US-02 — Avoid unrelated automation work (P1)

- AC-06: A configured recipe's unrelated changed property avoids evaluation/context
  creation. Primary trigger and condition dependencies still process raw inputs.
- AC-07: Recipe replacement/reload rebuilds routing; deletion/disable retains existing
  behavior. Manual triggers and repeated physical events remain unchanged. Unknown
  unconfigured/custom automation implementations retain a conservative fallback.
- AC-08: Mixed payloads preserve relevant changes; presence conditions still re-evaluate
  on relevant illuminance changes; cached events are never replayed by those changes.

### US-03 — Configure and recover (P2)

- AC-09: Existing device/global settings expose minimum-change overrides with units,
  accessible keyboard controls at 390px and 1280px; removing entries restores inheritance.
  Existing settings payloads remain valid; old clients may omit the optional map.

- AC-10: Cached unchanged writable configuration in a noisy mixed payload does not
  force broadcast/storage. An explicit UI command marks one configured-duration confirmation
  interest per expose before publish; only an actual subsequent report consumes it.
  Expired/consumed interests do not emit or evaluate automation. Request never sets state.

## Requirements and success

- FR-01: One backend significance policy for UI observations, separate from raw state,
  automation eligibility and metric sampling (AC-01–05).
- FR-02: Derived immutable automation input routing on recipe configuration (AC-06–08).
- FR-03: Optional persisted threshold map with validated writes and UI (AC-04/09).
- NFR-01: O(1) significance lookup per field, O(changed keys) unrelated automation
  routing; bounded baseline state per known expose. No new dependencies/workers/timers.
- SC-01: Recorded LQI oscillation replay yields zero live updates after seeded baseline;
  10,000 irrelevant updates invoke zero recipe evaluations. Tests/race/build/lint pass.

## Entities and external interfaces

`reportingThresholds`: optional property→finite nonnegative delta in native units.
`metricOverrides`: optional property→boolean, device then global then schema category.
`commandConfirmationTimeoutMs`: optional positive whole milliseconds, device then global
then the declarative compatibility bootstrap data. The browser and server resolve the
same saved configuration; no second operational timeout constant in client/server code.
All values can be added, removed and adjusted in the shared hub/device UI. There are no
implicit sensor-name/units defaults. Blank restores inheritance; zero disables threshold
filtering. The bootstrap JSON contains only the legacy command timeout, not field names.
No unit conversion/smoothing or relative tolerance invents values.

## Questions and assumptions

User confirmed system/automation performance is in scope. Implement lightweight
significance filtering first; meaningful changes can still arrive rapidly. No hard live
rate cap is claimed. Metrics interval/retention/timestamps stay unchanged. Automation
consumers always use raw values, including subthreshold changes to relevant conditions.

Amendment requested during implementation: user rejected code-defined operational
thresholds and command timeout. Replace builtin policy logic with generic configured
maps, configurable history eligibility and shared declarative compatibility data.
Cleanup stays in affected settings/reporting/command paths; existing assertions are
preserved when consolidating source-owned tests.
