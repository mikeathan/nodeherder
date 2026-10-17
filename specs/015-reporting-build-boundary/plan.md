# Plan: Reporting build boundary and review fixes

Spec: [spec.md](spec.md) | Status: Complete | Updated: 2026-10-16
Constitutions: parent/backend/frontend v1.0.0 (proposed).

## Design

Keep backend/models/settings/reporting-defaults.json authoritative. A dependency-free
scripts/sync-reporting-defaults.mjs copies exact validated JSON bytes into checked-in
frontend/src/domain/reporting-defaults.json. --check fails on mismatch/missing output;
--write explicitly regenerates it. CI stack-tools runs both the check and source-owned
Node tests. Do not invoke root tools in frontend npm build hooks: that would recreate
the isolated-build failure. Frontend domain and mock tooling read the local artifact.
Existing Dockerfiles/Compose/CI container contexts remain unchanged.

Normalize notification ownership in AppConfigCache: snapshot input before override
save, then clone per listener for both defaults/overrides. Keep synchronous notification
and save-before-publication. Existing engine selection checks remain at service and
engine boundaries; document their distinct callers/reload race. Correct skip logging.

## Constitution Check

| Rule | Design evidence | Before implementation | Final |
| --- | --- | --- | --- |
| NH-01 | AC-01–05 and tasks before changes | PASS | PASS |
| NH-02 | No filtering/event/automation semantics changes | PASS | PASS |
| NH-03 | Defaults-only artifact; unchanged context/access | PASS | PASS |
| NH-04 | Build-only generated artifact with drift guard | PASS | PASS |
| NH-05 | Isolated build, sync regressions, ownership tests | PASS | PASS |
| NH-06 | Explicit generation; no active-state aliasing | PASS | PASS |
| BE-01 | Cache owns notification snapshots | PASS | PASS |
| BE-02 | Log/comments only in update/automation flow | PASS | PASS |
| BE-03 | Race checks on settings/listeners | PASS | PASS |
| BE-04 | Failed-save regression; schema unchanged | PASS | PASS |
| BE-05 | No route/assistant changes | N/A | N/A |
| FE-01 | No state/UI ownership changes | N/A | N/A |
| FE-02 | Same generated defaults; sync validation | PASS | PASS |
| FE-03/04 | No UI or network-status behavior changes | N/A | N/A |
| FE-05 | Artifact has compatibility defaults only | PASS | PASS |

## Contracts and decision

Runtime contracts and defaults unchanged. Artifact is a build-time derivative, not
another editable source. Alternative root Docker context expands deployment/CI scope
unnecessarily; handwritten copies without a CI guard could silently diverge. See
[ADR](adr-001-generated-defaults.md). No migration. Regenerate when canonical JSON
changes, check in both; rollback restores previous artifact and source together.

## Verification

AC-01: reproduce isolated frontend TS2307, then full isolated production build;
try actual frontend Docker build if Docker daemon is available.
AC-02: Node companion tests cover drift/missing files/check immutability/regeneration;
run generator --check, lint, Jest/build and CI actionlint.
AC-03: cache_test.go companion tests demonstrate listener isolation and no failed-save
notification before fix; run full Go tests/vet/build and affected race checks.
AC-04: inspect call sites; preserve existing dependency regressions.
AC-05: record current full mock Playwright results and distinguish the independent
reviewer's earlier evidence. Prior plans remain historical; add linked follow-up notes.
Actual results are recorded below. No tests ran against household devices.

## Risks and completion

Repository CI must enforce sync; standalone frontend builds intentionally do not
require the canonical backend file. Docker daemon/dependency access may limit actual
image verification; report limitations honestly. No commit/branch changes authorized.

## Actual results and final review

- Reproduced: frontend-only `npm run build -- --mode production` failed with TS2307
  on the cross-boundary backend JSON import. The fixture was copied to
  /private/tmp/nodeherder-frontend-context-teysvh4a/frontend, excluding installed deps,
  generated outputs and local/development/test environment overrides. Only node_modules
  was symlinked to existing locked dependencies; no backend directory existed.
- PASS after fix: that same isolated frontend production build on Node v24.21.0,
  using its generated local artifact. Existing DeviceCard conflict/large-chunk warnings
  remain. Full image assembly is UNRUN: `docker info` reports the Colima Docker socket
  does not exist. No VM, Docker daemon, services or deployment were started. Existing
  CI container jobs still build ./frontend with no wider context.
- PASS: isolated mock state initialization on Node 24 without a backend directory,
  supplying only its existing docs/hub_state.json fixture dependency. The mock still
  requires that documented repository fixture; no claim of fully standalone mock
  packaging is made.
- Reproduced: TestConfigurationListenersOwnIndependentSnapshots failed for both
  defaults and overrides before the fix. Those assertions pass unchanged afterward;
  TestFailedConfigurationSaveDoesNotNotifyListeners also passes. Each listener owns
  a deep snapshot; override capture is independent from the original caller.
- PASS: `go test -count=1 -timeout=3m ./...` on final functional changes.
- PASS: `go test -race -count=1 -timeout=3m ./models/settings ./internal/services ./internal/automations ./internal/controllers ./models/devices`.
- PASS: `go vet ./...` and `go build ./...`.
- PASS: Node 24 `npm test -- --runInBand`: 32 suites / 242 tests.
- PASS: Node 24 `npm run lint`: zero errors, seven pre-existing script warnings.
- PASS: Node 24 `node --test scripts/sync-reporting-defaults.test.mjs scripts/dev.test.mjs`:
  nine tests; generated artifact drift/missing output/invalid JSON/check immutability,
  deterministic regeneration, launcher failures and descendant cleanup.
- PASS: generator --check and mock state syntax check.
- PASS: `go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12` on CI changes.
- PASS: Node 24 `CI=true npm run test:e2e`: full 72/72 desktop/mobile tests, 1.8 minutes.
  CI mode refused reuse of existing servers; tests ran against owned mock/Vite servers.
  This independently verifies the reviewer's earlier 72/72 report. Historical spec
  012/014 check notes remain unchanged, with links to this follow-up evidence.
- PASS: git diff --check and final source/CI review: no frontend src/tools imports
  of backend reporting defaults; generated file included by existing COPY/.dockerignore;
  sync check is included in stack-tools, required by the aggregate Run Tests gate.

The repeated dependency selection guard remains intentional: service prefilter avoids
unnecessary dispatch; the engine boundary protects seed/direct callers and replacement
between checks. Only its explanatory comment changed. Skip logging now includes both
disabled and irrelevant updates. No runtime policy values, filtering semantics, new
network access, dependencies or per-reading work were introduced.

Final applicable constitutional gates PASS with the evidence and Docker limitation
above. The only additional configuration copies occur on settings changes, not on
sensor readings. Documentation describes canonical/editable versus generated data.
All work remains uncommitted on fix/pin-view-wording, as previously instructed; no PR,
branch rename or deployment was performed. Specs 012/014 must include this build fix
before merge. No exception or maintainer approval is claimed.
