# Tasks: Integration refresh scheduler

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: implemented; automated checks passed; browser check unrun

Each task maps to requirement/acceptance IDs or a constitutional gate and includes paths,
dependencies and completion evidence. Do not mark complete without evidence.

## Preparation

- [ ] T001 — Resolve product assumptions A-01/A-02 (new-instance default, legacy behaviour) and A-03
  (retention split) in `spec.md` with the maintainers; IDs: FR-09, AC-06, AC-18; done: recorded decision
  and any changed acceptance text. These do not block model/coordinator design. The implementation
  follows the documented assumptions; maintainer confirmation remains outstanding.
- [x] T002 — Confirm the schedule contract and provider policy grammar in
  `contracts/schedule.md`; IDs: FR-01–FR-03, FR-11; done: reviewed contract with no unresolved
  auth/contract question.
- [x] T003 — Recheck the plan's Constitution Check against the final spec; IDs: NH/BE/FE gates; done:
  recorded design review, no unapproved violation.

## Coverage and implementation

- [x] T004 — Add compiler tests for `refresh_default_seconds`/`refresh_min_seconds` defaults, missing
  fields, boundary values, default-below-minimum and Form/YAML parity in
  `backend/internal/integrations/definitions/compiler_test.go`; IDs: FR-02, AC-17, AC-18; depends:
  T002.
- [x] T005 — Add schedule/DTO serialization tests, including nil legacy schedules and additive field
  compatibility, in `backend/internal/integrations/runtime/` companion tests; IDs: FR-01–FR-03,
  FR-12, AC-06, AC-20. The model record is exercised through the runtime schedule tests.
- [x] T006 — Add controlled-clock scheduler tests in
  `backend/internal/integrations/runtime/scheduler_test.go`: not-due, due, interval boundaries,
  disabled/manual, per-instance coalescing, manual reset, exactly-one admission, jitter range,
  scheduled 304 content-age behavior, 429/5xx backoff, existing-operation schedule backoff, cadence
  edit preserving a retry deadline, instance/schedule re-enable preserving a retry deadline,
  interrupted refresh using its operation retry deadline instead of regular cadence, blocked
  definition failure persistence and repair clearing, definition repair preserving rejected-credential
  failures while clearing destination-review failures, retry_required resume, pending renewal before
  expiry, restart, queue-capacity admission, shutdown/no re-arm. The existing four-worker pool and
  goroutine count are not separately asserted by scheduler tests.
  IDs: FR-04–FR-09, FR-14, FR-16–FR-19, NFR-01–NFR-04, AC-07–AC-15, AC-22, AC-24–AC-27;
  depends: T003.
- [x] T007 — Add schedule API and frontend ingress tests in
  `backend/internal/api/integrations_test.go` and
  `frontend/src/__tests__/contracts/integrations.test.ts`; IDs: FR-10, FR-03, AC-19, AC-01; depends:
  T002.

## Implementation

- [x] T008 — Add `RefreshSchedule`, `Instance.Schedule`, `Operation.Trigger` and
  `RunSummary.Trigger` in `backend/models/integrations/state.go`; IDs: FR-01, FR-03, FR-06; depends:
  T005.
- [x] T009 — Add the policy fields and strict defaults/bounds in
  `backend/models/integrations/definition.go` and `backend/internal/integrations/definitions/compiler.go`;
  update `backend/configs/integrations/weather/*.yaml` and
  `backend/configs/integrations/waste_collection/whenisbins.yaml`; IDs: FR-02, FR-11; depends: T004.
- [x] T010 — Extract shared refresh admission in `backend/internal/integrations/runtime/controls.go`
  (`admitRefresh`) and continuation admission from `CheckOperation`, update manual `Refresh` to use
  them and reset the next due time; add `UpdateSchedule`; set the future due time on enable in
  `runtime/controls.go` and create the initial schedule in `runtime/setup_selection.go`;
  IDs: FR-01, FR-05, FR-08, FR-09, FR-12; depends: T008.
- [x] T011 — Add the coordinator and outcome arithmetic in
  `backend/internal/integrations/runtime/scheduler.go`; wire the timer and `Close` in
  `runtime/service.go`; make queue insertion context-aware so shutdown cannot wedge on a full queue;
  update terminal paths in `runtime/publication_result.go`, `runtime/resources.go` and
  `runtime/operation_state.go`; include the user-action error skip list;
  IDs: FR-04, FR-06, FR-07, NFR-01–NFR-04, AC-07–AC-16; depends: T006, T008, T010.
- [x] T012 — Add `ScheduleDTO`/`InstanceDTO.Schedule` in `runtime/dto.go`; IDs: FR-03, NFR-06.
- [x] T013 — Add the authenticated schedule route/input/action in
  `backend/internal/api/integrations.go`; IDs: FR-10, AC-19; depends: T010, T012.
- [x] T014 — Add frontend types, parser, service, pure schedule helpers and detail composable:
  `frontend/src/types/integrations.type.ts`, `frontend/src/contracts/integrations.ts`,
  `frontend/src/services/integrations.service.ts`, `frontend/src/domain/schedule.ts`,
  `frontend/src/composables/integrations/useConnectionDetail.ts`; IDs: FR-02, FR-03, FR-10, AC-01,
  AC-17; depends: T012, T013.
- [x] T015 — Add the settings UI and provider policy fields in
  `frontend/src/components/integrations/ConnectionSettings.vue`,
  `ConnectionDetail.vue`, `ProviderPolicyFields.vue` and CSS; IDs: FR-03, FE-04, NFR-06, AC-01,
  AC-17; depends: T014.
- [x] T016 — Add frontend unit tests for duration conversion/formatting, exact-second intervals,
  pristine-editor reconciliation on initial and subsequent reads, draft base-revision rebasing and
  schedule-change conflicts, and schedule parsing in
  `frontend/src/__tests__/domain/schedule.test.ts`,
  `frontend/src/__tests__/composables/integrations/useConnectionDetail.test.ts` and
  `frontend/src/__tests__/contracts/integrations.test.ts`, and update affected provider/instance
  fixtures for the additive fields; IDs: FR-02, FR-11, FR-13, FR-15, FR-20, AC-04, AC-17, AC-21,
  AC-23, AC-28; depends: T014.

## Handoff

- [x] T017 — Record the schedule contract, defaults, rollback and provider-quota behavior in
  `docs/connections.md`; keep the retention warning and 034 follow-up explicit; IDs: NH-01, NH-04,
  AC-20; done: docs review.
- [x] T018 — Run backend checks (`go test ./...`, `go build ./...`, `go vet ./...`,
  affected `go test -race`) and frontend checks (`npm test -- --runInBand`, `npm run lint`,
  `npm run build`); gates: NH-05, BE-03, FE-02; done: exact commands/results recorded in `plan.md`.
- [ ] T019 — Verify the settings flow in an isolated browser at desktop and mobile widths: manual-only,
  interval save, next/last display, disabled instance, failure text, keyboard/focus; IDs: FE-01,
  FE-03, FE-04, AC-01, AC-05, AC-11; done: recorded procedure and result, no household devices.
  Not run; automated build/tests passed.
- [x] T020 — Review the final diff against the Constitution Check, update any scope/decision changes in
  this spec/plan, and record unresolved risks; gates: NH-01/NH-04/NH-05/NH-06; done: review notes and
  truthful pass/fail/unrun summary.
