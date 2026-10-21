# Spec: Faster CI and cleaner device controls

ID: 023 | Status: Implemented locally; CI rollout pending | Owner: Codex | Updated: 2026-10-20
Constitutions: parent/frontend 1.0.0 (proposed) | Components: CI/frontend

## Problem, scope, exclusions

PR run [39298242659](https://github.com/mikeathan/nodeherder/actions/runs/39298242659)
took 6m57s, dominated by 405s of serial desktop/mobile E2E (368s browser execution,
21s Chromium installation). Matching [main run 39299251928](https://github.com/mikeathan/nodeherder/actions/runs/39299251928)
took 7m15s, including 413s E2E. Backend ordinary/race jobs already ran in parallel
(88s/110s); automations/controllers are their slowest sampled test packages.

Device screenshots show heavy horizontal sliders, awkward mixed offline/online card
spacing and large footers. Users want natural card heights, compact metadata, easy
device navigation and no flashing command progress indicators.

Scope: isolated E2E partitions/diagnostics, device cards and horizontal controls,
device shortcuts, removal of command progress presentation and branch planning reuse.
Preserve all CI gates, test safety/coverage, entity vertical-slider design, device
semantics, filters/readings and authenticated routes. No backend, storage, protocol,
dependency, branch-protection or retry-policy changes.

This single branch spec consolidates unmerged drafts 023–028. Rejected equal-height
cards, taller footers and delayed command indicators are excluded from final acceptance.
Completed earlier specs remain unchanged.

## Scenarios and acceptance

### US-01 — Earlier complete verification (P1)

- AC-01: Desktop/mobile × two file-level shards covers every discovered case exactly
  once. Each runner owns a fresh mock hub/Vite pair; one worker and within-file order
  remain. PR/main/merge-group checks, assertions, deadlines and leak-cycle counts stay.
- AC-02: Any failed/cancelled/skipped required job prevents `Run Tests` succeeding and
  main tagging. Other E2E partitions finish for diagnostics. Timing JSON is retained
  for completed runs; failures retain uniquely named reports/traces/screenshots.
- AC-03: Document measured costs, local timings and estimated hosted improvement
  separately, including additional runner cost, queues and future shard imbalance.

### US-02 — Compact, usable device views (P1)

- AC-04: Cards retain natural unequal heights, equal column widths, consistent gaps,
  compact wrapping footers and unsplit offline/online panels at 320/390/768/1280/1920px
  in light/dark. DOM/keyboard reading order follows column flow; no readings are hidden.
- AC-05: Card and device-page horizontal sliders have thin muted rails, round handles,
  readable values/presets, visible focus and 44px targets. Device rows fit at
  320/390/768/1280px in light/dark; narrow presets wrap separately. Vertical entity
  controls retain their existing design and native gesture behavior.
- AC-06: Friendly-name card headers and entity-dialog device names open the correct
  device page, including offline cards. Ordinary navigation closes the dialog and
  releases focus/scroll locks; Back does not reopen it. Modified clicks keep native
  link behavior. Missing entities show no broken link. Navigation sends no commands.

### US-03 — Stable controls without command progress decoration (P1)

- AC-07: Pending commands add no Sending/waiting text, spinner, reserved slot, pulse,
  color change or fade to cards, rows, entity adjustments/power, Home or panel controls.
  Card/rail/footer and neighboring geometry stays stable during confirmation.
- AC-08: Preserve native values, one command per release, presets, keyboard/cancel,
  duplicate prevention, unknown/offline handling, confirmation and timeout recovery.
  Requested slider drafts remain distinct from confirmed readings; toggle text stays
  confirmed. Keep aria-busy and existing errors; no automatic retries or fake state.

### US-04 — One plan for the branch's evolving work (P2)

- AC-09: Guidance reuses one spec/plan/tasks set for current branch follow-ups, even
  after locally complete iterations. New work after completed earlier work gets a
  new spec. Consolidation retains final requirements/evidence and repairs references.

## Requirements and success

- FR-01: MUST satisfy AC-01–03 without reducing CI coverage or weakening gates.
- FR-02: MUST satisfy AC-04–06 with existing controls/router and theme tokens.
- FR-03: MUST satisfy AC-07/08 without changing shared command/draft lifecycle.
- FR-04: MUST satisfy AC-09 in workflow, root agent guidance and planning skill.
- NFR-01: Target longest hosted E2E partition below 180s on comparable warm Ubuntu
  runners, versus 405s before. This remains a rollout target, not a local guarantee.
- SC-01: Partition inventories match the full suite; isolated browser, Jest, lint,
  build and workflow checks pass. Retain existing intentional skips explicitly.

## Entities, interfaces and assumptions

`Run Tests` stays required; main-only `Tag Build` depends on it. Matrix/artifact names
include project/shard. Device commands keep existing native-unit payloads and timeout
semantics; device links reuse the existing route. No wire/storage/auth migration.
Four E2E runners need account capacity; columns may rebalance when real content changes.
Only isolated mocks are authorized for routine tests. No unresolved design questions.
