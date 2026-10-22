# Tasks: Restore automation toggle commands

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: Complete

- [x] T001 — Inspect regression, read applicable rules/skill, create spec/plan/tasks
  before production edits; NH-01/04. Evidence: current branch starts at b60d2bdb;
  editor drops backend values.toggle; specification above.
- [x] T002 — Add failing companion domain regressions for native options, boundary
  metadata, strict readings, action summaries; FR-01/02, AC-02/04; depends T001.
- [x] T003 — Implement domain helpers and action-only editor opt-in; FR-01,
  NFR-01; depends T002. Done: passing focused Jest and typed source checks.
- [x] T004 — Reuse hubharness/production routes in opt-in synthetic browser hub;
  add real-engine UI-create/save/reload/two-event E2E in automations.spec.ts and
  managed Playwright/CI setup; AC-01/02/03, NFR-02; depends T003.
- [x] T005 — Run full frontend/backend checks, affected race checks and desktop/
  mobile browser suite; AC-04/05, NH-05; depends T004. Record exact outcomes.
- [x] T007 — Correct the zero-delay save contract in `contracts/automations.ts`
  and the shared Vuex save action, with companion tests; AC-06; discovered in T004;
  required before T005 completes.
- [x] T008 — Synchronize the callback registry in `backend/store/store.go` with
  copy-on-registration and unlocked notification; companion regression tests;
  AC-07, BE-03; discovered during T005; rerun full/race/backend and browser checks.
- [x] T006 — Review diff, fixture cleanup, failures, rules and compatibility;
  update plan evidence; report branch/checkout handoff in the final response; NH-01–06;
  depends T005. Do not claim unrun checks or live-device verification.
- [x] T009 — Preserve empty/replaced trigger lists atomically in
  `backend/internal/automations/trigger.go`; source-owned regression tests and
  UI empty-save/reload coverage; AC-08; rerun affected full/race/browser checks.
- [x] T010 — Apply review follow-ups: warm the Go fixture build and allow 300s
  startup in CI/Playwright; drain stale fixture queues and reset the light OFF;
  clarify round-trip comments and the zero-wait copy; update plan contract wording.
  AC-06, NH-05, BE-03, FE-04; depends T006/T009; done: component and browser checks
  recorded in the plan.
- [x] T011 — Make the Playwright config and browser tests type-check clean: add
  explicit node types, `tsconfig.e2e.json`, `type-check:e2e` in the frontend CI job,
  and an explicitly typed blocked-storage stub. NH-05, FE-04; depends T010; done:
  type-check/lint/build/Jest and the affected panel browser case pass.
- [x] T012 — Bump the frontend minor version from 0.12.0 to 0.13.0 in
  `frontend/package.json` and `package-lock.json`; releases derive from the package
  version via `scripts/tag_build.sh`. NH-01 mechanical; depends T011; done: lockfile
  in sync.
- [x] T013 — Make the TypeScript config TypeScript 6-compatible: explicit `rootDir`,
  remove deprecated `baseUrl`, relative `paths`, `moduleResolution: bundler`, a vuex
  types path override, and `rootDir: "."` for the e2e project. NH-05, FE-04; depends
  T012; done: build, Jest, `type-check:e2e`, and TS 6/next e2e checks pass.
- [x] T014 — Resolve actionable frontend audit findings: Vue 3.5.44, eslint-plugin-vue
  10 + vue-eslint-parser 10, unplugin-vue-components 32, Jest 30/ts-jest 29.4/
  @types/jest 30, and a qs 6.16.x override; record the residual dev-only sprintf-js
  chain. NH-04/05/06, FE-02; depends T013; done: production audit clean, full
  Jest/build/lint/type-check and affected browser suites pass.
- [x] T015 — Reduce the initial bundle: import only the used Aura component token
  modules instead of the full theme, and enable nginx gzip for text/proxied assets.
  NH-04/06, FE-04; depends T014; done: main entry 501.7 -> 438.8 kB (149.4 -> 141.5
  kB gzip), warning gone, full Playwright 179 passed.
- [x] T016 — Run the CI-equivalent checks locally (workflow lint, stack tools,
  backend/race, frontend, full browser suite, container/nginx smoke) on Node 24 and
  Go 1.24.8. NH-05; depends T015; done: all local checks passed; GitHub-hosted
  amd64/buildx specifics remain unrun.

Completion evidence: [plan verification](plan.md#traceability-and-verification).
T002/T008/T009 reproduced failures before fixes; T003/T007 have passing companion
and real-engine save tests; T004/T005 passed full component/browser checks and final
race-instrumented repeated browser cases; T006 includes final diff/rule self-review;
T010–T016 address review follow-ups, the 0.13.0 version bump, TS 6 editor
compatibility, dependency security, bundle size, and the CI-equivalent local run
without changing evaluated behavior.
