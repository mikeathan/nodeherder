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
- `vue-tsc` and `vite build` pass at this checkpoint.

## Left to do

1. **T109:**
   - Restyle Settings with tabs and add an Appearance tab that reuses
     `shell/AppearanceForm.vue`.
   - Restyle the Console and Assistant wrappers.
   - Check the legacy dialogs (`EntityViewDialog` and others) under the new preset.
2. **T110:**
   - Rewrite `frontend/e2e/support/ui.ts` page objects for the new UI. The specs assert wire
     frames and should not need changes.
   - Add e2e tests for: drag/arrow reorder (`saveDashboardGroup` with `order`), panel mode,
     the activity switch, and no horizontal overflow on the phone viewport.
   - Run `npm run test:e2e`.
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
