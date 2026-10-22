# Tasks: History, charts and mobile shell

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | ADR: [adr-006](adr-006-chart-interactions.md)
Status: Planned (no implementation started)

Each task maps to requirement/acceptance IDs, names the paths it owns, and states its completion
evidence. New tests go in the companion file of the source they cover (for example
`frontend/src/__tests__/domain/history/stats.test.ts` for `src/domain/history/stats.ts`).

## Phase 1 — Preparation and proof of the baseline

- [ ] T001 — Record the current baseline in the plan: chart heights (235/105), history selectors,
      bundle size from `npm run build`, and the existing history e2e list. Reproduce the three
      stated gaps (no zoom, no Yesterday, sheet dead end) before editing. IDs: NH-01, NH-05;
      done: baseline numbers and screenshots/traces attached to the PR description.
- [ ] T002 — Confirm `chartjs-plugin-zoom` v2 against Chart.js 4.5.1 on a throwaway chart: wheel,
      drag-select, pinch and pan, with `limits` clamping to 5 min…31 days and to the loaded window.
      IDs: FR-03, AC-04; depends: T001; done: findings recorded; ADR amended with the pinned versions
      and any gesture that needs our own handling.

## Phase 2 — Pure domain (no Vue, fully unit tested)

- [ ] T010 — Extend `frontend/src/domain/history/range.ts` with the `yesterday` preset and exported
      clamp/step helpers (5 min…31 days, 20 % pan); IDs: FR-01, AC-01, AC-04; depends: T001;
      done: `src/__tests__/domain/history/range.test.ts` covers midnight boundaries, DST day
      lengths, custom ranges and clamping.
- [ ] T011 — Add `frontend/src/domain/history/stats.ts`: bucket the **raw decoded** samples by local
      day (`null` for empty buckets, reuse `numericStats`, sample count, covered window, partial
      flag); IDs: FR-04, AC-07, AC-08, AC-13; depends: T010; done:
      `src/__tests__/domain/history/stats.test.ts` covers DST, empty buckets, single sample, all-null
      series, partial input, and asserts the result differs when decimated input is passed by mistake.
- [ ] T012 — Add `frontend/src/domain/history/export.ts` producing CSV rows (ISO timestamp, value,
      unit) and a guard for partial/empty series; IDs: FR-04, AC-09; depends: T011; done:
      `src/__tests__/domain/history/export.test.ts` covers escaping, units, partial refusal.
- [ ] T013 — Add `frontend/src/domain/history/preferences.ts` for `nh-history-pref:<deviceId>`
      (versioned read/write, defensive parse, no throw on storage failure); IDs: FR-01, AC-02,
      FE-02, FE-05; depends: T010; done: `src/__tests__/domain/history/preferences.test.ts` covers
      version mismatch, unreadable storage and unknown expose.

## Phase 3 — Chart component and interactions

- [ ] T014 — Add the zoom plugin to the lazy chart chunk and implement
      `frontend/src/components/chart/history/interactions.ts` around it: `limits` from the range
      policy, button and keyboard actions (plugin has no keyboard), readout scrub, `touch-action`
      rules and reduced-motion handling; IDs: FR-03, AC-04/05/06, NFR-01/02; depends: T010, T002;
      done: `src/__tests__/components/chart/history/interactions.test.ts` covers clamp mapping,
      keyboard steps, disabled gestures and no-op bounds.
- [ ] T015 — Add `HistoryChart.vue` (responsive height by breakpoint, focusable container with
      accessible name and range, readout slot, end marker, `touch-action`); IDs: FR-02, FR-03,
      AC-03/05/06, FE-04; depends: T014; done: component covered by options/render tests and used by
      the view below.
- [ ] T016 — Extract `HistoryToolbar.vue` (presets, expose chips, history/statistics switch,
      refresh/export) and `HistoryStats.vue`; IDs: FR-01, FR-04, AC-01/02/07; depends: T015, T011;
      done: `src/__tests__/components/chart/history/*` cover state wiring without mounting Vue.
- [ ] T017 — Recompose `HistoryView.vue` from the parts, keep the `compact` prop and every existing
      loading/error/empty/partial message; IDs: FR-06, AC-13/14; depends: T016; done: existing
      history e2e selectors still resolve.
- [ ] T018 — Move the samples table into `HistorySamples.vue` unchanged in behaviour; IDs: AC-14;
      depends: T017; done: table paging test still passes.

## Phase 4 — Pages, sheet and shell

- [ ] T019 — `DeviceMetrics.vue`: wide layout for the history route, preference restore, statistics
      mode, export, keyboard hints, and a “narrow this range” action when the series is partial;
      IDs: FR-01, FR-02, FR-04, AC-01/02/03/08/09; depends: T017, T013; done: desktop e2e covers
      preset/expose switch, persistence across reload, export, and a truncated series showing its
      covered window plus the narrower-range action without changing the selection on its own.
- [ ] T020 — `EntityViewDialog.vue`: compact chart, "View full history" to
      `/devicepage/:id?tab=history`, drag-to-dismiss (top-of-scroll, 120 px, reduced-motion aware),
      safe-area actions; IDs: FR-05, AC-11; depends: T015; done: sheet e2e covers drag, Escape,
      close and deep link; existing dialog tests unchanged.
- [ ] T021 — Shell: mark primary destinations in `navigation.ts`, add `BottomNav.vue`, render it in
      `AppShell.vue` below 720 px, keep the drawer, add `--nh-safe-bottom`; IDs: FR-05, AC-10,
      NFR-03; depends: T001; done: layout e2e asserts 44 px targets, `aria-current`, no horizontal
      scroll and safe-area padding at 390×844.
- [ ] T022 — Platform: `index.html` gains `viewport-fit=cover` and `theme-color`;
      `public/manifest.webmanifest` plus icon; verify the app still boots when the manifest 404s;
      IDs: FR-05, AC-12, NFR-03; depends: T021; done: e2e asserts the meta tags and manifest fetch.

## Phase 5 — Documentation and supersession

- [ ] T023 — Retire the root `charts_todo.md`: move anything still relevant into `plan.md` debt,
      delete the file, and note the supersession. IDs: NH-01; depends: T001; done: no dangling
      references (`rg charts_todo`).
- [ ] T024 — Mark spec 007 T209–T212 superseded with a pointer to 029 and update the history
      description in the docs that still show the select-based flow. IDs: NH-01; depends: T023;
      done: links resolve and the older tasks are not silently re-opened.

## Phase 6 — Verification and handoff

- [ ] T030 — Component checks from `frontend/`: `npm test -- --runInBand`, `npm run lint`,
      `npm run build`, `npm run type-check:e2e`. IDs: NH-05, NFR-01; depends: T017–T022; done: pass
      with the recorded bundle delta, compared per chunk (initial entry unchanged, chart chunk
      within +15 kB gzip).
- [ ] T031 — Browser checks: `CI=true npx playwright test` on desktop and mobile projects, including
      new gesture tests driven by real input (mouse drag, wheel, CDP two-pointer pinch, keyboard) and
      the a11y assertions (names, focus ring, 44 px targets, no horizontal scroll). IDs: FR-03,
      AC-04…AC-12, NFR-03; depends: T030; done: pass, with the manual desktop/mobile pass recorded.
- [ ] T032 — Prove no backend or contract drift: `backend/`: `go test ./...`; confirm no diff under
      `backend/` and no change to `contracts/history.ts` request shape. IDs: AC-13/14, BE-01…BE-05;
      depends: T030; done: command output recorded, diff shows frontend + docs only.
- [ ] T033 — Review pass: re-read the diff against every applicable NH/FE rule, confirm truthfulness
      cases (gaps, unknown, partial), update the plan's traceability table with actual results, and
      list anything unrun. IDs: NH-01…NH-06, FE-01…FE-05; depends: T031, T032; done: plan evidence
      complete and honest; no unapproved violation.
