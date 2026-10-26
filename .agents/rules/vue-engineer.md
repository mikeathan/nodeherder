# Vue/TypeScript rules

Scope: UI implementation, refactoring, review. Authority:
[frontend constitution](../../frontend/CONSTITUTION.md), including checks and FE-04/05.
Code paths below are relative to `frontend/`.

- Prefer established `<script setup>`/Composition API for new components; extract shared
  behavior into composables. Preserve compatible styles in existing code.
- Choose `ref`, `reactive`, or `shallowRef` by ownership/reactivity needs. Expose readonly
  state when mutations belong behind actions. Clean up listeners, timers, subscriptions.
- Type props, emits, and shared contracts; expose only needed public APIs. Prefer
  concrete types or narrowed `unknown`; justify unavoidable boundary `any`.
- Preserve reactivity when destructuring. Choose `v-if/v-show` by render/toggle cost;
  split heavy routes/modules when useful. Stabilize props when rendering cost warrants it.
- Reuse PrimeVue/shared controls, theme tokens, and scoped styles. New styling/state
  libraries need specified benefit and design review.
- Reuse `src/contracts/api.ts` for authenticated calls and `src/config/env.ts` for environment.
  Jest maps that environment module to `src/__tests__/mocks/env.ts`.
- Use existing Jest; plan additional mounting/browser tooling rather than assume
  Vitest/Playwright migration.
- Treat external/assistant content as untrusted; raw HTML requires justified sanitization
  satisfying FE-05. Prefer `const`, then `let`.

- Keep functions and components focused on one responsibility. Separate calendar/domain rules,
  formatting, view-model mapping, transport and persistence using cohesive modules and named helpers.
  Avoid nested ternaries and large inline mapping callbacks; prefer guard clauses or typed label maps.
- Add JSDoc to every new or changed exported/public frontend function and composable, including
  methods returned as its public API. Explain purpose, inputs/output, nullable values, units/timezones,
  side effects and ownership where relevant; do not merely repeat the function name or TypeScript type.
- Add concise comments for non-obvious control flow: cancellation/stale replies, lifecycle cleanup,
  copying versus mutation, reference mapping, date-only semantics and ordering. Keep simple code
  self-explanatory rather than adding line-by-line narration.

Review contracts, state ownership, bundle cost, cleanup, and affected UI behavior. Before finishing,
audit touched public APIs for useful JSDoc and complex flows for explanatory comments; passing tests
alone does not satisfy the documentation/readability rules.
