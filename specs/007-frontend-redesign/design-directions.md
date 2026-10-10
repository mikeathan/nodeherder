# Design directions (round 2)

Part of [plan 007](plan.md). Owner feedback on round 1 (2026-10-10): "I like Hearth, but the rest
are the same layout with different themes. I want a few completely different designs."
Round 2 keeps **Hearth** and replaces the four reskins with four directions that differ in
information architecture, navigation, page structure and editor interaction, not only colour.
Method: the `frontend-design` skill (plan tokens, check against generic defaults, then build) and
the Vercel Web Interface Guidelines as the review checklist.

Subject: a self-hosted hub for one household's Zigbee and Wi-Fi devices. Audience: the owner
(technical, configures automations) and the family (switches lights, checks rooms). Primary
job: see the state of the house and change it safely.

| # | Direction | The one bold idea | Navigation | Automation editor |
| --- | --- | --- | --- | --- |
| 1 | Hearth (kept) | Areas of soft tiles | Sidebar | When/If/Then flow cards |
| 2 | Floorplan | The house plan is the home page | Top tabs, room inspector sheet | Step-by-step wizard |
| 3 | Workbench | IDE: explorer tree, open tabs, inspector | Activity bar + tree + status bar | Outline tree + property inspector + live JSON |
| 4 | Brief | The house written as sentences; command palette | Command bar (Ctrl K) + text menu | Fill-in-the-blanks sentence |
| 5 | Deck | Wall-tablet panel readable from across the room | Room pager + bottom dock | Node canvas with inspector drawer |

## 2 · Floorplan

- Colour: Drafting paper `#F5F8FC`, Ink `#1D2B45`, Plan line `#8FA6C6`, Room fill `#E6EDF6`,
  Live yellow `#F2C230` (only for things that are on), Alert `#D2463D`. Dark: night plan `#102033`.
- Type: Barlow Condensed for room names (architectural signage), Barlow for everything else.
- Layout: plan canvas (rooms as proportional rectangles, garden outside the walls) with
  device pins; selecting a room opens an inspector sheet with its controls.

```text
[logo] Plan  Devices  Rules  System              (status)
┌───────────────────────────────┐ ┌──────────────┐
│ Attic room     │ Bedroom      │ │ Living room  │
├────────────────┼──────────────┤ │ 21.2 °C      │
│ Living room ●  │ Kitchen  ●   │ │ Light   [on] │
├────────┬───────┴──────────────┤ │ Presence     │
│ Garage │   (garden, outside)  │ │ …            │
└────────┴──────────────────────┘ └──────────────┘
```
- Principles: space before lists; yellow means "on" and nothing else; the wizard numbers steps
  because they are a real sequence.

## 3 · Workbench

- Colour (dark default): Graphite `#22262D`, Panel `#2A2F37`, Rule `#3A414C`, Text `#D9DEE5`,
  Signal blue `#5AA9E6`, Amber `#E0A33B`. Light: `#F4F5F7` panels on `#FFFFFF`.
- Type: IBM Plex Sans; IBM Plex Mono only where the content is code (IEEE addresses, payload JSON).
- Layout: activity bar, explorer tree (areas → devices → exposes; automations → triggers),
  tabbed main area, property grid, status bar with connection and permit-join, collapsible log panel.

```text
┌─┬────────────┬──────────────────────────────┬─────────┐
│▣│ EXPLORER   │ [Overview][Attic Light ×]    │         │
│▤│ ▾ Attic    │ property │ value │ control   │ details │
│⚙│   ▸ Light  │ state    │ ON    │ [toggle]  │         │
│ │ ▾ Rules    ├──────────────────────────────┤         │
│ │            │ log: 18:02 info device …     │         │
├─┴────────────┴──────────────────────────────┴─────────┤
│ ● connected   17 devices   join closed   MCP 1 client │
└────────────────────────────────────────────────────────┘
```
- Principles: density with alignment; every value in a grid row with its type and access
  mode; nothing hides behind hover.

## 4 · Brief

- Colour: Cool paper `#FAFBFC`, Ink `#1F2430`, Graphite `#5E6675`, Marker `#FFE79A` (highlights
  the parts of a sentence you can change), Teal `#0F7E6E`, Rust `#B4482E` for problems.
- Type: Atkinson Hyperlegible throughout (designed for low-vision legibility; fits family use).
  One size step larger than usual for body text.
- Layout: one centred column (≤ 44rem). Home is a written status ("16 of 17 devices are online…")
  with inline links and switches; a command bar finds any device, room or action.

```text
            [ Search or type a command…   Ctrl K ]
   Good evening. The house is quiet.
   16 of 17 devices are online. Bedroom radiator valve went
   offline 2 hours ago. The attic is humid (84 %).
   Lights on: Attic room Light [●], Living Room Light [●]
   ─ Today ─────────────────────────────────────────
   18:02  Front door closed
```
- Principles: words are the interface; marker highlight only on editable tokens; no cards.

## 5 · Deck

- Colour (dark default): Lagoon night `#13252C`, Surface `#1B343D`, Raised `#24444F`,
  Text `#EAF4F4`, Lamp `#FFB547` (light), Climate `#6FD3E8`. Light: fog `#E6EEF0`.
- Type: Lexend (designed for reading at a distance); numerals very large.
- Layout: clock and outside temperature strip; one room per page with a pager; big widgets
  (temperature dial, light buttons with brightness arc); dock at the bottom.

```text
 18:42  Friday 10 Oct           Garden 12.2 °C      ● hub
 ‹ Attic room ›  Kitchen  Living room  Garden  Garage
 ┌──────────────┐ ┌──────┐ ┌──────┐ ┌──────┐
 │   ◯ 19.8°    │ │ 💡   │ │ 🔌   │ │ 🔔   │
 │  humidity 84 │ │Light │ │Socket│ │Alarm │
 └──────────────┘ └──────┘ └──────┘ └──────┘
 [ Rooms ] [ Devices ] [ Rules ] [ System ]
```
- Principles: every target ≥ 64 px; one glance per page; the dial is the signature element.

## Review against generic defaults (done before building)

- Brief first used a cream background with a serif display. That is the most common generated
  look, so it was changed to cool paper and a single legibility-first sans.
- Deck first used near-black with an acid accent, another common default. It was changed to a
  blue-green night and lamp amber, because lamp colour carries meaning here.
- Workbench uses mono only for real code values, not as decoration for labels.
- No direction uses ALL-CAPS eyebrow labels or "A · B" meta strings. Numbered markers appear
  only in the Floorplan wizard, where the steps are a real sequence.
