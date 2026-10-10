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

## Left to do

1. **T109:**
   - Check the Assistant view and the legacy dialogs (`EntityViewDialog` and others) under
     the new preset, and restyle them where needed.
2. **T110:** add an e2e test that edits an automation (change a value, then save).
3. **Remove unused legacy files:**
   - Automation editor: `automations/{Viewer,Editor,Creator,DeviceAutomation,Trigger}.vue`,
     `actions/*`, `conditions/*`, `schedule/*`.
   - `controls/Navigation*`, `controls/PermitJoinTimer.vue`, `layout/*`,
     `mixins/composables/useAuthentication.ts`, `services/theme.service.ts`,
     `themes/material_blue.js`, `assets/styles/variables.css`.
   - The popup OAuth code in `auth.service.ts` and `OAuthDialog`, if they are unused.
   - Check each removal with grep and a build.
4. **T111:**
   - Self-review the full diff.
   - Run `npm test -- --runInBand`, `npm run lint`, `npm run build`, `npm run test:e2e` and
     `go test ./...`.
   - Compare bundle size with NFR-03. Baseline gzip: JS 641 KB, CSS 66 KB. Charts are now
     loaded only when needed.
   - Update the docs and repository map: `src/domain` replaces the planned `registry/` and
     `selectors/` folders.
   - Write ADR-003 for the `@fontsource/figtree` and `@fontsource/lexend` dependencies.
