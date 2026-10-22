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
- [ ] T002 — Prototype the local interaction plugin against Chart.js 4.5.1 on a throwaway chart:
      wheel, drag-select (mouse and single finger), two-pointer pinch, pan, keyboard, and `touch-action`
      behaviour, all clamped to 5 min…31 days and to the loaded window. IDs: FR-03, AC-04/06;
      depends: T001; done: findings recorded (including any browser that intercepts pinch), and the
      plugin registered only after `Chart.register` in `HistoryCanvas.vue`.

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
- [ ] T012 — Add `frontend/src/domain/history/export.ts` with the two documented schemas
      (`timestamp,value,unit` for history; `bucket_start,bucket_end,min,average,max,last,samples,unit`
      for statistics), ISO-8601 timestamps with offset, empty cells for gaps and empty buckets, and a
      partial/empty guard; export always covers the selected range, never the zoomed viewport.
      IDs: FR-04, AC-08/09; depends: T011; done:
      `src/__tests__/domain/history/export.test.ts` covers both headers, gap/empty-bucket cells,
      escaping, offset formatting and partial refusal.
- [ ] T013 — Add `frontend/src/domain/history/preferences.ts` for `nh-history-pref:<deviceId>`
      (versioned read/write, defensive parse, no throw on storage failure); IDs: FR-01, AC-02,
      FE-02, FE-05; depends: T010; done: `src/__tests__/domain/history/preferences.test.ts` covers
      version mismatch, unreadable storage and unknown expose.

## Phase 3 — Chart component and interactions

- [ ] T014 — Implement `frontend/src/components/chart/history/interactions.ts` as our own Chart.js
      plugin: wheel, drag-select, two-pointer pinch, pan, keyboard and `touch-action`, all through one
      `applyRange` that clamps and never leaves the loaded window; the buttons call the same function.
      No new dependency. IDs: FR-03, AC-04/05/06, NFR-01/02; depends: T010, T002; done:
      `src/__tests__/components/chart/history/interactions.test.ts` covers clamp mapping, keyboard
      steps, pinch scale maths, disabled gestures and no-op bounds; the build shows no new import in
      the chart chunk.
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

- [ ] T019 — `DeviceMetrics.vue`: wide layout for the history route, preference restore (including
      validated custom bounds with the `24h` fallback), statistics mode for numeric series only,
      export, keyboard hints, and a “narrow this range” action when the series is partial. Expose
      precedence is URL → saved preference → first chartable expose. IDs: FR-01/02/04,
      AC-01/02/03/07/08/09; depends: T017, T013; done: desktop e2e covers preset/expose switching,
      restore after reload, a stale/invalid custom range falling back to `24h`, a binary series
      without the switch, export, and a truncated series naming its covered window.
- [ ] T020 — `EntityViewDialog.vue`: compact chart, "View full history" to
      `/devicepage/:id?tab=history&expose=<name>`, drag-to-dismiss (top-of-scroll, 120 px,
      reduced-motion aware), safe-area actions; IDs: FR-05, AC-11; depends: T015; done: sheet e2e
      covers drag, Escape, close, and the deep link landing on the same expose even when a different
      one is saved in preferences.
- [ ] T021 — Shell: add a `primary` marker and a derived drawer list in `navigation.ts`, add
      `BottomNav.vue` (Home, Devices, Automations, Assistant, More→drawer), render it in
      `AppShell.vue` below 720 px, keep the drawer, add `--nh-safe-bottom`; IDs: FR-05, AC-10,
      NFR-03; depends: T001; done: layout e2e asserts 44 px targets, `aria-current` for tabs and the
      `includes` rule, More current on other routes with `aria-expanded` toggling the drawer, no
      horizontal scroll and safe-area padding at 390×844.
- [ ] T022 — Platform: `index.html` gains `viewport-fit=cover` and `theme-color`;
      `public/manifest.webmanifest` plus icon; verify the app still boots when the manifest 404s;
      IDs: FR-05, AC-12, NFR-03; depends: T021; done: e2e asserts the meta tags and manifest fetch.

## Phase 5 — Documentation and supersession

- [ ] T023 — Retire the root `charts_todo.md`: move anything still relevant into `plan.md` debt,
      delete the file, and note the supersession. IDs: NH-01; depends: T001; done: no dangling
      references (`rg charts_todo`).
- [ ] T024 — Supersede **only the covered portions** of spec 007: T209 (history query correctness),
      T210 (pure history domain) and T211 (history rebuild) are delivered by 029; T212 stays open for
      Assistant layout, palette/contrast checks and legacy cleanup, with a note that its
      history-dependent parts move to 029; T213 stays open for the flows 029 does not cover. Update
      the docs that still show the select-based history flow. IDs: NH-01; depends: T023; done:
      links resolve, each superseded task names where its obligation went, and nothing unrelated is
      closed.

## Phase 6 — Verification and handoff

- [ ] T030 — Component checks from `frontend/`: `npm test -- --runInBand`, `npm run lint`,
      `npm run build`, `npm run type-check:e2e`. IDs: NH-05, NFR-01; depends: T017–T022; done: pass
      with the recorded bundle delta, compared per chunk (initial entry unchanged, chart chunk
      within +15 kB gzip).
- [ ] T031 — Browser checks: migrate the existing `e2e/DeviceMetrics.spec.ts` Entity/Period
      `selectOption()` assertions to the new chips in this same change, then run
      `CI=true npx playwright test` on desktop and mobile projects, including new gesture tests driven
      by real input (mouse drag, wheel, CDP two-pointer pinch, keyboard) and the a11y assertions
      (names, focus ring, 44 px targets, no horizontal scroll). IDs: FR-03, AC-04…AC-12, AC-14,
      NFR-03; depends: T030; done: pass, no skipped selector, with the manual desktop/mobile pass
      recorded.
- [ ] T032 — Prove no backend or contract drift: `backend/`: `go test ./...`; confirm no diff under
      `backend/` and no change to `contracts/history.ts` request shape. IDs: AC-13/14, BE-01…BE-05;
      depends: T030; done: command output recorded, diff shows frontend + docs only.
- [ ] T033 — Review pass: re-read the diff against every applicable NH/FE rule, confirm truthfulness
      cases (gaps, unknown, partial), update the plan's traceability table with actual results, and
      list anything unrun. IDs: NH-01…NH-06, FE-01…FE-05; depends: T031, T032; done: plan evidence
      complete and honest; no unapproved violation.
