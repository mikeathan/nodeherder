# ADR-001: Runtime design tokens over PrimeVue styled mode

Status: Proposed (not approved) | Date: 2026-10-09 | Spec: [007](spec.md)

## Context

The redesign needs switchable designs, light/dark, user accent, density, radius and
font scale at runtime (FR-02). Today a static Material preset plus hard-coded colours
makes that impossible. FE-01 requires an ADR to replace the component system; FE-04
requires shared controls and theme tokens.

## Decision

Keep PrimeVue 4 (styled mode) for behaviour/accessibility. Introduce `--nh-*` CSS custom
properties as the single source of visual truth; a minimal `definePreset(Aura, …)` maps
PrimeVue semantic tokens (`primary`, `surface`, `content`, `formField`, …) to
`var(--nh-*)`. Designs are token files + shell variant selected by
`<html data-design>`; user knobs set inline custom properties. No new runtime
dependency; accent scale generation is a ~40-line utility.

## Alternatives

1. New UI kit (Vuetify/Quasar/Naive) — migration of every screen, FE-01 replacement, no
   need demonstrated.
2. Tailwind + PrimeVue unstyled — full control, but restyling all PrimeVue widgets and a
   new build dependency; reconsider if token mapping cannot reach a chosen skin.
3. Multiple static PrimeVue presets compiled per design — no custom accent/density at
   runtime; larger CSS.

## Consequences

+ Runtime customisation, one place for colours, retro skins as CSS only.
− Some directions (Floorplan, Deck) need component-level overrides via `pt` classes, and
  each layout direction is a separate shell/route component set, not only tokens.
− Two token systems coexist (PrimeVue's and `--nh-*`) — mitigated by mapping one way only.
Operational cost: none at runtime; bundle expected smaller after removing the 660-line
preset (verify NFR-03).
