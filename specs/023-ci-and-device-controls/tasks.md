# Tasks: Faster CI and cleaner device controls

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: Complete locally; CI rollout pending

- [x] T001 — Measure PR/main job/step costs and inspect rules/contracts; FR-01/NH-01.
  Evidence: linked runs, 405s/413s E2E bottleneck, existing shared-hub restriction.
- [x] T002 — Specify behavior/acceptance before source edits and update for branch
  feedback; NH-01/04. Evidence: drafts consolidated into this final spec/plan/tasks set.
- [x] T003 — Partition E2E into four isolated runners and retain gates/artifacts;
  FR-01/AC-01–03. Depends T001/02. Evidence: discovery/execution union exact, four
  fresh-server partitions pass; actionlint and aggregate negative cases pass.
- [x] T004 — Implement natural cards, compact footers, thin responsive horizontal
  sliders and device links; FR-02/AC-04–06. Depends T002. Evidence: source-owned
  geometry/navigation/native-gesture checks; light/dark screenshot review.
- [x] T005 — Remove progress labels/spinners/pulses/fades throughout device controls;
  FR-03/AC-07/08. Depends T004. Evidence: no progress DOM; confirmed text/native
  commands/duplicate prevention/timeout recovery preserved; shared lifecycle unchanged.
- [x] T006 — Run complete isolated browser suite, Jest/lint/build; SC-01/NH-05.
  Depends T003–05. Evidence: 173 E2E passes / nine existing skips, 282 Jest tests,
  lint/build pass; actual environment, warnings and limitations in plan.
- [x] T007 — Consolidate current unmerged drafts and fix branch planning reuse;
  FR-04/AC-09. Depends T006. Evidence: single 023 directory; workflow/AGENTS/skill
  distinguish current branch iterations from completed earlier work; references repaired.
- [x] T008 — Validate consolidation links, skill, whitespace and final partition
  inventory; FR-04/SC-01. Depends T007. Record actual results in plan.
  Evidence: final 182-case union exact (59/32 per viewport); bundled skill validator,
  relative links, stale-reference search and whitespace pass; earlier specs untouched.

Hosted E2E runtime target/artifact execution remain rollout checks; local completion
does not imply GitHub CI passed. Future requests in this branch update these files.
