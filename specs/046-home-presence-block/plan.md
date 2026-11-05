# Plan: Home presence block

Spec: [spec.md](spec.md) | Status: Implemented (unreviewed) | Updated: 2026-11-05

## Context and design

Home persists `HomeCard` sources; adding a `presence` kind reuses placement, drag, import/export
and revision fencing. The card reads the derived device from the existing store through
`useDerivedPresence`; no new request or route.

- Backend `models/settings/home.go`: accept `presence` with a device id and no expose/integration
  fields; `internal/api/home.go` admits it when the device exists.
- Frontend types/layout/cards/panel: `HomePresenceSource`, identity as `[kind, device_id]`,
  title "Presence", Panel projection skips it.
- Picker: one Presence choice when `supportsSpatialMapping(device)`.
- `HomePresenceCard.vue`: status, derived occupied/count/last zone/empty, zone chips from
  `zoneId`/`zoneName` attributes, explicit missing/underived states, link to the device.
- `HomeCardItem`/`GroupDashboard`: presence branch and responsive span.

## Constitution Check

| Rule | Evidence | Before |
| --- | --- | --- |
| NH-01 | Spec 046 | PASS |
| NH-02 | Derived values only; unknown never shown as clear | PASS |
| NH-03 | No new route/data exposure; authenticated Home layout | PASS |
| NH-04 | Additive persisted kind, server validation, existing card mechanics | PASS |
| NH-05 | Go validation test; frontend domain tests; browser specs | PASS |
| BE-01/04 | Validation in settings; projection unchanged | PASS |
| FE-01/03 | Domain decision; explicit missing states | PASS |

## Traceability

| IDs | Paths | Procedure | Result |
| --- | --- | --- | --- |
| FR-02, AC-07 | backend home settings/API | `home_test.go` accepted/rejected presence references | pass (API admission code-reviewed, not directly tested) |
| FR-03, AC-01 | `picker.ts` | `picker.test.ts` presence choice only for spatial devices | pass |
| AC-02–AC-05 | `HomePresenceCard.vue`, `HomeCardItem.vue` | type-check/build; card is not browser-driven yet | pass (type) |
| AC-06 | `domain/home/panel.ts` | `panel.test.ts` skips presence | pass |
| AC-08, SC-02 | all | `npm test` (744), build, lint, type-check, `home.spec.ts` + `devices.spec.ts` (38 passed) | pass |

## Completion

Implemented and checked for placement, persistence, identity, admission and rendering. The card
itself is type-checked and its domain decisions tested, but it is not yet browser-driven with a
spatial device; no real-device presence values have been inspected.
