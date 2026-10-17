# Spec: Target distance measurement classification

ID: 017 | Status: Complete | Updated: 2026-10-16
Constitutions: parent/backend v1.0.0 proposed.

Target distance currently inherits diagnostic metadata and its 30-second live interval.
It measures the environment and must receive measurement semantics instead.

- AC-01: New target-distance entities classify as measurement even when bridge metadata
  says diagnostic. Existing stored diagnostic target-distance entities normalize on
  decode, retaining readings, units and access. Unrelated diagnostics remain diagnostic.
- AC-02: Measurements have zero default live delay and minimum change. Small actual
  distance changes publish immediately; unchanged state remains deduplicated.
- AC-03: Raw state and relevant automation deltas remain current independently of
  explicit live-output filtering. User-configured overrides remain authoritative.

No new default thresholds, intervals, UI, device commands, subscriptions or dependencies.
Diagnostic default remains 30 seconds. Existing measurement history policy applies to
reclassified distance, subject to metrics enablement and explicit history overrides.
