# ADR-002: Browser end-to-end tests with Playwright against the mock hub

Status: Proposed (not approved) | Date: 2026-10-10 | Spec: [007](spec.md) NFR-06

## Context

The redesign replaces every screen. The owner requires that automations, panels, dashboards
and sign-in keep working. Existing checks are Jest unit tests for store/contracts/utilities
(44 tests) plus lint and build; no test renders a component or exercises a user flow. The
frontend constitution states Vue mounting tests are not assumed configured and asks plans to
choose browser tooling explicitly. `frontend/tools/server` already mocks the HTTP API, the
WebSocket protocol and the OAuth redirect.

## Decision

Add `@playwright/test` as a **dev-only** dependency and a `frontend/e2e/` suite that:

- starts `tools/server` (mock hub, port 4110) and `vite --mode e2e` (port 4100, `.env.e2e`
  points to localhost) through Playwright `webServer`;
- runs critical flows in Chromium at desktop (1280×800) and phone (390×844, touch) sizes;
- asserts behaviour through accessible roles/labels, and asserts **wire contracts** by
  recording WebSocket frames sent by the browser (e.g. exact `saveAutomation`,
  `deviceSetValue`, `saveDashboardGroup` payloads);
- never contacts real devices (NH-05).

Pure logic stays in Jest (fast, isolated). No Vue Test Utils/jsdom component layer is added:
e2e covers the integration risk, Jest covers logic, avoiding a third test stack.

## Alternatives

1. Vue Test Utils + jsdom in Jest — needs an SFC transformer (`@vue/vue3-jest`), PrimeVue/
   ApexCharts stubs, and still misses routing, WebSocket and layout/viewport behaviour.
2. Vitest browser mode — migration of the Jest suite (FE-01/Vue rules discourage assuming it).
3. Manual checklist only — not repeatable; does not satisfy NFR-06.

## Consequences

+ Real-browser evidence for the flows that matter, at two viewports, in CI.
− CI needs a Chromium download (~150 MB, cached) and ~1–2 minutes; one extra job.
− Mock hub fidelity limits what is proven (backend parity is still covered by Go tests);
  the mock must be updated alongside contract changes (e.g. `order`).
Operational cost: dev dependency only; nothing ships in the production bundle.
