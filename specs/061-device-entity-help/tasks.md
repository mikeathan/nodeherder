# Tasks: Entity help for device configuration

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

## Handoff

- [x] T009 — Run selected checks: backend `go test ./...`, `go build ./...`,
  `go vet ./...`; frontend `npm test -- --runInBand`, `npm run lint`, `npm run build`,
  `npm run type-check:e2e`; browser `e2e/devices.spec.ts` desktop and mobile; gate:
  NH-05/FE-04; done: 54 Go packages, 792 Jest tests, 10 desktop and 10 mobile browser
  tests passed; lint reports 0 errors and 23 pre-existing warnings.
- [x] T010 — Final diff review for stale comments, duplication, unrelated changes and
  affected documentation; gates: NH-04, FE-01; done: review found no unrelated changes;
  the only behavioral refactor is `ProviderEntitySection` delegating description
  suppression to `ExposeRow` so there is one resolution path.
