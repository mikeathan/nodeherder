# Spec: Home picker native selection under live reports

ID: 039 | Status: Implemented; verification recorded in [tasks.md](tasks.md) | Updated: 2026-11-03
Constitution: proposed 1.0.0 | Components: frontend

## Problem, scope, exclusions

Choosing an entry in Home's **Add tiles** source menu does nothing when a device report
arrives while the native menu is committing. Reproduction (deployed and local):

1. Open **Add tiles** in edit mode and open the source menu.
2. The native menu changes the select value and fires `input`; `change` follows.
3. A hub `deviceUpdated` report re-renders `UiSelect` before `change`; Vue re-applies the
   unchanged model value, restoring "Choose a source".
4. `change` then reads the restored value, so no option is committed.

Same class as spec 038 AC-22 for the spatial mapping selectors: a native dropdown's pending
choice must survive reports before its change event. `UiSelect` listened only to `change`, and
`AddTilesDialog` rebuilt its option list from the live device array on every report.

Scope: the shared native `UiSelect` commit path, the Home picker's source/filter option
projections, and regression coverage. Exclusions: no backend, contract, routing or layout
change; no redesign of the dialog; handler order or option types stay unchanged.

## Scenarios and acceptance

- AC-01: Given a live-updating hub, choosing a source commits on the menu's first event and the
  select shows that source after later reports.
- AC-02: Given the choice committed, the group or result list for that source is shown immediately
  and is not cleared by a report in the same interaction.
- AC-03: Devices or groups that actually change (added, removed, renamed, new value group) still
  refresh the menus; only reading-only reports reuse the previous selectors.
- AC-04: One native choice emits one model update; the paired `change` must not commit a command
  a second time. Typed option values (numbers, booleans, strings) and disabled options are
  unchanged.

## Requirements

- FR-01: `UiSelect` MUST commit the selected option on `input`, and the immediately paired
  `change` MUST NOT read a value that a render may have restored (AC-01, AC-04).
- FR-02: The Home source and filter option lists MUST reuse their previous array identity while
  the choosable entries are unchanged (AC-01–AC-03).
- FR-03: The projections MUST remain pure and derived; no stored layout, card or integration data
  changes (AC-03).

## Success

- SC-01: Jest, lint and build pass; the new browser regression passes on desktop and mobile.
