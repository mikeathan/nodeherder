# Spec: Category-based live reporting

ID: 014 | Status: Complete | Updated: 2026-10-16
Constitution: v1.0.0 (proposed) | Components: both

## Problem, scope, exclusions

Spec 012 introduced minimum-change rules but left fields unfiltered without manual
configuration. Existing category debounce controls history only. Users need quiet
numeric diagnostics automatically and clear, independent live/history controls.
Scope includes shared policy resolution, bounded live sampling, settings cleanup,
compatibility and responsive UI. No adaptive learning, automation timing changes,
new sensor classification rules, MQTT commands or history schema/retention changes.

## Scenarios and acceptance

- AC-01: Numeric read-only diagnostics inherit a configurable 30-second live interval
  without creating per-device overrides. Numeric measurements default to immediate.
- AC-02: Device field overrides beat global field overrides, which beat global category
  defaults. Zero explicitly disables that filter; missing/removing means inherit.
  Minimum change remains in native units, never generalized across category units.
- AC-03: Within an interval retain only the latest reading per field. At the original
  deadline publish that reading if it differs sufficiently from the last published
  value, even without another report. Frequent arrivals never postpone the deadline.
  Minimum change is measured from the last actual publication; intermediate readings
  never advance its baseline.
- AC-04: Raw state and relevant automation inputs remain immediate. Repeated physical
  events, discrete/binary states, writable evidence, invalid/unknown recovery and
  availability changes bypass live sampling. History remains independently sampled.
- AC-05: Disable, shutdown and offline transitions cancel pending live output. Config
  edits re-evaluate pending readings against current rules without discarding the
  final value or moving the last-publication anchor.
  No output after cancellation completes; state is bounded to known fields and one
  outstanding timer per device, not timers per incoming reading.
- AC-06: Default/device forms share one editor. Show effective inherited values and
  source before an override exists. Use Live updates, History recording, Field
  exceptions and Advanced headings; no raw config keys as labels. Explain that live
  settings do not throttle automation. Confirmation timeout belongs under Advanced.
- AC-07: Keyboard operation, labels, focus, invalid/offline feedback and no horizontal
  overflow at 390px and 1280px; removing a field exception restores inheritance.
- AC-08: Old stored history intervals, thresholds, history inclusion and timeout retain
  meaning. New defaults hydrate only missing live settings, not explicit zero/maps.
  Validation failures and failed saves do not publish active backend settings.

## Requirements and success

FR-01: Common policy resolution for category/field inheritance (AC-01/02/08).
FR-02: Independent bounded live publication (AC-03/04/05).
FR-03: Shared readable responsive settings and scoped removal of unused editors,
lookup paths, imports and registry entries (AC-06/07).
NFR-01: No new dependencies; no goroutine/timer per report, no unbounded history of
suppressed readings. Fake-clock and race tests verify lifecycle.
SC-01: A numeric diagnostic stream at 1 Hz produces at most one field publication
per configured interval after the initial snapshot, and flushes its final value.

## Entities and interfaces

Keep existing saved fields and WS save/snapshot contracts; add optional category and
field live interval maps to DeviceConfig. Intervals have explicit units/value; zero
means immediate. No secrets, access changes, history backfill or device writes.

## Assumptions

30 seconds is an initial editable diagnostic live default, not a sensor-specific rule.
Only read-only numeric diagnostics inherit it; battery-low and physical events remain
immediate. Category classification is existing normalized expose metadata.
