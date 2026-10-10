# Tasks: Frontend redesign and customisable design system

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: Phase 0 complete; Phases 1–6 blocked on Q-01

## Phase 0 — Documentation and samples (authorised 2026-10-09)

- [x] T001 — Document current code paths, contracts and gaps in `current-state.md`;
  gate: NH-01/NH-04; done: file committed.
- [x] T002 — Draft spec, plan, design system, device support, automation editor docs and
  ADR-001; gate: NH-01; done: files committed, open questions listed.
- [x] T003 — Build 5 static design samples (plus single-file `samples/standalone/*.html` with an All pages view) covering every FR-01 screen with customiser in
  `samples/`; IDs: AC-01, AC-02; done: rendered in headless Chromium for every design ×
  screen (see plan traceability).
- [x] T003b — Round 2 after owner feedback (2026-10-10): keep Hearth, add four structurally
  different designs (`design-directions.md`, `samples/designs/*.js`); done: verified in Chromium.
- [x] T003c — Round 3 (2026-10-10): Hearth + Panel merge with logo, shared themes, activity
  switch and phone editor layout; done: verified in Chromium (desktop + 375 px).
- [ ] T004 — Owner review: choose design / combination; IDs: Q-01, SC-01; done: decision
  recorded in spec.

## Phase 1 — Foundations (blocked: T004)

- [ ] T010 — Token files + `useThemeSettings` (migration from `localStorage.theme`,
  validation, system mode); IDs: FR-02, AC-18; Jest tests first.
- [ ] T011 — PrimeVue preset mapped to tokens; remove hard-coded colours in touched
  components; contrast check script; IDs: NFR-02; depends: T010.
- [ ] T012 — `AppShell` variants + header status (connection text, permit-join chip);
  IDs: AC-16, AC-17; depends: T011.
- [ ] T013 — `components/ui/` primitives; IDs: FR-01; depends: T011.
- [ ] T014 — Registries (`protocols`, `capabilities`) with Jest coverage incl. unknown
  protocol; IDs: FR-07, FR-08, AC-07, AC-08.
- [ ] T015 — `useDeviceCommand` pending/confirm/timeout with Jest; IDs: FR-04, AC-04.
- [ ] T016 — Appearance settings tab + header quick customiser; IDs: AC-02, AC-18.

## Phase 2 — Dashboards (depends: Phase 1)

- [ ] T020 — `selectors/network.ts` + `useActivityFeed` (bounded, memory only); IDs:
  FR-05, AC-09, NFR-04; Jest + perf script.
- [ ] T021 — Overview route; Home areas (replace GroupDashboard layout, keep contract);
  fix missing `renameDashboardGroup` mutation and `alert()` usage with regression
  tests; IDs: AC-03, AC-05, AC-06.
- [ ] T022 — Devices dashboard + device list (search, filters, sort, chips); IDs: AC-07.

## Phase 3 — Device page (depends: Phase 1)

- [ ] T030 — Header with status/protocol chips; tabs About/Controls/Settings/Metrics;
  capability-driven controls; protocol-gated actions; fix `setDeviceDeConfigfaults`
  typo with regression test; IDs: AC-04, AC-07.

## Phase 4 — Automations (depends: Phase 1)

- [ ] T040 — `describeTrigger`, `validateAutomation`, `useAutomationDraft` with round-trip
  fixtures from existing automations; IDs: AC-11, AC-11b, AC-14 (tests before UI).
- [ ] T041 — Flow editor UI replacing panel stack; IDs: AC-10, AC-12, AC-15; depends: T040.
- [ ] T042 — Run lock via `post()` helper; automations list redesign; IDs: AC-13.

## Phase 5 — Remaining screens

- [ ] T050 — Assistant, console, settings, login, dialogs, toasts, permit join; IDs:
  FR-01, AC-16, AC-17.

## Phase 6 — Handoff

- [ ] T060 — Remove legacy preset/switch; bundle comparison; IDs: NFR-03.
- [ ] T061 — `npm test -- --runInBand`, `npm run lint`, `npm run build`; viewport,
  keyboard, offline/reconnect manual checks at 360/768/1280/1920; gate: NH-05; record
  actual results in plan.
- [ ] T062 — Final constitution check; update `docs/sdd/repository-map.md`; gates:
  NH-01/04/05.
