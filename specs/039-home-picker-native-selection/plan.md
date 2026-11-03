# Plan: Home picker native selection under live reports

Spec: [spec.md](spec.md) | Updated: 2026-11-03 | Constitution: parent/frontend 1.0.0 (proposed)

## Context and design

The native `<select>` is the shared control (`components/ui/UiSelect.vue`). macOS menus can
dispatch `input` before `change`; between them, a hub report re-renders the picker and Vue's prop
patch re-applies the select `value` from the unchanged model. The commit path had no ownership of
the pending choice, and the Home picker's `options` was rebuilt from the live device array on every
report, so the menu was also rebuilt mid-interaction.

Two small, complementary changes, mirroring spec 038 AC-22's spatial fix:

- `UiSelect.vue` — `@input` commits the option immediately; the paired `@change` is consumed
  without re-reading the DOM so a restored value cannot be committed and a device command is not
  sent twice. Only `change` commits when no `input` was seen.
- `domain/home/picker.ts` — `stableHomePickerOptions(previous, devices)` and
  `stableHomePickerFilters(previous, choices)` reuse the previous array when the choosable values
  and labels are unchanged, and return a fresh list when a device, name or group really changes.
  `AddTilesDialog.vue` uses them through `computed` previous-value getters, the same pattern as
  `stableSpatialInventory`.

The choice stays a native select: keyboard, typeahead and accessible naming are unchanged, so the
fix is compatible with FE-04 and adds no dependency.

## Alternatives

| Option | Rejected because |
| --- | --- |
| Freeze the device list when the dialog opens | Hides genuinely new devices for the dialog's lifetime; the spatial fix reuses identity instead |
| Debounce or pause hub reports while a dialog is open | Suppresses live data and does not fix other `UiSelect` users |
| Replace the select with a custom listbox | Broader redesign than the defect; loses native mobile/keyboard behavior |
| Commit only on `change` with a retry | Race remains; retries could duplicate a device command |

## Constitution Check

| Rule | Evidence | Result |
| --- | --- | --- |
| NH-01 | Spec 039 with reproduction and acceptance | PASS |
| NH-02 | No device pipeline, command routing or automation change | N/A |
| NH-03 | No entry point, secret or household-data exposure added | N/A |
| NH-04 | Shared control fixed at its owner; no competing state | PASS |
| NH-05 | Domain regression plus browser regression; suite/lint/build recorded | PASS |
| NH-06 | Choice is committed and shown; no inferred data or silent mutation | PASS |
| FE-01 | Presentation stays in components; projections stay in the picker domain | PASS |
| FE-02 | No API/WS contract change; existing device payloads only | N/A |
| FE-03 | Paired `change` is ignored after `input`, so a command is not duplicated | PASS |
| FE-04 | Native select keeps label, keyboard, typeahead and disabled behavior | PASS |
| FE-05 | No secret or household data added to storage/logs | N/A |

## Traceability

| Requirement | Paths | Test/procedure | Actual result |
| --- | --- | --- | --- |
| FR-01, AC-01/04 | `components/ui/UiSelect.vue` | `home.spec.ts` pending-source regression; affected device-control suites | see tasks |
| FR-02, AC-01–03 | `domain/home/picker.ts`, `components/home/AddTilesDialog.vue` | `picker.test.ts` identity tests | see tasks |
| FR-03, AC-03 | same | picker identity tests plus existing browser suites | see tasks |
