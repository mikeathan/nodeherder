# Entity/history quality check

From the repository root, after a fresh frontend coverage run:

```sh
cd frontend
npm test -- --runInBand --coverage
cd ..
node frontend/tools/quality/history-quality.cjs frontend/coverage/coverage-final.json
```

Use Node 24 and locked TypeScript 5.9.2. The script compares changed/new production
functions against `771d2a9b2f64529d5a927b3d7e1d2ff2f03f5f07`; deleted predecessors
remain in the baseline. It checks TS, Vue script blocks and mock-server JS.

Complexity is 1 plus if/conditional expressions, loops, cases, catches and logical
operators. Nested functions are scored separately. Vue script extraction preserves
source lines. CRAP is `C² × (1 − coveredFraction)³ + C`; fail at C > 10 or CRAP > 30.
Istanbul statement coverage supplies the fraction where available. Otherwise the
report uses a conservative zero-coverage upper bound, explicitly labelled. Baseline
scores also use that upper bound; they are not historical measured coverage.

Vue templates/generated render functions are not scored by this script. Source-owned
Playwright tests verify rendered controls, charts, focus, cancellation and cleanup.
Regenerate coverage after production edits; stale source positions are invalid evidence.
JSON reports are generated artifacts, not committed snapshots. See the spec's
[implementation evidence](../../../specs/007-frontend-redesign/implementation-evidence-2026-10-18.md)
for results and Go tooling.
