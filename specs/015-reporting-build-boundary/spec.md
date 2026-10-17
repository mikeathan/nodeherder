# Spec: Reporting build boundary and review fixes

ID: 015 | Status: Complete | Updated: 2026-10-16
Constitution: v1.0.0 proposed | Components: both

## Problem and scope

The frontend imports backend reporting-defaults.json outside the frontend Docker
context, breaking production images despite repository builds passing. The mock
server has the same coupling. Review also found inaccurate skip logging and settings
listeners receiving inconsistent pointer ownership. Preserve all existing behavior,
branch and uncommitted work; no deployment, commits or household commands.

## Acceptance

- AC-01: Frontend production build succeeds with only frontend/ present, without
  backend/ access. The mock server also uses frontend-local bootstrap data.
- AC-02: Backend JSON remains canonical. A deterministic generator produces the
  checked-in frontend artifact. CI rejects missing/stale artifacts without silently
  regenerating them; explicit regeneration is documented and tested.
- AC-03: Each configuration listener receives its own owned, persisted-value
  snapshot for both defaults and device overrides. Listener/caller mutation cannot
  change active settings or another listener's snapshot; failed saves notify nobody.
- AC-04: Trace messages accurately cover disabled and irrelevant updates. Retain and
  explain the engine boundary guard, including direct calls and recipe replacement.
- AC-05: Report actual current verification separately from reviewer-reported
  72/72 browser evidence; no retrospective claim that earlier checks were run.

## Quality and interfaces

No new dependencies, network entry points, runtime defaults or configuration schema
changes. Generated data contains compatibility defaults only, no household settings.
Docker context stays frontend/. Source-owned tests use temporary files/fake stores.
