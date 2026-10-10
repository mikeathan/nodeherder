# Spec 007 implementation status

Branch `claude/eloquent-lovelace-rus03g`. This file says what is done and what is left, so work
can resume from any checkpoint.

## Done

- **Safety net (T100–T102):** Jest wire-contract tests for store actions, WS routing and the
  auth service. Playwright e2e harness against the mock hub at 1280×800 and 390×844, with a CI
  job. Backend `DashboardGroup.order` field with Go tests.
- **Domain layer (T103):** framework-free and Jest-tested. Modules in `src/domain/`: exposes,
  devices, network, dashboard, automation, activity, commands, time, panel. Also
  `src/theme/theme-settings.ts`.
- **Foundations (T104):**
  - Tokens: `assets/styles/{tokens,themes,base}.css`, with 10 themes in light and dark.
  - PrimeVue preset mapped to the tokens (`theme/primevue-preset.ts`). The Figtree and Lexend
    fonts replace Roboto.
  - UI primitives in `components/ui/`. They use an MDI-based icon registry.
  - Composables:
    - `useThemeSettings`, `useHub`
    - `useDeviceCommand` (pending state, 5 s timeout, no retries)
    - `useActivityFeed` (on/off switch, kept in memory only)
    - `useDashboardGroups`, `usePermitJoin`, `useSession`, `useHubConnection`
    - `useAutomations`, `useAutomationDraft`
    - `useConfirm`, `useClock`, `useMediaQuery`
- **Screens (T105–T108):**
  - Shell: sidebar, drawer on screens up to 960 px, header with logo, connection status, mode
    and appearance controls, status banners, permit join.
  - Sign-in page.
  - Overview with the activity switch.
  - Home: drag areas by their handle, or use the arrow buttons, in Edit layout mode. Each move
    sends `saveDashboardGroup` with `order`.
  - Devices (cards), Device list (sort and filters), and Device page (tabs; fixes the deep-link
    back bug).
  - Panel mode at `/panel`.
  - Automations list, creator (pick a device) and a single-page When/If/Then editor. Each
    action type has its own editor (strategy map); conditions and actions can be reordered;
    validation messages appear on the fields; there is an unsaved-changes guard; unchanged
    drafts save exactly as loaded.
- **T109 (part):** Settings has sections in a vertical tab list, including Appearance (it reuses
  `AppearanceForm`). The Console wrapper is restyled.
- **T110:** page objects are rewritten. New e2e specs: area reorder (arrows and touch/mouse drag),
  activity switch, panel mode, and a horizontal-overflow check on 11 screens.
- Checks at this checkpoint:
  - Jest: 135/135 pass.
  - e2e: 26/26 pass at desktop 1280×800 and mobile 390×844.
  - Lint: 0 errors.
  - `vue-tsc` and `vite build` pass.

## Also done (final pass)

- Removed the legacy UI files listed earlier, checked with an import and tag scan plus a
  build.
- Removed the popup OAuth code and the Roboto font.
- ADR-003 records the Figtree and Lexend font choice.
- The plan's evidence table, the tasks checklist and `docs/sdd/repository-map.md` are
  updated.
- Fixes from the visual review:
  - Tile labels use the short expose name; the description is now a tooltip.
  - "Last seen" is hidden when the hub has not reported a time.
  - The mobile drawer slides closed instead of disappearing.
- Final checks:
  - Jest: 136/136 pass.
  - e2e: 26/26 pass at both viewports.
  - Lint: 0 errors.
  - `vue-tsc`, `vite build` and `go test ./...` pass.

## Left to do

1. **T109 (rest):** the Assistant view and the legacy dialogs (`EntityViewDialog`, the
   selection dialogs) follow the new theme through the PrimeVue preset and token aliases,
   but their layouts are unchanged.
2. **T061:** manual keyboard, screen-reader and offline/reconnect checks at 360, 768 and
   1920 px.
3. **T111:** a review by a person; the constitutions and ADRs are still proposed, not
   approved.
4. Optional: an e2e test that edits an automation (change a value, then save).
