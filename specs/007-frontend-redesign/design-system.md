# Design system: tokens, customisation, components, states

Part of [plan 007](plan.md). Proposal; the chosen sample (Q-01) fixes default values.

## Principles

1. **Glanceable first** — value and state readable in < 1 s; secondary data (LQI,
   battery, last seen) as compact chips, never hidden in footers.
2. **Retro by skin, modern by behaviour** — character comes from type, palette, chrome
   and texture; interaction patterns stay conventional and accessible.
3. **One token source** — no hex colours in components; everything resolves to `--nh-*`.
4. **Honest state** — requested ≠ confirmed (NH-02); offline/disabled/stale always
   labelled with icon + text.
5. **Customisable, not configurable-to-death** — five knobs + presets; advanced users get
   a custom accent colour.

## Token tiers

```css
/* 1. primitives (per design) */
--nh-p-neutral-0 … --nh-p-neutral-1000; --nh-p-accent-50 … 900 (generated from accent)
/* 2. semantic roles (all components use only these) */
--nh-bg            /* page */            --nh-surface      /* cards, panels */
--nh-surface-2     /* raised/inset */    --nh-border       --nh-border-strong
--nh-text          --nh-text-muted       --nh-text-inverse
--nh-accent        --nh-accent-contrast  --nh-accent-soft  /* tinted bg */
--nh-ok  --nh-warn  --nh-danger  --nh-info              /* status, each + -soft */
--nh-on            /* "device on" glow */ --nh-off
--nh-focus         /* focus ring colour */
/* 3. shape, type, density, motion */
--nh-radius-s/m/l  --nh-border-w  --nh-shadow-1/2  --nh-font-ui  --nh-font-mono
--nh-font-display  --nh-fs-base (scaled by fontScale)  --nh-space (density unit)
--nh-tile-h        --nh-motion (0ms when prefers-reduced-motion)
```

Domain tokens map sensor kinds to roles so icon colours stay consistent across skins:
`--nh-kind-temperature`, `-humidity`, `-energy`, `-light`, `-motion`, `-air`, `-alarm`,
`-contact` (each skin sets them; default derived from status/accent palette).

## Customisation model (`nodeherder_theme` v1)

| Setting | Values | Effect |
| --- | --- | --- |
| `design` | `hearth` `floorplan` `workbench` `brief` `deck` (+ `legacy` during rollout) | Token set + shell and page layouts |
| `mode` | `system` `light` `dark` | `data-mode`; system follows `prefers-color-scheme` |
| `accent` | preset swatch or `#rrggbb` | Generates accent scale; contrast-checked, auto picks `--nh-accent-contrast` |
| `density` | `compact` `comfortable` `spacious` | `--nh-space` 0.75/1/1.25 and `--nh-tile-h` |
| `radius` | `0`…`20` px | `--nh-radius-*` |
| `fontScale` | 90 %…125 % | `--nh-fs-base` |
| `preset` | per-design named palettes (e.g. Deck: lagoon, ember, frost) | Overrides primitives |
| `effects` | on/off | Retro texture (scanlines, pinstripes, grain); forced off for reduced motion/contrast |

Rules: invalid stored values fall back silently to defaults; storage contains no
household data (FE-05); an "Appearance" settings tab and a quick popover in the header
expose the same model; "Reset to default" restores the design's defaults.

## Components (new `components/ui/`)

| Primitive | Used by | Notes |
| --- | --- | --- |
| `NhTile` | Home areas, overview | icon, label, value+unit, state, pending spinner, offline/disabled badge, optional sparkline; click = toggle when actionable, otherwise open details; `Enter/Space` parity |
| `NhCard` / `NhSection` | all | header (title, actions), body, footer chips |
| `NhStat` | overview | big number, label, trend/threshold, link to filtered list |
| `NhBadge` / `NhChip` | lists | protocol, power, LQI, battery, availability (icon + text) |
| `NhToggle`, `NhSegmented`, `NhSlider`, `NhSelect` | exposes, editor | wrap PrimeVue ToggleSwitch/SelectButton/Slider/Select with tokens |
| `NhSentence` | editor | renders trigger summary as text tokens, no HTML injection |
| `NhTimeline24` | scheduler | 24 h bar with enable/disable markers |
| `NhEmpty`, `NhError`, `NhSkeleton` | all | standard states |
| `NhStatusDot` | header, lists | dot + text label |

## States

Every screen specifies: **loading** (skeletons, no layout jump), **empty** (explain + next
action), **error** (message from `operationFailed` / HTTP error mapped to plain language,
retry for reads only), **offline/reconnecting** (header banner "Reconnecting… (attempt n)";
device controls disabled with tooltip "Hub connection lost"), **stale** (value older than
device's expected interval shown muted with "last seen"), **pending** (spinner/ghost value
on the control until confirmed or 5 s timeout).

## Layout and breakpoints

360 (phone: bottom dock or drawer; tiles 2-up), 768 (tablet: rail; tiles 3–4-up), 1280
(desktop: sidebar; areas in masonry columns), 1920 (wide: 4–6 area columns). Area
columns use CSS grid `repeat(auto-fill, minmax(var(--nh-area-w), 1fr))`.

## Accessibility

AA contrast per preset (scripted check), focus ring `2px var(--nh-focus)` offset 2px,
44 px minimum touch targets in comfortable density (36 px compact), `aria-live=polite`
for toasts and pending → confirmed transitions, all icon buttons labelled.
