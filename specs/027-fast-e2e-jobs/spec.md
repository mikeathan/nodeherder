# Spec: Fast, deterministic browser E2E jobs

ID: 027 | Status: Implemented | Owner: NodeHerder contributors | Updated: 2026-10-22
Constitution: proposed 1.0.0 | Components: CI verification tooling

## Problem, scope, exclusions

Main-branch runs take up to ~15 minutes even though every test finishes quickly. In
[run 39781252241](https://github.com/mikeathan/nodeherder/actions/runs/39781252241) the
four browser jobs spent:

- 11m24s and 11m10s in `npx playwright install --with-deps chromium` (apt mirror
  variance: `fonts-wqy-zenhei` alone took 2m51s), while
- the same step took 18–21s on the other two jobs, and
- the browser tests themselves took ~2 minutes per shard.

The wall clock was therefore set by hosted apt downloads, not by the suite: the run
ended at 14m15s because two installs were slow.

Reduce and stabilize CI wall clock for the browser jobs without changing what the
tests verify. Exclude test content, assertions, sharding, retries, viewport coverage,
product behavior, and release versioning.

The first run after this change surfaced a second problem on main: the 50-cycle
`EntityViewDialog` leak assertion failed in [run 39790664098](https://github.com/mikeathan/nodeherder/actions/runs/39790664098)
with `nodes` 1520 → 1512 → 2075 against a `warm + 100` bound, while the same commit
passed two hours earlier. The extra 563 nodes and 22 listeners at the final sample are
a closed dialog that was still reachable, not growth across the 50 cycles. Make that
measurement deterministic without weakening what it proves.

## Scenarios and acceptance

### US-01 — Browser jobs no longer install browsers (P1)

- AC-01: Given a clean runner, when the browser jobs run, then no apt or Chromium
  download happens and tests start within seconds of dependency installation.
- AC-02: The same four partitions (desktop/mobile × shard 1/2) still run the identical
  suite, reports and artifacts, and still gate through `Run Tests`.

### US-02 — Version drift fails loudly (P1)

- AC-03: Given `@playwright/test` no longer matches the pinned browser image, when a
  browser job runs, then it fails before testing with a message naming both versions
  and the file to update.

### US-03 — The leak check measures settled state (P1)

- AC-05: Given a run where a closed dialog is still pending release, when the heap is
  sampled, then sampling settles first and a real per-dialog leak is still detected.
- AC-06: The leak assertions keep their existing bounds (`nodes <= warm + 100`,
  `listeners <= warm + 5`, `< 2 MiB`), so no guarantee is relaxed.

### US-04 — Names describe what runs (P2)

- AC-04: Job and step names state the work actually performed; `Run Tests` keeps its
  name because it is the required status check.

## Requirements and success

- FR-01: Run browser jobs inside a Playwright image matching the installed version;
  acceptance: AC-01, AC-02.
- FR-02: Add a fail-fast version guard; acceptance: AC-03.
- FR-03: Sample the leak check after release has settled and after pending reads
  finish; acceptance: AC-05, AC-06.
- NFR-01: No new dependency, secret, or self-hosted resource; no change to test
  assertions, shard count, retry policy, or timeouts other than a tighter job cap.
- SC-01: Slowest browser job well under the previous worst case, with variance removed;
  verified by the first run of the changed workflow.

## Entities and external interfaces

The external interface is the workflow contract: job names, the `Run Tests` aggregate,
uploaded `playwright-timings-*` and `playwright-report-*` artifacts, and the
`@playwright/test` version in `frontend/package.json` which the image tag must mirror.

## Questions and assumptions

No blocking questions. Assumption: GitHub-hosted container jobs are acceptable for
this repository; the Playwright image is published per version and architecture.
