---
name: vue-senior-frontend
description: Applies senior-level (10+ years) Vue 3 and TypeScript expertise. Focuses on Composition API design patterns, strict Type-safety, performance optimization, and scalable SFC architecture.
---

# Senior Vue 3 & TypeScript Frontend Skill

When architecting or writing frontend code, you must adhere to the following Senior Engineer standards:

## 1. Composition API & Logic Reusability

- **Composables over Mixins**: Encapsulate logic in "Composables" (e.g., `useAuth`, `useAsync`). Ensure they return readonly state to prevent external mutations.
- **Ref vs Reactive**: Prefer `ref()` for all state for consistency and better TS inference. Use `shallowRef()` for large datasets (e.g., API responses) to avoid deep reactivity overhead.
- **Lifecycle Discipline**: Always clean up side effects (event listeners, timers) using `onUnmounted`.

## 2. Advanced TypeScript Integration

- **Zero `any` Policy**: Use generics `<T>` or `unknown` with type guards.
- **Strict Props/Emits**: Use the `defineProps<Props>()` and `defineEmits<Emits>()` macros with externalized interfaces for shared components.
- **Component Public API**: Use `defineExpose` strictly to limit what parent components can access in a child's internal state.

## 3. Performance & Architecture

- **Lazy Loading**: Default to `defineAsyncComponent` for heavy UI modules or route-level splitting.
- **v-if vs v-show**: Choose based on the cost of initial render vs. frequency of toggling.
- **Props Stability**: Avoid passing complex inline objects/functions to child components to prevent unnecessary re-renders.

## 4. Design System & Clean Code

- **CSS-in-JS/Utility-First**: Default to Tailwind CSS patterns or Scoped CSS variables for theming. No "magic numbers" in styles.
- **Accessibility (A11y)**: Ensure components use semantic HTML, ARIA attributes where needed, and support keyboard navigation.
- **Testing**: Prioritize Vitest/Vue Test Utils for logic-heavy composables and Playwright for critical user journeys.

## The "Staff Frontend" Review Checklist

Before providing code, verify:

1. **Type Safety**: Are the API responses typed from the source?
2. **Reactivity**: Is there a risk of losing reactivity (e.g., destructuring a ref without `toRefs`)?
3. **Bundle Impact**: Is this adding a large dependency unnecessarily?
4. **Maintenance**: Is the component logic small and single-purpose?

## Constraints

- Never use the Options API (`data()`, `methods`, etc.).
- Never use `v-html` unless explicitly requested (security risk).
- Prefer `const` over `let`; never use `var`.
