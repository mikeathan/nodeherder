# ADR-003: Split the viewer into Tracked and Census

Status: proposed | Date: 2026-11-12
Spec/plan: [spec.md](spec.md), [plan.md](plan.md) | Owner/review: none recorded |
Rules: FE-01, FE-03, FE-04. Relates to
[064 ADR-003](../064-wireless-observations/adr-003-observation-viewer.md).

## Context and options

The viewer was built for a census: one flat table and one radar, with filters. Against real
data that produces roughly forty rows, of which four matter. Adding classification makes the
signal findable, but the presentation still serves the census and buries the tracked set.

Options:

1. **Keep one surface, add a class filter.** Smallest change. But the radar then shows either
   a sparse tracked set or an unreadable census, and the default is a filter state the
   operator must not accidentally lose.
2. **Two surfaces, one switch (chosen).** Tracked is the default and answers "what is
   happening in my home"; Census is deliberate and answers "what can the hub hear".
3. **Separate routes and navigation items.** Two destinations for one dataset. Rejected:
   doubles the navigation surface for a distinction better expressed as a switch.

## Decision and consequences

Choose **option 2**, with the structure:

```
Wireless
├── Tracked (default): Radar · Focus · Timeline
└── Census:            List, grouped by class
```

Plus one new pane, **Timeline**, showing presence per identity per hour across a day. This is
the direct answer to "I cannot find patterns": a radar shows where things are now, a table
shows what exists, and only a timeline shows *when*. Without it the classification improves
the list but still hides routine.

Consequences:

- The radar is only meaningful for a small set, so it belongs to Tracked and gains
  class-coloured blips with a worded legend, plus labels for the few identities present.
- Census keeps the grouped list and the distance bands, grouped by class rather than vendor,
  because class is the question being asked there.
- Two-level control: a Tracked/Census switch, then view tabs within it. One level cannot
  express two different questions.
- A "N hidden" footer appears wherever a filter suppresses rows, so the surface never
  silently misrepresents the household.
- Deliberately preserved from 064 ADR-003: no bearing claim anywhere, no re-use of
  `SpatialCanvas.vue`, and range-only wording on the radar.
- Timeline is horizontally scrollable on mobile with a sticky identity column, not a squeezed
  grid; a grid at 390 px would be unreadable.

Verification: browser checks at 1280 px and 390 px; legend states classes in words so colour
is never the only cue; keyboard reachability of the switch and view tabs recorded.

## Supersession

Extends 064 ADR-003's viewer decision from one surface to two. Its rejection of
`SpatialCanvas.vue` and of bearing claims still stands.
