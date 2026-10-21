# Plan: Runtime security and resource lifecycle

Spec: [spec.md](spec.md) | Status: Implementation verified; container/deployment checks unrun; MCP auth deferred | Author: Codex | Updated: 2026-10-21
Constitutions: proposed parent/backend/frontend v1.0.0; no adoption/approval claimed.

## Context and design

Keep existing ownership boundaries. Use bbolt's concurrency/transaction guarantees;
remove redundant wrapper locking, clone Get values inside View, and delete via a cursor
without skipping adjacent keys. Storage callbacks borrow bytes only within transactions.

WebSocket transport uses small byte buffers and a 256-message queue. Overflow closes the
slow connection so frontend reconnect/snapshot recovery can reconcile. Retain 1 MiB ingress.
The existing useHubConnection owner refreshes the snapshot on reconnect, coalesces
in-flight live device deltas before applying it, and retains newer config changes;
its per-request subscription/timer are disposed. Auth middleware supplies stream validation/lifetime hooks reused by WS and SSE; expiration,
revocation and disconnect own bounded timers/subscriptions. Blacklist cleanup has no
unowned periodic worker. Origin validation uses configured origins plus same-origin.

HTTP ingress gets one shared size-limit middleware after authentication and before
handler mutation (wrapped once per route, without buffering rejected credentials) and header/body/idle deadlines without
short global write deadlines that break SSE. Outbound assistant/OAuth calls use bounded
clients, request contexts and limited reads. Legacy metrics share one scan/device and
existing collectors, with context/scan guard budgets; preserve aggregate/filter semantics.

MCP keeps HTTP responses and targets legacy SSE replies using a unique advertised
session endpoint instead of global fan-out. Shared resource notifications are distinct.
See [trust ADR](adr-trust-and-streams.md). The user deferred MCP/context authentication
migration; public access stays unchanged. Dockerfile/Compose/Nginx establish the
host.docker.internal:4110 path. Keep that routing; add TRUSTED_PROXY_CIDRS (comma-separated
IPs/CIDRs, no implicit trust). Parse peers and headers as IPs, reject malformed trust
configuration, and refuse forwarded local sign-in from unlisted peers. A trusted proxy
must supply a valid client address: one X-Real-IP, otherwise X-Forwarded-For traversed
right-to-left through trusted hops. Bound the fallback to 32 hops/4096 bytes. Direct
local callers without forwarding metadata retain local login. Nginx overwrites incoming
client-IP headers. Configuration is read on the existing sign-in decision path with no
DNS, caches or background workers. Deployment-specific allowlist values are explicitly
configured before upgrade, never guessed from private address ranges.

Frontend permit-join work is owned by the application shell with explicit disposal;
consumers read the same confirmed bridge store and shared countdown. Assistant service
calls accept AbortSignal; composable owns controllers, request generations, session
revision checks and scroll timer cleanup. Vuex owns sensitive reset. Activity uses its
existing store subscription and clears on hub/session reset. No new state library.

Patch compression narrowly (lockfile resolution/override only if required by serve).
Do not bundle unrelated major upgrades. New tests go into exact source companions;
existing internal/scenario filename debt remains outside this behavior fix.

## Constitution Check

Before implementation PASS below refers to independent design, not executed checks.
Q1 is resolved by explicitly deferring public-route authentication migration. Q2 is
resolved from repository Docker configuration at the user's direction. Proxy design
and isolated regression evidence pass NH-03/04/05 and BE-05. Actual Nginx container
validation and deployment checks are unrun, as recorded below; no live rollout is claimed.

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | Spec with stable ACs, branch scope and resolved scope/topology | PASS | PASS scoped implementation; evidence below |
| NH-02 | No automation/reporting/command semantics changes; existing golden/E2E checks | PASS | PASS independent scope; evidence below |
| NH-03 | Explicit trusted peers; fail-closed headers; MCP auth deferred | PASS proxy design | PASS code/regressions; deployment validation unrun; no approval claimed |
| NH-04 | Existing repositories/services/Vuex; trust ADR and contracts below | PASS | PASS independent scope; evidence below |
| NH-05 | Companion regressions and full isolated suites | PASS design | PASS independent scope; evidence below |
| NH-06 | Resource limits/cleanup, owned bytes, unchanged disk schema | PASS design | PASS independent scope; evidence below |
| BE-01 | HTTP/WS adapt; domain/storage behavior stays testable with fakes | PASS | PASS independent scope; evidence below |
| BE-02 | Lifetime/event processor unchanged; golden/repeated-event checks retained | PASS | PASS independent scope; evidence below |
| BE-03 | Cancellation, locks, expiry, streams; affected/full race suite | PASS design | PASS independent scope; evidence below |
| BE-04 | Legacy fixtures, inclusive UTC ranges, no concealed prune failures | PASS design | PASS independent scope; evidence below |
| BE-05 | Mode/local-start share existing decider; targeted SSE; MCP auth deferred | PASS proxy design | PASS isolated mode/session tests; container/deployment checks unrun |
| FE-01 | Existing Vuex modules, owned composables | PASS | PASS independent scope; evidence below |
| FE-02 | Existing contracts; optional signals; two-client/ingress checks | PASS | PASS independent scope; evidence below |
| FE-03 | Confirmed state remains authoritative; no automatic command retry | PASS | PASS independent scope; evidence below |
| FE-04 | No visual/control change; existing keyboard/viewport E2E suite | PASS design | PASS independent scope; evidence below |
| FE-05 | Sensitive reset, late-response guards, no raw HTML or secrets | PASS design | PASS independent scope; evidence below |

## Contracts, data, rollout

- HTTP payloads unchanged; 413 oversized body, existing controlled gateway/query errors.
  Implemented shared ingress cap 1 MiB; upstream assistant reply cap 4 MiB; assistant
  deadline 60s and OAuth 15s. Header 5s/body 15s/idle 60s; streaming write deadline is
  refreshed per event, not a fixed stream lifetime. Document final values with evidence.
- WS unchanged event envelopes/1 MiB ingress; slow/expired clients close and recover
  through existing frontend reconnect. No new command retry or confirmation behavior.
- MCP direct POST returns JSON as before. Legacy SSE endpoint includes an opaque
  sessionId query parameter; clients post to the advertised endpoint. HTTP-only clients
  need no session parameter. Notifications never contain other callers' query results.
- No disk/key/metric-unit/retention migration. Rollback is the prior binary/assets;
  auth deployment settings must roll back together. No destructive data operations.
- Direct HTTP local operation retained. Public-route authentication migration deferred
  at the user's request. Set TRUSTED_PROXY_CIDRS in backend/.env to the observed Nginx
  peer before upgrading; only trust dedicated proxy addresses. Missing/invalid settings
  cannot authorize forwarded local sign-in. Keep existing ports/Docker network routing.
  Verify local mode/session and external Google mode after coordinated upgrade;
  rollback restores prior image/config together, with restricted backend access while
  the old forwarded-header vulnerability exists. No live deployment is performed here.

## Traceability and verification

| IDs | Owners / regression evidence | Actual result |
| --- | --- | --- |
| FR-01 / AC-01–02 | utils/storage/key-value.go → key-value_test.go; legacy fixture/concurrency | PASS: failed before fix; focused/full race and legacy reopen pass |
| FR-01 / AC-03 | websocket.go/websocket_test.go; useHubConnection.ts companion | PASS: small allocation, overflow, close, reconnect/latest-report/timeout tests |
| FR-02 / AC-04–05 | internal/api middleware/routes/server; auth/auth; companion tests | PASS: ingress boundary, startup error, cancellation, oversize reply, failed writes; OAuth bounds inspected, no live provider test |
| FR-02 / AC-06 | query_service.go/query_service_test.go; shared history constants | PASS: one scan, repeated exposes, cancellation/budget, existing aggregate/filter/history suites |
| FR-03 / AC-07–11 | utils/network.go/network_test.go, auth, ws, mcp/transport/http companions | AC-07 PASS direct/trusted/spoofed/malformed/chain/mapped-IP regressions and mode/session checks; AC-08/09 PASS including real SDK; AC-11 code PASS direct TLS/explicit flag; AC-10 migration deferred by user; actual proxy rollout unrun |
| FR-04 / AC-12 | usePermitJoin, AppShell; companion + StatusBanners E2E | PASS: owned timer, disposal/remount; desktop/mobile panel countdown |
| FR-04 / AC-13–14 | useAssistant/service, store/index, useActivityFeed; exact companions/contracts | PASS: cancellation, latest history, wrong-chat prevention, logout, disposal, malformed responses |
| FR-05 / AC-15 | scoped serve/compression override and lockfile | PASS: locked install, compression absent from audit, gzip asset and SPA fallback smoke |
| NFR-01 / SC-01 | full backend tests/race/vet/build; frontend Jest/lint/build/E2E | PASS independent scope; exact commands/results below |

Baseline audit checks passed on the equivalent pre-main commit; current branch checks
will be recorded separately. Use Node 24 and module-compatible Go. No household targets.

## Debt, risks, exceptions

No exceptions claimed. Existing public routes are documented debt, not retroactive
proof of a constitutional violation. Historical internal/scenario test names have seven
ownership mismatches; avoid expanding them. Streaming clients ignoring the advertised
endpoint need migration evidence; do not silently route replies to all subscribers.
Hard resource limits reject formerly unbounded requests; choose/document limits and
verify ordinary requests and fixture workloads. No claim of zero possible future bugs.

## Completion

Scoped implementation/check evidence is recorded below. MCP authentication migration
is deferred. Actual container/deployment verification and maintainer review remain
unperformed; configure trusted peers before upgrading. No deployment is claimed.

### Discovered verification repair

The first exact-Go-1.24 full race run failed only the unchanged wall-clock
TestAddTimeSchedule (execution-order assertion); three isolated race reruns passed.
Replace its sleeps/real timers with the existing MockClock and the same alternating
schedule assertions. This is a test-only reliability repair under NH-05/BE-03;
production scheduler semantics remain unchanged.

## Verification record (2026-10-21)

- Branch fix/runtime-security-and-resource-lifecycle from clean, pulled main 682ebf07.
- Storage regressions before production fix: GetOwnsBytes failed (borrowed bytes);
  PruneWithConcurrentWrites failed after bounded wait. After fix, full storage race
  suite passes, including missing/empty values, rollback and legacy read/write/reopen.
  Final API review also verifies authorization is rejected before reading payloads.
- Backend: `go test -count=1 -timeout=5m ./...` PASS (initial Go 1.26.1 and final
  exact Go 1.24.8), and
  `GOTOOLCHAIN=go1.24.8 go test -race -count=1 -timeout=5m ./...` PASS after the
  documented test-only scheduler repair. A final focused auth/MCP race plus exact
  Go 1.24.8 `go vet ./...` / `go build ./...` PASS. No races reported.
- Scheduler repair: exact-Go-1.24 race run exposed the pre-existing wall-clock ordering
  test; three real-clock isolated reruns passed, then ten controlled-clock runs passed.
  Production scheduler unchanged. Initial listen-failure and HTTP cancellation fixture
  errors were corrected (invalid port; consumed request body and bounded teardown).
- Frontend Node 24.21.0: clean `npm ci --ignore-scripts --no-audit --no-fund` PASS;
  `npm test -- --runInBand` PASS: 48 suites / 302 tests. `npm run lint` PASS with
  seven existing script warnings; `npm run build` PASS with the existing DeviceCard
  auto-import naming warning.
- `CI=true npm run test:e2e` PASS: 175 cases / nine existing desktop skips, 4.0m,
  fresh mock servers only. Includes desktop/mobile command confirmations, repeated
  interactions/leak cycles, auth/logout, keyboard/layout, history, reconnect and new
  permit-join panel-navigation case. Earlier full run passed 173 cases before the
  new two-viewport case; its isolated run also passed.
- `go test -run '^$' -bench '^BenchmarkWebSocketConnection$' -benchtime=20x -benchmem
  ./internal/ws` under Go 1.24.8: 115071 ns/op, 57378 B/op, 171 allocs/op locally.
  This is setup allocation evidence, not a production throughput guarantee. Heap
  regression also enforces <4 MiB/idle connection (audit baseline approximately 42 MiB).
- MCP actual SDK SSE client: initialize/resource-list round trip PASS using the new
  advertised session URL; two clients cannot receive each other's replies; HTTP-only
  calls remain supported. Stream/listener/auth cleanup tested with race detection.
- Temporary npm-start smoke: patched compression 1.8.2, gzip of a 491062-byte built
  asset exactly matches its original bytes, SPA route returns 200; own server stopped.
  Initial temporary harness errors were corrected before the passing smoke.
- `npm audit --json` completes with exit 1: 42 package entries remain (34 high,
  eight moderate), down from 44; compression/serve no longer flagged. Remaining
  packages include development-tool chains and the SSR renderer advisory; no blanket
  exploitability/clean-audit claim. Broad upgrades remain excluded for compatibility.
- No checks contacted household devices. No deployment, real Google account flow,
  Docker/Linux image run or independent maintainer review was performed. API provider
  timeout/TLS-terminator policy is covered by code/isolated tests, not a live rollout.
- User follow-up: defer public MCP/context authentication migration and revisit later;
  deployment uses the shipped Docker/Nginx setup, confirmed by the user's direction
  to inspect Dockerfile/Compose. Public access stays unchanged. Proxy hardening uses
  explicit configured peers with a coordinated upgrade; T015 evidence follows.

### Proxy follow-up verification

- Before the fix, network_test.go reproduced spoofed X-Real-IP/X-Forwarded-For local
  authorization, first-hop spoofing, missing-client fallback and ambiguous headers.
  After the fix, focused `GOTOOLCHAIN=go1.24.8 go test -race -count=1 -timeout=2m
  ./utils ./internal/auth` PASS, including configured Nginx local-session issuance,
  external Google mode, direct-LAN compatibility, IPv6/mapped addresses, CIDR boundaries,
  malformed trust lists, duplicate headers, chain limits and no-cookie denials.
- Full `GOTOOLCHAIN=go1.24.8 go test -count=1 -timeout=5m ./...` PASS; full
  `GOTOOLCHAIN=go1.24.8 go test -race -count=1 -timeout=5m ./...` PASS on the final
  proxy implementation, no races reported. `go vet ./...` and `go build ./...` under
  the same toolchain PASS. Sandbox attempts first blocked build-cache access and an
  IPv6 loopback test listener; authorized reruns outside the sandbox passed.
- Reviewed both Nginx examples: existing routes/ports retained; X-Real-IP remains
  overwritten and X-Forwarded-For now also overwrites incoming claims. Docker CLI
  reports no running daemon/socket, and no native Nginx executable is installed.
  Nginx syntax/container integration and actual LAN/external sign-in verification
  are UNRUN; isolated tests do not substitute for those deployment checks.
- Frontend source did not change in this follow-up, so previously passing 302 unit
  tests, lint/build and 175 browser cases were not repeated for Nginx/doc changes.
  `git diff --check` PASS. Tests remain isolated; no household services were contacted.
- Rollout requires the operator's explicit TRUSTED_PROXY_CIDRS in backend/.env before
  upgrade. No values were guessed or deployed. See the linked runtime guide for
  peer identification, address changes, checks and coordinated rollback.

### Environment documentation follow-up

T016 is documentation-only: add commented proxy/cookie examples to the checked-in
backend env files and distinguish direct HTTP, Nginx HTTP and terminated HTTPS in the
setup guide. Link to the existing runtime policy rather than duplicate trust rules.
Preserve current effective assignments and URLs; no production/runtime code change.
Verification PASS: active assignments in both backend env files are identical to HEAD;
both files include commented examples for TRUSTED_PROXY_CIDRS/AUTH_COOKIE_SECURE. Local
file links across six affected setup/README/spec documents and the new target headings
resolve; `git diff --check` PASS. Application suites were not repeated for this
comments/documentation-only follow-up. No values, secrets or deployment settings changed.

Final operational contracts: [runtime behavior](../../docs/runtime-reliability.md).
Temporary evidence logs are in /private/tmp/nodeherder-024-*.log; the assertions and
acceptance outcomes are version-controlled in source companions and this record.
