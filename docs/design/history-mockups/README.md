# History, chart and mobile design samples

Static HTML prototypes for the history/chart rework (spec 007 follow-up, tasks T209–T212) and
the mobile shell. **Not application code**: nothing here is imported, built, bundled or served
by the app. No build step, no dependencies, no network access.

## Open them

```bash
open docs/design/history-mockups/index.html          # macOS
xdg-open docs/design/history-mockups/index.html      # Linux
```

Start at `index.html` for the gallery, or open a sample directly. Each page has a top bar to
jump between samples and to switch theme (Fleece / Daylight) and mode (dark / light); the
choice is remembered per browser.

## Samples

| File | Explores | Main decision it informs |
| --- | --- | --- |
| `v1-history-chips.html` | Range chips (incl. Today/Yesterday), entity chips with sensor colours, full-width 360 px chart, hover readout, min/avg/max/last strip, and zoom that works by button, pinch, drag, wheel or keyboard | Chips vs today's `<select>`; whether history gets its own nav entry |
| `v2-multi-compare.html` | Two series in one window, legend toggles, previous-day ghost, comfort band, automation marker | Whether to support mixed units (second y-axis) or same-unit comparisons only |
| `v3-zoom-study.html` | Zoom −/+/Reset, drag-to-zoom, wheel zoom, double-click reset, expand to 540 px, line/step/bars | Whether the zoom plugin (~8 kB) is worth it versus preset ranges |
| `v4-statistics.html` | History vs Statistics modes, daily aggregate bars, min/avg/max/samples table, CSV export | One page with a mode switch versus a separate Statistics page |
| `v5-mobile-sheet.html` | Phone sheet with a 170 px preview and a primary “View full history”, then the full screen at 220 px with pinch/button zoom | Where history is entered from on a phone |
| `v6-mobile-nav.html` | Five-tab bottom navigation, scrollable chip rows, pull-to-refresh, safe-area padding | Bottom tabs vs the current hamburger drawer |

## Fidelity

- Colours, radii and spacing come from the real tokens in
  `frontend/src/assets/styles/{tokens,themes}.css`; Fleece is the shipped default theme.
- Fonts fall back to the system UI stack because the `@fontsource` files are not copied here.
  Expect slightly different text metrics from the app.
- All data is synthetic and deterministic (seeded), generated at page load. No household data,
  no network calls, no device commands.
- The chart is a small hand-written SVG renderer, not Chart.js. It exists to test layout,
  density and interaction affordances — not rendering performance or plugin behaviour.
- Phone frames are fixed 390 × 780 with a fake status bar. They are illustrations, not
  device emulation; resize tests belong in the real app. `v1` is responsive, so it can also be
  opened at a real phone width (`390 × 844`) to test the touch controls directly.

## Chart gestures (all samples)

| Input | Action |
| --- | --- |
| Mouse drag across the plot | Select a range and zoom to it |
| Mouse wheel | Zoom around the pointer |
| Double-click | Reset to the selected preset window |
| Two-finger pinch (touch) | Zoom around the pinch midpoint |
| One-finger drag (touch) | Move the readout/crosshair |
| `+` / `-` / `0` keys | Zoom in, zoom out, reset (chart must have focus) |
| `←` / `→` keys | Pan within the loaded window |
| `−` `+` `Reset` buttons | Same actions, 44 px targets on phones |

Every gesture has a button equivalent, so nothing depends on a trackpad, a mouse or a
precise tap.

## What is deliberately not built

Tablet split view (list + detail), PWA install prompt and offline shell, chart screenshot
sharing, and a cross-device comparison page. They are listed in `index.html` as open options.

## After a direction is chosen

Create the implementation spec (for example `specs/029-history-and-mobile-ux/`) that absorbs
T209–T212 plus the mobile shell work, and retire the stale root `charts_todo.md`, whose
remaining items are now covered by that plan. These mockups are evidence for that spec, not a
replacement for it.
