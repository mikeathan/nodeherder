# Tasks: Weather integration

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: in progress

## Preparation

- [ ] T001 — Confirm provider v2 endpoints, fields, limits, licensing, key transport and location selection in `plan.md` and contracts; IDs: FR-01, AC-01–03; done: sourced request/response fixtures and resolved questions.
- [ ] T002 — Define versioned weather resource, freshness, bounded multi-request lifecycle, trust and rollback; record ADR and update `plan.md`; gates: NH-03/04/06, BE-04/05, FE-02/05; depends: T001; done: reviewed contracts and resolved design gates.

## Implementation

- [ ] T003a — Add fixed Met Office header-key auth to shared compiler, request builder, credential lifecycle and frontend definition/setup contracts; IDs: AC-02, NFR-01; depends: documented auth increment in plan; done: origin, missing/invalid key and existing auth regression tests pass.

- [ ] T003 — Implement source-owned backend capability, provider definitions and necessary shared setup/transport/runtime extensions under `backend/internal/integrations/`, `backend/models/integrations/`, `backend/configs/integrations/`; IDs: FR-01–03; depends: T002; done: fake-provider and existing waste regressions pass.
- [ ] T004 — Extend Home settings/API compatibility and source-owned tests under `backend/models/settings/` and `backend/internal/api/`; IDs: AC-04, AC-07; depends: T002; done: old layout and missing-source tests pass.
- [ ] T005 — Implement setup, validated contracts, Connections and responsive Home/results UI in `frontend/src/`, with matching mocks and source-owned tests; IDs: FR-01–03, NFR-02; depends: T002–T004; done: synthetic provider and UI acceptance checks pass.

## Handoff

- [x] T006 — Run backend/frontend checks and affected race and browser tests listed in plan; gate: NH-05; depends: T003–T005; done: actual results recorded.
- [x] T007 — Verify keyboard, stale/error states and 320/375/768/1440 viewport layouts; IDs: AC-04–07; depends: T005; done: recorded evidence.
- [ ] T008 — Update documentation, inspect full diff and complete NH/BE/FE rule review; gates: NH-01/04/05; depends: T006–T007; done: findings and limitations recorded.

## Implementation progress

Shared location-select recipe, credential modes, combined forecast execution, three shipped provider definitions, weather mapping/validation, frontend ingress, setup/catalogue, Home references and responsive result component are implemented. Provider fixtures and application checks are running; completion checkboxes remain open until acceptance and final review finish.

Final verification: full backend tests and targeted race tests passed; backend build/vet, frontend unit/lint/build, and 10 selected browser tests passed. The source graphs were refreshed structurally; semantic refresh and merged graph regeneration remain outstanding. Live provider verification was not performed.

- [x] Support declared UK postcode lookup and contextual location labels.
- [x] Recover manual status checks from owner revision changes without duplicate lookups.
- [ ] Verify provider configuration, regression tests and applicable checks.

- [x] Accept and test Met Office minute-precision timestamps in hourly and daily forecasts.

- [ ] Make Home integration cards readable at desktop and mobile widths and verify placement.
- [x] Select one integration result at a time in Home while retaining device multi-selection.
- [x] Show only the complete bin schedule in the Home picker and retain existing service cards.

- [ ] Make weather and other integration cards readable in panel mode at desktop and mobile widths.

- [ ] Improve the shared device tile and responsive Home grid; verify desktop and mobile layout.

- [x] Preserve full Home block content during drag and provide a clear placement placeholder.
- [ ] Verify desktop and touch previews, persisted reorder, existing keyboard controls, and frontend checks.

Drag-preview verification: frontend unit tests, lint, build and e2e type-check passed. Browser acceptance remains unverified because port 4110 is occupied; do not target the running household application.

- [x] Correct condition mappings, icon presentation and next-hours filtering with regression coverage.
- [x] Correct forecast navigation accessibility, setup name initialization and resource-dispatch readability.
- [ ] Resolve WeatherAPI data retention and disclosure requirements without adding scheduled refreshes.
- [ ] Run review regression checks and document unresolved limitations.

Second review: WeatherAPI retention is not yet enforced; provider remains available but must not be treated as compliant. Existing host provider files require a reviewed shipped update. Browser drag/forecast checks remain unrun because port 4110 is occupied.

- [ ] Future shared integration scheduling: design and implement provider-aware scheduled retention cleanup, including startup recovery after shutdown, database storage and backup retention; verify WeatherAPI's terms. No automatic provider refresh is authorized by this task.

- [x] Clarify shared integration, token and provider-definition removal wording.
- [x] Evict removed integration data and fence stale reads; verify existing Home placeholders and regressions.

- [x] Reuse integration status rendering and show errors consistently on Home and the integrations board.

- [ ] Future shared scheduling work: specify a generic server-owned freshness and retention policy model, trusted provider limits and safe legacy migration.
- [ ] Future shared scheduling work: expose effective data policies in the existing provider UI, allow validated edits within verified limits, and show shared freshness/expiry feedback on all integration views.
- [ ] Future shared scheduling work: design and test category-level expiry, cleanup/restart races, strict offline-retention constraints and database/backup handling before enabling provider-specific retention guarantees.

- [x] Consolidate weather icons and timezone/date formatting into shared owners with regression coverage.
- [x] Review and simplify new frontend composables; preserve lifecycle and presentation behavior.
- [ ] Audit all runtime and operations sources; refactor justified complexity and remove verified dead code.
- [ ] Audit remaining integration packages and consolidate shared URL and network rules without weakening destination security.
- [ ] Review diff and run frontend and backend checks.

Cleanup progress: shared status-check recovery, forecast date/icon formatting, provider response decoding, test credential preparation, publication and cleanup phases were separated. Provider URL validation and origin canonicalization now have a shared owner, with network address enforcement retained in transport. Definition activation delegates instance rebinding; waste normalization separates source details, service identity and date coverage. A complete source-by-source review and remaining complexity analysis are still open. Browser acceptance remains blocked by the occupied fixture port. The WeatherAPI retention limitation remains deferred to the shared scheduling work above.
