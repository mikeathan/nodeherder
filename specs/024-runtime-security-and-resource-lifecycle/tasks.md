# Tasks: Runtime security and resource lifecycle

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: implementation verified; container/deployment validation unrun

- [x] T001 — Pull clean main; create fix/runtime-security-and-resource-lifecycle;
  read constitutions/rules/SDD skill and create this one artifact set. Gate NH-01.
- [x] T002 — Q1 resolved: user deferred public MCP/context authentication migration.
  Q2 resolved from Dockerfile/Compose/Nginx at user's direction; keep host-gateway routing
  and explicitly configure deployment-specific trusted peers before upgrade;
  IDs FR-03/AC-07/10/11; evidence: user's actual answers and documented route/proxy/cookie rollout.
- [x] T003 — Storage ownership/prune concurrency regressions in existing key-value_test.go,
  then fix key-value.go. FR-01/AC-01–02; depends T001; evidence: focused race/legacy tests.
- [x] T004 — WS queue/allocation/disconnect/overflow regressions in websocket_test.go;
  implement websocket.go and bounded reconnect snapshot reconciliation in
  useHubConnection.ts and its companion test. FR-01/AC-03; depends T001; evidence: measured bound and tests.
- [x] T005 — Shared stream auth lifetime/origin/blacklist regressions and fixes in auth/ws
  exact companion tests. FR-03/AC-08; depends T001; done: focused/full race and origin policy tests.
- [x] T006 — MCP two-client routing, disconnect/write cleanup tests and transport fixes;
  FR-03/AC-09; depends T001; done: two-client isolation and actual SDK compatibility.
  Public-route authentication migration is tracked separately in T002.
- [x] T007 — Bounded ingress/outbound HTTP regression/fixes in api/auth companions;
  FR-02/AC-04–05 and cookie helper AC-11; done: cancellation, body/reply limits,
  direct-TLS/explicit secure-cookie tests. Proxy configuration is tracked in T002.
- [x] T008 — Single-scan metrics cancellation/budget/compatibility regressions and service
  changes; FR-02/AC-06; depends T001; evidence: existing/new query/aggregation tests.
- [x] T009 — Permit-join ownership/remount tests and shell/composable fix; FR-04/AC-12;
  depends T001; evidence: fake-timer scope tests and panel E2E.
- [x] T010 — Assistant request/switch/logout/unmount regressions; signal/generation fixes,
  centralized sensitive reset/activity cleanup. FR-04/AC-13–14; depends T001.
- [x] T011 — Patch compression lock resolution; FR-05/AC-15; depends T001;
  evidence: locked install, audit-specific resolution, Node serving smoke/build.
- [x] T012 — Update deployment/contracts docs and final rule checks; all FRs; depends
  T002–T011; evidence: no placeholders, traceability and actual check results in plan.
- [x] T013 — Run full backend test/race/vet/build and frontend unit/lint/build/E2E;
  gate NH-05/NFR-01; depends implementations; record failures/unrun checks accurately.

- [x] T014 — Make existing TestAddTimeSchedule in scheduler_test.go use mocks.MockClock;
  gate NH-05/BE-03; discovered by Go 1.24 race failure, isolated three reruns passed.
  No production scheduler edits; evidence: deterministic repeated test and full race rerun.

- [x] T015 — Add network.go-owned network_test.go trust regressions and auth_test.go
  end-to-end mode/session regressions, implement explicit proxy trust, overwrite
  incoming forwarded headers in Nginx, document coordinated upgrade/recovery;
  FR-03/AC-07/11, NH-03/04/05 and BE-05. Focused/full race, full Go tests/vet/build
  PASS. Nginx examples reviewed; actual container syntax/integration UNRUN because
  the Docker daemon is unavailable. Deployment verification is documented, not performed.

- [x] T016 — Make local HTTP versus Nginx HTTP/HTTPS configuration discoverable in
  checked-in backend env comments and setup samples; AC-07/11, NH-01 lightweight path.
  Preserve all active env assignments; unchanged assignments, local links/target headings
  and diff hygiene PASS. No application code changes or repeated runtime suites.

Independent tasks are checked against the verification record in plan.md.
Public MCP/context authentication migration is deferred at the user's request.
Scoped implementation/docs checks are complete; actual container/deployment verification
remains unrun, with cause and consequences in plan.md. Configure proxy trust before upgrade.
No deployment or maintainer approval is implied.
