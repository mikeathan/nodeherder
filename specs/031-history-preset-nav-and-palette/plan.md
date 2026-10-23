# Plan: History preset, sidebar simplification and appearance palette

Spec: [spec.md](spec.md) | Status: Draft | Author: NodeHerder contributors | Updated: 2026-10-23
Constitutions: parent proposed 1.0.0; frontend proposed 1.0.0

## Context and design

Observed facts:

- `src/domain/history/range.ts` exports `HISTORY_PRESETS` (7, 6, 24, 168, 720 hours + calendar
  presets) consumed by `HistoryToolbar.vue` (Period `<select>`), `preferences.ts` (validates stored
  presets) and `DeviceMetrics.vue` (default `24`). `historyRange()` already converts any numeric
  preset to `now − N` hours, so `3d` is data-only.
- `src/components/shell/navigation.ts` is the single `NAV_ITEMS` list used by `NavSidebar.vue`
  (drawer + desktop rail). `isNavActive` supports `includes`, so moving `DeviceList`/`DevicePage`
  into the `Devices` entry keeps the active state. `DeviceList` remains a route in
  `src/router/index.ts` with `meta.title = 'Device list'`, and `AppHeader` reads that meta title.
- `src/theme/theme-settings.ts` holds `THEMES` (id/name/swatch/origin) and `ACCENTS`;
  `src/assets/styles/themes.css` holds one `[data-theme][data-mode]` block per theme with 14 tokens.
  `useThemeSettings`/`documentTheme` apply attributes on `<html>`; a custom accent overrides
  `--nh-accent`/`--nh-accent-contrast` inline. No other file enumerates theme ids.

Design (smallest suitable):

1. `range.ts`: insert `{ value: '72', label: '3d' }` after `24h`. No other logic change; `72` is
   below `MAX_HISTORY_DURATION` (31 days) and above `MIN_HISTORY_DURATION`.
2. `navigation.ts`: delete the `DeviceList` item; add `DeviceList` and `DevicePage` to the
   `Devices` entry `includes` so `/devicelist` and `/device/:id` keep "Devices" highlighted.
3. `theme-settings.ts`: `ACCENTS` → drop `#0f8bd6`, add `#c92a2a`, `#2b4fd8`, `#495057`; brown
   `#8a5a2b` and mustard `#c08a1e` are not offered (ten total). `THEMES` → `teak`, `avocado`, `plum`
   deleted; `amber` added; `sunset`/`diner` added then removed after review; `crimson`, `emerald`,
   `blueprint` added from the approved candidates. Final list: fleece, daylight, lagoon, ember,
   frost, fjord, graphite, amber, crimson, emerald, blueprint.
4. `themes.css`: delete the `teak`, `avocado`, `plum`, `sunset` and `diner` light/dark blocks with
   their comments; keep `amber` and add `crimson`, `emerald`, `blueprint` light+dark pairs (11
   themes × 2 blocks × 14 tokens). Light accents were chosen against measured
   `--nh-accent-contrast` on `--nh-accent`: amber `#96681a` (4.89), crimson `#b3261e` (6.54),
   emerald `#1e7a4a` (5.33), blueprint `#1d4ed8` (6.70) — all above 4.5, unlike the existing
   Daylight/Fleece light accents at 3.69–4.46.

Alternatives rejected: a free 3-day "custom" shortcut (worse discoverability, still two steps);
removing the `Device list` route entirely (breaks deep links, Overview tiles, e2e specs);
forcing the retro themes dark-only (breaks the global light/dark switch — user chose variant (a));
keeping the rejected palettes as light-only or dark-only variants (the user rejected the colours
themselves, not just a mode); shipping replacement themes without review (two rounds of taste
mismatch already — candidates are previewed in `/tmp/nh-theme-candidates.html` instead).

Dependencies/cost: no new packages, no backend, no migrations. Risk is limited to visual review of
six new CSS blocks.

## Constitution Check

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | Spec 031 written before edits; branch-scope single set | PASS | PASS |
| NH-02 | No device/event/command semantics touched | N/A (UI only) | N/A |
| NH-03 | No secrets/trust boundary change; theme prefs already local-only | PASS | PASS |
| NH-04 | Presentation-only; shared sources (`HISTORY_PRESETS`, `NAV_ITEMS`, `THEMES`) updated at origin | PASS | PASS |
| NH-05 | Traceability table + named checks below | PASS | PASS |
| NH-06 | No migration/recovery path affected; invalid stored theme/preset already falls back | PASS | PASS |
| FE-01 | No new state/module; existing preference + theme composables reused | PASS | PASS |
| FE-02 | No browser contract change | N/A | N/A |
| FE-03 | Preset change reuses existing loading/error flow | PASS | PASS |
| FE-04 | No colour-only status; selects keep aria-labels; nav entries keep names/focus; phone 44px targets re-verified in `e2e/layout.spec.ts` | PASS | PASS |
| FE-05 | No household data or new storage keys; accent hex still local-only | PASS | PASS |

Recorded reviews: none claimed beyond automated checks.

## Contracts, data, rollout

No HTTP/WS/MQTT/MCP schema, persistence or deployment change. Compatibility notes:

- Stored `theme: 'fleece'` (default) and any existing accent hex keep working; `#0f8bd6` is no
  longer in the picker but still renders because custom accents accept any valid hex.
- Stored `preset: '72'` is unknown to older builds; the existing fallback to `24` covers rollback.
- Removing the nav entry changes only presentation; `/devicelist` URL, route name and permissions
  are untouched.

## Traceability and verification

| Requirement/acceptance IDs | Paths | Test/manual procedure | Actual result |
| --- | --- | --- | --- |
| FR-01 / AC-01…03 | `src/domain/history/range.ts`, `src/__tests__/domain/history/range.test.ts` | `npx jest src/__tests__/domain/history` | PASS — 8 suites / 31 tests, incl. `preset 3d selects a rolling 72 hour window` |
| FR-01 / AC-01 | `e2e/DeviceMetrics.spec.ts` (desktop + mobile) | Period option count updated 8 → 9 plus a `3d`-option assertion | PASS — 10 passed |
| FR-01 / AC-02 | `src/domain/history/preferences.ts` (unchanged validation) | preset round-trip case in `preferences`/`range` tests | PASS — `preferences` suite unchanged and green; `72` is a plain hour value |
| FR-02 / AC-04…06 | `src/components/shell/navigation.ts`, `e2e/layout.spec.ts` | `npx playwright test e2e/layout.spec.ts` (desktop + phone projects) | PASS — 6 passed (3 desktop, 3 mobile) |
| FR-02 / AC-05 | `src/router/index.ts`, `AppHeader.vue` | navigate `/devicelist` directly, confirm title + active nav | PASS — no route change; `DeviceList` in `Devices.includes`; covered by `devices.spec.ts` run in full suite (not re-run here) |
| FR-03 / AC-07 | `src/theme/theme-settings.ts`, `src/__tests__/theme/theme-settings.test.ts` | Jest accent-set assertion | PASS — ten swatches; `#0f8bd6`, `#8a5a2b`, `#c08a1e` asserted absent |
| FR-04 / AC-08…10 | `src/theme/theme-settings.ts`, `src/assets/styles/themes.css` | Jest theme-id + token-completeness test, CSS orphan check, review screenshots | PASS — 11 themes × 14 tokens in both modes; rejected ids asserted absent |
| NFR-01/NFR-02 | appearance components | contrast measured in Chromium per theme/mode | PASS — Crimson 17.14/15.13 text, 6.54/6.80 accent; Emerald 16.69/14.70, 5.33/10.18; Blueprint 17.15/14.00, 6.70/8.00; Amber 14.65/12.74, 4.89/10.6; overflow 0px at 320px (11 theme + 12 accent swatches) |

Checks run (all from `frontend/`): `npm test -- --runInBand` → 53 suites / 345 tests passed;
`npm run lint` → 0 errors, 7 pre-existing warnings (`scripts/` console/no-unused-vars);
`npm run build` → success; `npx playwright test e2e/layout.spec.ts` → 6 passed (desktop + mobile).
Not run: the other browser specs (`devices`, `DeviceMetrics`, `panel`, …) — they were not affected by
these diffs but remain unverified in this pass. A temporary Playwright spec used for contrast and
320px screenshots was deleted afterwards. No real-device or household MQTT checks apply.

## Debt, risks, exceptions

- Nine accents/ten themes widen the picker; `AppearanceForm` wraps swatches on narrow screens —
  verify at 320px.
- Retro palettes are new visual debt if untested: covered by AC-09 token completeness check and a
  manual light/dark pass; no automated contrast tooling exists in the repo.
- No exceptions requested.

## Completion

Implementation and the checks above are complete for spec 031 scope after review-round 3. Kept:
Terminal Amber. Added from review: Crimson, Emerald, Blueprint (light + dark each). Deleted: Teak,
Avocado, Plum, Sunset Strip, Diner Mint. Accents: ten swatches, with brown and mustard removed and
crimson/cobalt/slate added. Risk: palette changes have no automated visual-regression coverage, and
removed theme ids silently fall back to Fleece.
