# Tasks: Device configuration help and defaults

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: complete

## Preparation

- [x] T001 — Resolve the description-source, synthesis and ownership questions in
  `spec.md` (Q1–Q3, A1–A2); IDs: NH-01; done: spec questions replaced with resolutions.
- [x] T002 — Check parent, backend and frontend rules against the design; IDs: NH-01,
  NH-04, BE-01, FE-01; done: `plan.md` Constitution Check.

## Coverage and implementation

- [x] T003 — Add Go regression cases for `object_id` capture, rejection and the written
  attribute in `backend/internal/deviceproviders/esphome/{entity,device_sink}_test.go`;
  IDs: FR-01, AC-03; depends: T002; done: `TestDescribeDiscoveredEntities` and
  `TestDeviceSinkPublishesOnlyBoundedPublicMetadata` cover the new field.
- [x] T004 — Implement `Descriptor.ObjectID`, bounded capture in `Describe`, and the
  `object_id` attribute in `deviceEntity`; IDs: FR-01, AC-03; depends: T003; done:
  `go test ./internal/deviceproviders/esphome/...` passes.
- [x] T005 — Add Jest cases for normalization, precedence, label fallback, device class,
  unknown entities and catalog integrity under
  `frontend/src/__tests__/domain/entityHelp/`; IDs: FR-02, FR-03, FR-05, AC-01…AC-06,
  AC-10, AC-11; depends: T002; done: 18 cases pass.
- [x] T006 — Implement `src/domain/entityHelp/{types,catalog,resolve,index}.ts` and the
  `ExposeAttributeKeys.ObjectID` contract; IDs: FR-02, FR-03, FR-05, NFR-01, NFR-02;
  depends: T005; done: Jest domain suite passes.
- [x] T007 — Render provider help inline and catalog help through the accessible
  disclosure in `ExposeRow.vue`; thread `show-help` through `ProviderEntitySection.vue`
  and `DevicePage.vue`; IDs: FR-04, AC-01, AC-02, AC-05…AC-09; depends: T006; done:
  browser case `device configuration explains known settings without inventing unknown
  ones` passes on desktop and mobile.
- [x] T008 — Update the plan's traceability and Constitution Check with actual results;
  gate: NH-05; depends: T004, T006, T007; done: `plan.md` records commands and outcomes.

## Reset to defaults

- [x] T011 — Extend spec/plan with the reset source, scope and exclusions (Q4–Q7,
  AC-12…AC-18, FR-06…FR-09, NFR-04/05); IDs: NH-01, NH-02; done: artifacts updated before
  implementation.
- [x] T012 — Add Jest cases for default applicability, label fallback, numeric
  range/step, select-option drift, switch/light defaults, bulk filtering and confirmation
  copy; IDs: FR-06, FR-09, AC-14, AC-16, AC-17; done: 10 cases pass.
- [x] T013 — Implement the shared `domain/entityKey.ts` plus
  `domain/entityDefaults/{types,catalog,resolve,index}.ts`; IDs: FR-06, FR-09, NFR-02;
  done: domain suites pass.
- [x] T014 — Add `composables/device/useEntityReset.ts`, the opt-in reset control in
  `ProviderEntitySection.vue`, and wire it from `DevicePage.vue`; the composable confirms
  the named targets and sends through `useDeviceCommand`; IDs: FR-07, FR-08, AC-12,
  AC-13, AC-15; done: build and reset browser case pass.
- [x] T015 — Add the browser case proving two documented defaults are sent, unknown and
  button entities are skipped, and the 375 px layout does not overflow; IDs: AC-13,
  AC-14, AC-16, AC-18; done: desktop and mobile cases pass.

## Handoff

- [x] T009 — Run selected checks: backend `go test ./...`, `go build ./...`,
  `go vet ./...`; frontend `npm test -- --runInBand`, `npm run lint`, `npm run build`,
  `npm run type-check:e2e`; browser `e2e/devices.spec.ts` desktop and mobile; gate:
  NH-05/FE-04; done: 54 Go packages, 144 suites / 802 Jest tests, 11 desktop and 11
  mobile browser tests passed; lint reports 0 errors and 23 pre-existing warnings.
- [x] T016 — Final diff review for stale comments, duplication, unrelated changes and
  affected documentation; gates: NH-04, FE-01; done: review found no unrelated changes;
  help and defaults share one identity normalizer, and reset reuses the existing command
  lifecycle.
