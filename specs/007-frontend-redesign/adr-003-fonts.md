# ADR-003: Self-hosted Figtree and Lexend fonts

Status: Proposed (not approved) | Date: 2026-10-10 | Spec: [007](spec.md) FR-02, NFR-03

## Context

The approved Hearth + Panel design uses Figtree for the app and Lexend for panel mode. The
previous UI loaded Roboto from `@fontsource/roboto`. Loading fonts from a CDN at runtime
would add a third-party request on every page load, and would fail on hubs without
internet access.

## Decision

- Replace `@fontsource/roboto` with exact-pinned `@fontsource/figtree@5.3.0` and
  `@fontsource/lexend@5.3.0` (both OFL-licensed).
- Import only the weights in use: Figtree 400–800 and Lexend 300/500.
- Bundle the fonts with the app. The browser downloads the font files only when it renders
  text in them, so this adds font files to the build but no JavaScript.
- Fallback stacks in `tokens.css` keep the UI readable if a font file fails to load.

## Alternatives

| Option | Why not |
| --- | --- |
| Google Fonts CDN | Runtime third-party request; breaks on offline hubs |
| System fonts only | Does not match the approved design |
| Keep Roboto | Does not match the approved design |

## Consequences

The build contains the font files. Measured bundle after the change: gzip JS 560 KB (was
641 KB) and CSS 62 KB (was 66 KB). Rollback: swap the imports in `src/main.ts`.
