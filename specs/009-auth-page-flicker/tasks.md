# Tasks: Stable authentication page handoff

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: Complete

- [x] T001 — Inspect rendering/session/router paths and rules; NH-01/04;
  evidence: documented root race and existing completed spec 008.
- [x] T002 — Record spec, design, rule checks and acceptance before code; NH-01/05.
- [x] T003 — Add delayed-route regression in frontend/e2e/auth.spec.ts;
  FR-01, AC-01/02, NFR-01; depends T002; done: pre-fix shell visible during held route; see plan evidence.
- [x] T004 — Gate shell in frontend/src/App.vue on auth and current protected route;
  mount after initial router readiness in frontend/src/main.ts; keep authenticated
  handoff status in frontend/src/components/auth/LoginPage.vue;
  FR-01/02; depends T003; done: OAuth/cookie handoff passes on both widths.
- [x] T005 — Run auth browsers on both viewports, Jest, lint, build; AC-03, NH-05;
  depends T004; done: 26 browser cases, 152 Jest tests, lint/build passed; plan records warnings.
- [x] T006 — Review final diff/acceptance/rules and update artifacts; NH-01/04/05;
  depends T005; done: diff/rule review and evidence in plan; no approval implied.

- [x] T007 — Extract/test login composable and simplify presentation/browser observer;
  FR-01/02, NH-04/05, FE-01; evidence: companion unit tests and retained auth browsers.
- [x] T008 — Record scoped before/after complexity/coverage/CRAP, run Node 24
  Jest/lint/build and auth/layout/panel browsers; NH-05; update final evidence.

T007/T008 completion evidence: [quality.md](quality.md), reproducible per-callable report,
100% composable branch/statement coverage, 160 Jest tests and 26 browser checks passed.
