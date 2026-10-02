# Tasks: Scoped CI repair

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md)

- [x] Remove expanded runtime/dependency/legacy-fixture changes at user request.
- [x] Retain mapper mutex and failed-initialization DB close; FR-01/02.
- [x] Verify mapper concurrency, failed-init reopen, worker and authorized mock regressions,
  vet/build/ordinary suite; FR-01–03. Mock-only follow-up uses the existing CI-repair work.
- [x] Record actual results and remaining domain/Bolt failures for user review;
  no production fixes or suppressed CI checks in mock follow-up.

Authorized Bolt follow-up (same CI work): FR-04 legacy fixture/import replacement
and storage/ordinary/toolchain checks completed; remaining full-race failures
reported separately. Ubuntu execution remains unverified; no deployment performed.

FR-05 low-risk follow-up: WS map mutex and atomic logger flag implemented/verified
with existing tests plus focused concurrency cases; remaining runtime failures
stay visible. Automation, Z2M and device lifecycle changes remain deferred.

FR-06 test-only synchronization completed in existing test files; numeric/count/
permit-join expectations preserved. Repeated focused checks and ordinary suites
pass; expiry/configuration and other runtime race failures remain reported, not hidden.

FR-07 authorized registry/creation callback follow-up: Seed outside registry lock,
protect reads/publication; verify callback re-entry and concurrent configuration
lookup/registration in existing tests completed. Repeated focused races, ordinary
tests, vet/build pass; full-race failures reported in plan, not suppressed.

FR-08: bridge snapshot ownership and cleanup duration synchronization; add existing-
file regressions, rerun permit expiry/cleanup races, ordinary checks and full races:
completed. Targeted races/ordinary checks pass; four domain packages still fail
full races. Task shutdown/duplicate workers require separate lifecycle review.

FR-09 authorized cleanup shutdown: cancel/join workers and timers; existing-file
lifecycle regressions and retention/CI verification completed. Repeated lifecycle
and store races pass; four unrelated domain packages fail full race checks.
In-flight Prune must finish before Stop/reload returns; limitation recorded in plan.

FR-10 LastSeen-only: setter/local payload timestamp; existing-file concurrent-read
regression and ordinary/race verification completed. Focused races/vet/build/
ordinary suite pass; four domain packages still fail full races (see plan).

FR-11 monitor ownership, disable/reset/re-entry regression in existing file, verify
unchanged semantics and report unrelated race failures completed. Targeted races,
Go1.24, vet/build/ordinary checks pass; four packages still fail full races.
In-flight callback/update and transition limitations recorded; no broad device lock.

FR-12 availability accessors in Seed/Update only; existing assertions and focused/
ordinary/full-race checks completed. Ordinary/vet/build and separate seed/disabled/
monitor race cases pass; broader focused and full races still fail direct-reader/
entity/automation cases. Tests unchanged; exact results in plan.

FR-13 entity ownership: add existing-file model/semantic regressions, demonstrate
pre-fix race, synchronize value access only, verify and report remaining failures:
completed. Repeated focused/operation races, vet/build/ordinary pass; three packages
still fail full races. Pointer/composite ownership limitations recorded in plan.

FR-14 existing availability tests: safe reads, controller completion signals,
fixture cleanup completed; expectations preserved, no production changes. Repeated
focused/ordinary checks pass; full races now fail automations/controllers only.

FR-15 enable-state/accessor/JSON synchronization and existing-file regressions:
completed; scheduler, trigger/action lifecycle and configuration unchanged.
Focused races/Go1.24/vet/build/ordinary pass; remaining full-race failures in plan.

FR-16 delayed scheduling/cancellation/completion: regression tests, narrow repair,
compatibility and full-race verification completed; configuration remains deferred.
Repeated native/Go1.24 automation races, ordinary/vet/build pass. Controller action
configuration races remain; exact checks and cancellation limitations in plan.

FR-17 rapid physical-event integration tests/feedback characterization in existing
processor test file: completed; test-only, existing assertions unchanged. Twenty
native/Go1.24 focused runs, affected races, vet/build/full ordinary pass; feedback
and transport/concurrent-delivery limitations recorded in plan.

FR-18 approved atomic configuration binding: regressions, snapshot implementation,
failure/re-entry and compatibility verification completed; reuse existing files.
Full backend race/ordinary/vet/build pass; Go1.24 affected/focused races pass. Failure
and ownership boundaries recorded in plan; no live-device validation or performance claim.

FR-19 authorized audit follow-up: add first-update/disable/re-entry regressions;
implement single-owner creation and queued config/update delivery; verify focused
and full backend checks: completed. User-requested DRY cleanup uses one registry and
shared routing/delivery helpers. Reload readiness and preset coverage remain next.

FR-20 authorized ready-generation follow-up: add reload/error/scheduler/storage
regressions; implement prepared generation publication and resource rollback;
verify failure boundaries, original semantics and toolchains: implementation and
checks completed. Final audit found executable explicit-source-disable fallback;
user-approved exception implemented with red-before-green regressions, fail-closed
runtime removal, retained recipes, cleanup retry and re-enable coverage. Preset action
coverage remains separate; do not suppress failed assertions or race detection.

FR-21 requested clean-code follow-up: review changed ownership/duplication, centralize
management gates and constructors, replace unnecessary shared configuration with
worker-local snapshots, remove unused cooldown state; preserve regression assertions
and rerun focused/full checks: completed. No new feature or lifecycle-policy change.
FR-20 disable follow-up verified separately; live-device validation remains unrun.

FR-22 review follow-up: add dropped-reload and Hub-snapshot regressions (red first);
defer busy reloads until gate release; clone Hub in LoadAppConfig; gofmt touched
storage file; full race/vet/build and repeated focused races: completed. Go1.24.8
toolchain run and live-device validation remain unrun.
