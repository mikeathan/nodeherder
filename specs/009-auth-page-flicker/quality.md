# Authentication cleanup quality evidence

Scope: login check/recovery callables moved from LoginPage.vue to useLoginSession.ts.
Measured 2026-10-14 with locked TypeScript 5.9.2 and Jest/Istanbul under Node 24.21.0.
The retained UI is covered by browser tests; it is not Jest-instrumented.

## Method and limits

The companion script parses each callable separately, excluding nested callables.
Cyclomatic complexity starts at one and adds if/conditional, loop, case, catch,
and &&/||/?? decisions. Coverage maps Istanbul statement start locations inside
that callable, excluding nested callable spans. CRAP = C²(1 - coverage fraction)³ + C.
No statement mapping is treated as unknown, with a zero-coverage upper bound.
This scoped checker is not an established repository-wide CRAP gate or an external
scorer. No new runtime dependency was introduced.

Baseline: committed LoginPage script at 6f299a13 (its callable logic is unchanged
by the initial flicker fix). Jest did not instrument the Vue script, so baseline
coverage is unknown. Its reported CRAP values are conservative upper bounds,
not measured pre-refactor scores. After: all composable statements, branches,
and functions have 100% coverage. Every callable has C <= 4 and CRAP <= 4.
The old check function has C=4 and a CRAP upper bound of 20; the replacement
check has C=3 and measured CRAP=3. runCheck remains C=4 with measured CRAP=4.
This demonstrates reduced check coordination complexity and tested risk, without
claiming a measured 20-to-4 reduction from unknown baseline coverage.

## Per-callable results

| Before callable | C | Coverage | CRAP upper bound |
| --- | ---: | --- | ---: |
| callback@20 | 2 | Unknown | 6 |
| check | 4 | Unknown | 20 |
| sessionCheck | 2 | Unknown | 6 |
| callback@36 | 2 | Unknown | 6 |
| callback@36 | 2 | Unknown | 6 |
| recover | 2 | Unknown | 6 |
| visible | 2 | Unknown | 6 |
| callback@46 | 1 | Unknown | 2 |
| callback@52 | 1 | Unknown | 2 |

| After callable | C | Statement coverage | CRAP |
| --- | ---: | ---: | ---: |
| useLoginSession | 1 | 100% | 1 |
| callback@14 | 2 | 100% | 2 |
| callback@15 | 2 | 100% | 2 |
| callback@16 | 2 | 100% | 2 |
| checkSession | 2 | 100% | 2 |
| runCheck | 4 | 100% | 4 |
| check | 3 | 100% | 3 |
| callback@43 | 2 | 100% | 2 |
| recover | 2 | 100% | 2 |
| onVisibilityChange | 2 | 100% | 2 |
| callback@58 | 1 | 100% | 1 |
| callback@65 | 1 | 100% | 1 |

## Reproduction

From frontend/ (use Node 24 on PATH):

```sh
npm test -- --runInBand --coverage --collectCoverageFrom=src/composables/useLoginSession.ts --coverageDirectory=/private/tmp/nodeherder-login-coverage
```

From the repository root:

```sh
node specs/009-auth-page-flicker/measure-quality.cjs /private/tmp/nodeherder-login-coverage/coverage-final.json
```

## Final checks

- Jest: 24 suites, 160 tests passed; eight new composable tests cover concurrency,
  OAuth completion, mode/session failures, retries, visibility/network recovery,
  and disposal with late success/failure.
- Auth/layout/panel browser checks: 26 passed at desktop/mobile widths against
  localhost mock servers. No household devices involved.
- Lint/build: passed; existing script warnings, duplicate DeviceCard warning,
  large bundle warnings, and Jest delayed-exit warning remain.
- UI controls/styles and API/storage contracts are unchanged. Browser mount tracking
  now uses a Set and a DOM data attribute, removing nested mutation loops and casts.
- Backend and unrelated device/automation browser suites were not run.
