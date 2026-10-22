# Plan: Reliable proxy peer discovery for local sign-in

Spec: [spec.md](spec.md) | Status: Implemented | Author: NodeHerder contributors
Updated: 2026-10-22 | Constitutions: parent 1.0.0

## Context and design

Spec 026 guessed the network name from the checkout directory. Compose instead
normalizes the project name (lowercase, invalid characters removed), so the guess only
worked for a directory called `nodeherder`. Verified locally with a mixed-case temporary
directory: `compose create` produced `tmpwef8xaqbmj_nodeherder` while the script looked
for `tmp.wef8xaqbmj_nodeherder`, warned, and injected nothing.

Design:

1. `docker compose config` prints `name: <project>` for the normalized project, so read
   that instead of deriving it from `$ROOT_DIR`.
2. List the project's networks by Compose's own label
   (`label=com.docker.compose.project=<project>`), then read each gateway. This also
   covers a default network and multi-network projects without knowing their names.
3. Keep the frontend container address lookup, dedupe peers, and keep single-host
   `/32` / `/128` prefixes and the existing warning path.
4. Print the generated override path so operators know the value is not in an env file,
   and document the `docker inspect` command.

Alternatives rejected: parsing `docker compose config --format json` with `sed`
(`config` YAML already exposes `name:`); reading `config --networks` (returns the
Compose key, not the resolved Docker name); trusting all project subnets (wider than the
gateway); `docker network ls` without a label filter (would trust unrelated networks).

## Constitution Check

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | Spec records the reproduction and outcomes | PASS | PASS |
| NH-02 | No device semantics touched | N/A | N/A |
| NH-03 | Only project gateway addresses trusted; no wildcards; explicit config wins | PASS | PASS |
| NH-04 | Existing script boundary; no new dependency | PASS | PASS |
| NH-05 | Script tests plus a real container run | PASS | PASS |
| NH-06 | Generated file is per-run and gitignored | PASS | PASS |
| BE-03 | No concurrency or lifecycle change | N/A | N/A |

## Contracts, data, rollout

Deployment-only change: `start`/`rebuild` now resolve the project name from Compose.
Existing deployments produce the same peers (the directory is already `nodeherder`), and
an explicit `TRUSTED_PROXY_CIDRS` still wins. Rollback reverts the script; no data,
image, or API change.

## Traceability and verification

| IDs | Paths | Procedure | Actual result |
| --- | --- | --- | --- |
| FR-01, AC-01/02 | scripts/manage-stack.sh | Unit tests for normalized project name, label filter, duplicate peers | PASS |
| FR-02, AC-03 | scripts/manage-stack.sh | No-gateway and malformed-value tests | PASS: warning, no override file, stack starts |
| FR-03, AC-04 | scripts/manage-stack.sh, docs/setup.md, docs/runtime-reliability.md | Message and docs review | PASS |
| NH-05 | Compose + Docker | Real fixture run with a mixed-case directory | PASS: container env shows `TRUSTED_PROXY_CIDRS=172.19.0.1/32` |

Execution (2026-10-22, Docker Compose v5.1.0, local container run):

- `node --test scripts/manage-stack.test.cjs`: 64 passed.
- Real fixture stack (backend + stubs, network `nodeherder`, mixed-case directory):
  `start backend` reported `Trusted proxy peers: 172.19.0.1/32`, wrote
  `.generated/compose.proxy-allowlist.yml`, and the container's `Config.Env` contained
  `TRUSTED_PROXY_CIDRS=172.19.0.1/32`.
- `bash -n`, `shellcheck`, `actionlint`, `git diff --check`: passed.
- Not run: the household stack; GitHub CI runs on the PR.

## Debt, risks, exceptions

Peer addresses still change when Compose recreates networks, which the per-run
recomputation handles. The script now depends on `docker compose config` emitting a
`name:` line, which is stable Compose v2 behavior. No exceptions requested.

## Completion

Script, tests and deployment docs updated. No auth, product, or data change.
