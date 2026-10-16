# Plan: Presence recovery and truthful connection notices

Spec: [spec.md](spec.md) | Status: Implemented | Updated: 2026-10-16
Constitutions: parent/backend/frontend 1.0.0 (proposed). No adoption or human approval claimed.

## Design and compatibility

Backend: `Condition.DependsOn` declares inputs; `Trigger.AffectedBy` selects primary updates
or declared condition dependencies for state sources. `Device` orchestrates; `DeviceTrigger.Process`
owns condition checks, cancellation and action execution. No eligibility cache, copied dependency
list, concrete condition casts or duplicate predicate evaluation. Custom conditions use the same
contract. Time conditions have no device dependencies. Cached physical events cannot replay on
condition changes. Existing HasValueChanged/action behavior remains, including repeated ON on
changed eligible readings; identical reports stay deduplicated by the lifetime service.

Frontend: Vuex socket state distinguishes startup, connected, reconnecting and disconnected.
The socket module owns one connection, a 10-second handshake timeout, at most ten automatic
retries with 1–30-second capped backoff, stale-callback guards and disposal. The composable
coalesces session/snapshot work and aborts stale HTTP requests; snapshot failures belong to the
hub module. Notices are dismissible fixed overlays; the header remains accessible and content
does not move. PanelMode owns its sole notice host.

No dependencies, wire schemas, auth policy, metrics sampling, persistence, recipe threshold or
five-minute OFF delay change. Internal Trigger/Condition interfaces gain selection methods;
base embeddings supply defaults. Backend/frontend deploy or roll back independently. Reconnect
never replays device commands. Avoid blanket deduplication removal, arbitrary retries, and the
rejected once-per-eligibility policy; those would change behavior beyond the demonstrated bug.

## Constitution check

Design checks preceded implementation; final agent review rechecked the following against the
diff and results below. No exceptions requested or claimed. Existing rules remain authoritative.

| Rule | Evidence | Design / final |
| --- | --- | --- |
| NH-01 | Spec, plan and tasks preceded edits and the user-directed revision | PASS / PASS |
| NH-02 | Real-recipe regressions, unchanged commands/timing, state/event isolation | PASS / PASS |
| NH-03 | Existing auth helpers; no new entry point, secret or data persistence | PASS / PASS |
| NH-04 | Existing domain/Vuex boundaries and explicit internal dependency contracts | PASS / PASS |
| NH-05 | Before/after regressions, suites, races and scoped quality evidence below | PASS / PASS |
| NH-06 | Bounded socket retries and request/timer cleanup; no storage changes | PASS / PASS |
| BE-01 | Domain tests isolate MQTT/storage with established fakes | PASS / PASS |
| BE-02 | Fresh-state, deduplication, sampling and repeated-event tests retained | PASS / PASS |
| BE-03 | No new backend shared state; manual isolation and races pass | PASS / PASS |
| BE-04/05 | No storage/query/routes/assistant changes | N/A |
| FE-01 | Vuex owns transport/data truth; components own presentation | PASS / PASS |
| FE-02 | Wire routing preserved; full producer/consumer suites pass | PASS / PASS |
| FE-03 | Startup, outage, snapshot failure and recovery are distinct | PASS / PASS |
| FE-04 | Desktop/mobile/panel geometry, keyboard retry/dismissal and focus verified | PASS / PASS |
| FE-05 | Established authenticated helpers; no new persistence/secrets | PASS / PASS |

## Acceptance and actual verification

Behavior checks completed 2026-10-15; frontend unit/lint/build checks repeated successfully
after type-ownership cleanup on 2026-10-16. Backend checks followed the final stateless revision.
Automated tests used fake MQTT or localhost mock servers only.

| Acceptance | Source-owned coverage | Result |
| --- | --- | --- |
| AC-01…04 / FR-01 | device_test.go: real recipe/lifetime split reports, threshold/zero, identical reports, OFF delay/cancellation, events, manual runs and action failures; trigger_test.go/condition_test.go: generic dependencies, one evaluation, unrelated updates, combined changes, time/missing source | PASS |
| AC-05…09 / FR-02 | ws/index.test.ts, useHubConnection.test.ts, hubstate.service.test.ts, shell/connection.test.ts; StatusBanners.spec.ts plus existing auth/layout/panel flows | PASS |
| NFR-01 | Per-callable quality below, including conservative bounds where coverage is unknown | PASS |
| SC-01 | Required component checks below and before-fix failures | PASS |

- Backend: full `go test -count=1 -timeout=5m ./...` with coverage on Go 1.26.1;
  full suite again with `GOTOOLCHAIN=go1.24.8`; `go vet ./...`; `go build ./...`: PASS.
- Races: `go test -race -count=1 -timeout=5m ./internal/automations/... ./internal/services ./internal/controllers`;
  focused presence/dependency/selection/manual/failure regressions repeated five times: PASS.
- Frontend (Node 24.21.0): `npm test -- --runInBand` with scoped coverage: 27 suites/184 tests PASS;
  `npm run lint`: zero errors/seven existing script warnings; `npm run build`: PASS with existing
  duplicate DeviceCard-name/large-chunk warnings.
- `npm run test:e2e`: 58 desktop/mobile checks PASS (1280×800/390×844); final strengthened
  `e2e/StatusBanners.spec.ts`: 12 PASS. Healthy startup observers catch transient warning DOM;
  outage/snapshot/recovery geometry stays fixed. Keyboard Enter retries/dismisses, focus returns
  to the header, and a later outage reopens the notice. Generated screenshots were inspected.
- Before-fix presence, socket and notice regressions failed as expected. The rejected cache also
  failed new tests for changed eligible readings, evaluation after failure and double evaluation.
  One intermediate browser test queried hidden mobile navigation incorrectly; its query was
  corrected and full/final runs passed. Initial sandbox cache/port failures passed after rerun
  with the required environment permission. No failing check was waived.

## Quality evidence

Scorer: existing `crap` v0.0.0-20260908235400-8eea294df112; Go coverprofile and Jest/Istanbul
Cobertura mapping, selected with `-diff-base=fa02e1da` from each component directory.
CRAP=C²(1−coverage fraction)³+C. Baseline ensureSocket: C=6/0%/CRAP=42; old socket.onmessage:
C=17/58.97%/CRAP=36.96. Final changed native callables have C≤6/CRAP≤6. No surrounding measured
Go/TS file has CRAP>30. Unchanged hub updateDevice has C=11/CRAP=11.07, outside changed callables.

Vue scripts lack Jest instrumentation. A temporary TypeScript 5.9.2 AST analysis using spec 009's
callable convention excluded nested callables and counted control-flow/conditional/logical
branches. Maximum C=3 gives a zero-coverage CRAP upper bound of 12; actual Vue coverage/scores
are unknown. Browser checks provide separate behavior evidence. No one-off analyzer is retained.
In the table, Unknown means the conservative zero-coverage upper bound, not measured coverage.
Source paths are relative to backend/ or frontend/ respectively.

| Source | Callable | C | Coverage % | CRAP |
| --- | --- | ---: | ---: | ---: |
| internal/automations/condition.go | automations.(*BaseCondition).DependsOn | 1 | 100 | 1 |
| internal/automations/condition.go | automations.(*ExposeCondition).DependsOn | 1 | 100 | 1 |
| internal/automations/device.go | automations.(*Device).Evaluate | 5 | 91.67 | 5.01 |
| internal/automations/trigger.go | automations.(*BaseTrigger).AffectedBy | 4 | 100 | 4 |
| internal/automations/trigger.go | automations.(*BaseTrigger).conditionsAffectedBy | 4 | 100 | 4 |
| internal/automations/trigger.go | automations.(*DeviceTrigger).Process | 2 | 100 | 2 |
| internal/automations/trigger.go | automations.(*DeviceTrigger).conditionsPermitExecution | 6 | 100 | 6 |
| internal/automations/trigger.go | automations.(*DeviceTrigger).stopActions | 2 | 100 | 2 |
| internal/automations/trigger.go | automations.(*DeviceTrigger).executeActions | 3 | 100 | 3 |
| src/components/shell/connection.ts | connectionText | 1 | 100 | 1 |
| src/components/shell/connection.ts | connectionNotices | 6 | 100 | 6 |
| src/composables/useHubConnection.ts | requestSnapshot | 6 | 100 | 6 |
| src/composables/useHubConnection.ts | loadHubState | 3 | 100 | 3 |
| src/composables/useHubConnection.ts | <anonymous@28:82> | 2 | 100 | 2 |
| src/composables/useHubConnection.ts | validateAndConnect | 4 | 100 | 4 |
| src/composables/useHubConnection.ts | startHubConnection | 2 | 100 | 2 |
| src/composables/useHubConnection.ts | <anonymous@51:56> | 2 | 100 | 2 |
| src/composables/useHubConnection.ts | stopHubConnection | 2 | 100 | 2 |
| src/composables/useHubConnection.ts | useHubConnection | 1 | 100 | 1 |
| src/composables/useHubConnection.ts | <anonymous@78:19> | 1 | 100 | 1 |
| src/services/hubstate.service.ts | fetchHubState | 2 | 100 | 2 |
| src/store/modules/hub-state/index.ts | state | 1 | 100 | 1 |
| src/store/modules/hub-state/index.ts | loadStatus | 1 | 100 | 1 |
| src/store/modules/hub-state/index.ts | clear | 1 | 100 | 1 |
| src/store/modules/hub-state/index.ts | setLoadStatus | 1 | 100 | 1 |
| src/store/modules/ws/index.ts | resources | 2 | 100 | 2 |
| src/store/modules/ws/index.ts | clearTimers | 3 | 100 | 3 |
| src/store/modules/ws/index.ts | scheduleReconnect | 2 | 100 | 2 |
| src/store/modules/ws/index.ts | <anonymous@42:30> | 1 | 100 | 1 |
| src/store/modules/ws/index.ts | decodeMessage | 6 | 100 | 6 |
| src/store/modules/ws/index.ts | routeMessage | 5 | 100 | 5 |
| src/store/modules/ws/index.ts | bindSocket | 1 | Unknown | 2 |
| src/store/modules/ws/index.ts | isCurrent | 1 | 100 | 1 |
| src/store/modules/ws/index.ts | socket.onopen | 2 | 100 | 2 |
| src/store/modules/ws/index.ts | socket.onmessage | 3 | 100 | 3 |
| src/store/modules/ws/index.ts | socket.onclose | 2 | 100 | 2 |
| src/store/modules/ws/index.ts | socket.onerror | 1 | 100 | 1 |
| src/store/modules/ws/index.ts | <anonymous@109:51> | 2 | 100 | 2 |
| src/store/modules/ws/index.ts | connect | 6 | 100 | 6 |
| src/store/modules/ws/index.ts | disconnect | 3 | 100 | 3 |
| src/store/modules/ws/index.ts | state | 1 | 100 | 1 |
| src/store/modules/ws/index.ts | setSocket | 1 | 100 | 1 |
| src/store/modules/ws/index.ts | setConnectionStatus | 1 | 100 | 1 |
| src/store/modules/ws/index.ts | sendMessage | 2 | 100 | 2 |
| src/store/modules/ws/index.ts | emit | 1 | 100 | 1 |

## Review, limitations and completion

Final self-review checked dependency ownership, stateless selection, at-most-once predicate
evaluation, event/manual isolation, cancellation, errors, unchanged contracts and lifecycle
cleanup; no unresolved blocking findings. Original trigger tests were retained verbatim when
renaming triggers_test.go to trigger_test.go; the WS companion likewise now matches index.ts.
Existing HasValueChanged policy is preserved, not redesigned. No independent human approval
is claimed. Current deployed recipe/raw sensor sequence and physical hardware were not verified;
no household commands or deployment occurred. Permit-join rendering is unchanged but was not
separately exercised; no screen-reader/hardware-touchscreen run is claimed.

Temporary analyzers, replays, logs, coverage, screenshots and generated build/test outputs were
removed after recording results here. Permanent regression tests remain. Prepared on fix/presence-and-connection-status;
deployment and household verification remain outside this change.


Type-ownership follow-up (2026-10-16): extract shared notice, hub data/loading and session
validation shapes into src/types. Keep store-only interfaces in existing type-only state.ts
files; keep private socket implementation types local. Explicit type-only imports express these
boundaries. Unit tests (27 suites/184 tests), lint (zero errors/seven existing warnings)
and the TypeScript/production build pass. Connection status values now form a literal union.
No runtime or wire behavior change; backend/browser suites were not repeated for this
type-only refactor. Reviewed imports/type ownership and removed generated build output.
