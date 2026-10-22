# Plan: Fast, deterministic browser E2E jobs

Spec: [spec.md](spec.md) | Status: Implemented | Author: NodeHerder contributors
Updated: 2026-10-22 | Constitutions: parent 1.0.0

## Context and design

`frontend-e2e` runs four matrix jobs (desktop/mobile × shard 1/2) on `ubuntu-latest`
and installs browsers with `npx playwright install --with-deps chromium`. Measured
step times on main: 18s, 20s, 11m10s, 11m24s — the apt mirror, not the suite, decides
the run length. Tests take ~2 minutes per shard.

Design: run those jobs in `mcr.microsoft.com/playwright:v1.56.1-noble`, which ships
Chromium, the headless shell, ffmpeg and every system library. Browsers resolve from
`/ms-playwright`, so the install step is deleted rather than cached (a cache would not
remove the apt work, which is the actual cost). `options: --ipc=host` follows
Playwright's container guidance for Chromium.

Tag drift is the one new failure mode: if `frontend/package.json` moves and the tag
does not, browsers and the test runner mismatch. A guard step compares
`node_modules/@playwright/test/package.json` with the pinned tag after `npm ci` and
fails with the exact file to update. The image tag is intentionally literal because
`container.image` cannot read `package.json` before the job starts.

Job timeout drops 20 → 15 minutes: with no install, a shard is compile + `npm ci` +
~2 minutes of tests, and 15 minutes still absorbs queue/slow-runner variance while
failing faster on a real hang.

Renamed for accuracy: `Stack Manager (Ubuntu)` → `Deployment Tooling (Ubuntu)`,
`Frontend Tests and Build` → `Frontend Tests, Lint and Build`, the E2E matrix label now
states project and shard only, and steps no longer claim to install or warm things
they do not. `Run Tests` is untouched because it is the required status check.

Alternatives rejected: caching `~/.cache/ms-playwright` (saves only the 6-second browser
download, not apt); dropping `--with-deps` on the hosted image (depends on runner-image
contents and fails silently after image updates); fewer shards (does not fix the
worst-case install); installing a hand-picked apt package list (fragile).

## Constitution Check

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | Spec records measured reproduction and outcomes | PASS | PASS |
| NH-02 | No device or automation semantics touched | N/A | N/A |
| NH-03 | No auth, secret, or data change; published image only | PASS | PASS |
| NH-04 | Workflow-only change; Playwright config untouched | PASS | PASS |
| NH-05 | Tests, shards, artifacts and gate unchanged; local container run | PASS | PASS |
| NH-06 | No persistence or resource-growth change | N/A | N/A |
| FE-02 | Playwright version pin/guard documented | PASS | PASS |
| FE-03 | No UI behavior change | N/A | N/A |

## Contracts, data, rollout

Workflow contract: job display names change (none are required checks except
`Run Tests`), artifacts keep their names, and the image tag must track
`frontend/package.json`. Rolling back means restoring the previous job definition; no
data or deployment change. Updating Playwright now requires a two-line change (package
version + image tag), enforced by the guard.

## Traceability and verification

| IDs | Paths | Procedure | Actual result |
| --- | --- | --- | --- |
| AC-01 | .github/workflows/ci.yml | Run the desktop shard 1/2 inside the pinned image locally | PASS: 52 passed, 9 skipped in 1.4 minutes, Go hub fixture included |
| AC-02 | .github/workflows/ci.yml, docs/ci.md | Compare matrix, reporters, artifacts and gate | PASS: unchanged partitions and artifact names; suite not modified |
| AC-03 | .github/workflows/ci.yml | Guard step logic and image/package versions | PASS: image v1.56.1 == @playwright/test 1.56.1; mismatch exits 1 |
| AC-04 | workflow, docs/ci.md | Name review | PASS |

Local validation (macOS arm64, Colima, Docker, image
`mcr.microsoft.com/playwright:v1.56.1-noble`):

- Image contents: `/ms-playwright/chromium-1194`, `chromium_headless_shell-1194`,
  `ffmpeg-1011` present; image Node v22.20.0 (job still installs Node from `.nvmrc`).
- Containerized run from a clean copy of the repo (Go 1.24.8 installed in the
  container, `npm ci`, then `--project=desktop --shard=1/2`): 52 passed, 9 skipped,
  1.4 minutes — including the Go automation fixture.
- `actionlint`, workflow YAML parse, and version-guard cross-check passed.
- Not run: the GitHub-hosted job itself until this PR runs; that run is the timing
  evidence for SC-01.

## Debt, risks, exceptions

The image tag is a second place to bump on Playwright upgrades; the guard converts that
into a clear failure instead of a flaky run. Container jobs cannot use `host.docker.internal`
for the stack, which is irrelevant here because the isolated servers run inside the same
container. No exceptions requested.

## Completion

Workflow and CI documentation updated. No product code, version, or test changed.
