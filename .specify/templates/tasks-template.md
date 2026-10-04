# Tasks: <name>

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: <draft/in progress/complete>

Each task MUST map to requirement/acceptance IDs or a constitutional gate and include
paths, dependencies, and completion evidence. Adapt phases; mark done only with evidence.

## Preparation

- [ ] T001 — Resolve <question> in `spec.md`; IDs: <IDs>; done: <evidence>.
- [ ] T002 — Check applicable rules in `plan.md`; depends: T001; done: <evidence>.

## Coverage and implementation

- [ ] T003 — Regression/acceptance test in <path>; IDs: <IDs>; depends: T002;
  reproduce failure where feasible; done: <evidence>.
- [ ] T004 — Implement <behavior> in <paths>; IDs: <IDs>; depends: T003;
  done: <passing test/procedure>.
- [ ] T005 — Update contracts/producers/consumers in <paths>; IDs: <IDs>;
  depends: <task>; done: <compatibility evidence>.

## Handoff

- [ ] T006 — Planned checks; gate: NH-05; done: recorded actual results.
- [ ] T007 — Failure/recovery/UI checks; IDs: <IDs>; done: <evidence>.
- [ ] T008 — Docs/final rule checks; gates: NH-01/04/05; done: acceptance evidence
  and any required exception approval.

Add discovered work; update spec/plan for scope changes.
