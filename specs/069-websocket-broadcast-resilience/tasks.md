# Tasks: WebSocket broadcast resilience

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: complete

Each task maps to requirement/acceptance IDs or a constitutional gate.

## Preparation

- [x] T001 — Resolved Q-01 (deployed remote-logger state unconfirmed; fixes unconditional)
  and confirmed the last two merges contain no backend WS/event change; spec/plan/tasks
  created. IDs: NH-01, FR-01/FR-04.
- [x] T002 — Constitution check recorded in `plan.md`; depends T001. Gate: NH-01..06,
  BE-01..05.

## Coverage and implementation

- [x] T003 — Added WS regression tests that reproduce the deadlock and the flood with
  remote logging enabled; IDs: AC-01, AC-02, AC-03, AC-04; depends T002; done: hybrid
  copy with pre-fix fan-out + remote `HandleError` fails both
  `TestRemoteLoggingDoesNotWedgeBroadcastOnClosedSession` (wedge) and
  `TestSlowClientDoesNotBlockFanoutAndErrorsStayLocal` (remote session-error frames).
  `TestLogLocalDoesNotFireRemoteHook` proves local logs emit no frames.
  `TestBroadcastRejectsOversizedFrame`, `TestLargestObservationSnapshotFitsBroadcastBound`
  and `TestExecuteReportsRejectedResult` cover FR-03.
- [x] T004 — Replaced `BroadcastFilter` with bounded per-session fan-out, decoupled
  `maxBroadcastBytes` from inbound `maxMessageSize`, enforced the frame bound, counted
  failures in `BroadcastDiagnostics` exposed through `WebSocket.Diagnostics()`;
  `backend/internal/ws/websocket.go`; IDs: FR-01, FR-02, FR-03, AC-01..04; depends T003;
  done: T003 tests pass, including `-cpu=1`.
- [x] T005 — Added hook-free local logging to `backend/utils/logger.go` (second
  hook-free logger on a mutex-protected shared writer, levels kept in sync) and routed
  all melody callbacks, the WS request handler, `handleHubEvents`/`execute` and the
  observation publish/drop paths through it; IDs: FR-02, AC-03; depends T004; done:
  local-log test passes; independent reviewer found no remaining re-entrancy path.
- [x] T006 — Split observation change detection into structural and telemetry signatures
  with an injectable telemetry cadence; `backend/internal/observations/service.go`;
  IDs: FR-04, AC-05, AC-06; depends T002; done: new and existing observation tests pass,
  `renderSignature` verified byte-identical to pre-fix. Added
  `TestStructuralChangeRetriesAfterMissedPublishGap` and
  `TestStructuralSignatureSeparatesSourcesFromIdentities` after review.
- [x] T007 — Updated `WsHandler.ServeHTTP` misleading disconnect log to an accurate local
  line; `backend/internal/api/routes.go`; IDs: FR-02; depends T005.

## Handoff

- [x] T008 — Updated `backend/docs/architecture.md` with the fan-out invariant, the
  outbound bound, diagnostics, failure feedback and the telemetry cadence; IDs: NH-04;
  depends T004, T006; done: reviewed against the final diff; doc wording corrected to
  name source IDs (not full descriptors) and to cover the upgrade-error line.
- [x] T009 — Ran the full verification matrix and recorded actual results, failures and
  unrun checks in `plan.md` (Completion/Negative controls); gate: NH-05, SC-01, AC-07,
  AC-08; depends T003..T008. Reverted unrelated `frontend/components.d.ts` drift from
  `npm run build`.
- [x] T010 — Independent reviewer pass on the final diff; all findings fixed and
  re-verified (test negative controls, `-cpu=1` flake, case-sensitive log assertion,
  outbound bound, diagnostics API, `fanoutMu` invariant comment, uncovered retry branch,
  stale traceability, generated drift). Gate: NH-01..06, BE-01..05; depends T009.
- [x] T011 — Second independent review pass (2026-11-25; agent, not human approval):
  re-ran build/vet/full suite/race/repeat tests and both negative controls; corrected the
  `Session.Write` callback claim (a full queue invokes the synchronous, hook-free
  `errorHandler` on the caller) and the architecture outbound-bound wording; updated the
  spec status. No code finding required a change. Gate: NH-05, BE-03, AC-08; depends T010.
