# ADR-001: One class-group model drives filtering, ordering and list sections

Status: proposed | Date: 2026-11-20
Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Owner/review: none recorded |
Rules: FE-01, FE-04, NH-04. Relates to [065 ADR-003](../../065-observation-classification/adr-003-classification-viewer.md).

## Context and options

The viewer needs three related behaviours over the same classification data: a filter,
a deterministic row order, and List-view sections. The classification vocabulary is seven
values (`infrastructure`, `personal`, `neighbour`, `transient`, `learning`,
`mobile-unidentifiable`, `likely-neighbour`), presented with five visual tones.

Doing each behaviour separately would create three places that map class to a category and
three places to state an order. They would drift, and the drift would be visible: a row
filtered as "Outside" but listed under a different heading, or a section order that does not
match the preset row.

Options:

1. **Use the seven raw classes everywhere.** Most precise, but seven presets plus "Any" is a
   crowded control, and two pairs are the provisional/settled form of the same idea.
2. **Use the five presentation tones.** Compact, but the tone is a colour, and "Yours" would
   also contain `mobile-unidentifiable`, whose badge reads "Someone's device". The filter
   would then contradict the row's own words.
3. **One six-group table (chosen).** `Yours`, `Someone's device`, `Installed`, `Learning`,
   `Outside`, `Passing`, plus `Any`. Only the two neighbour values merge under `Outside`,
   because both mean "probably not in my home" and both are hidden by default. The table is
   the single source of truth for preset order, sort rank and section order.

## Decision and consequences

Choose **option 3**. The table lives in a new
`frontend/src/domain/observations/classes.ts`, beside — not inside — the existing
`presentation.ts`, because it owns a different concern: classification grouping and ordering,
not badge/format copy. Labels are taken from `CLASS_PRESENTATION` where they already exist,
so a filter chip and a row badge cannot disagree.

Consequences:

- One ordered array (`CLASS_GROUPS`) supplies filter presets, `classGroupOf()` for a row,
  sort rank and List sections. Adding a class means editing one file.
- Every preset is rendered at all times with a fixed-width count slot; a zero count disables
  it. Omitting a group or letting a count gain a digit moved the row while it was being read.
- Ordering is class rank, then display title, then key. Volatile fields (`lastSeen`, RSSI,
  trend, packets) never participate, so snapshots cannot reorder the list.
- An identity with no classification falls into `Learning`, matching the badge fallback.
- Grouping no longer sorts sections by member count, removing the second source of movement.

Verification: unit tests for group mapping, ordering under volatile-only changes, and
fixed section order; browser checks at 1280 px and 390 px.

## Supersession

Supersedes the vendor grouping chosen for the List view in spec 064 (`identityGroup`).
Vendor remains available as a facet and through search. It does not supersede 065 ADR-003's
Tracked/Census split, which remains separate future work.
