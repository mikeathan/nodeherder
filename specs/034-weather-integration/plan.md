# Plan: Weather integration

Spec: [spec.md](spec.md) | Status: Implemented; final verification in progress | Author: NodeHerder contributors | Updated: 2026-10-28
Constitutions: parent/backend/frontend 1.0.0 (proposed)

## Context and implemented design

Weather extends the existing Connections lifecycle. The composition root installs waste and weather adapters; runtime owns encrypted persistence, admitted operations, revision fencing and publication. No scheduler or device-control dependency is introduced. Each shipped provider has a host-editable YAML file under `backend/configs/integrations/weather/`, preserving the definition compiler, host-file ownership, reload and recovery pattern used by WhenIsBins.

The `location_select` flow searches for and validates explicit coordinate choices before admitting a synchronous forecast request. Search/select labels are private flow metadata, preserved for resource presentation without entering shareable YAML. Met Office supports GB locations initially; Open-Meteo/GeoNames location search is disclosed before submission. WeatherAPI.com uses its own search. No-key setup hides credentials; keyed providers link to account setup. See [provider research](provider-research.md).

`operations.ParseDestination` is the shared URL boundary. Request construction keeps attempts immutable and explicitly controls credential forwarding. Fixed header/query modes use existing origin-bound slots; cross-origin public search gets no credentials. The HTTP client retains DNS/private-destination protections, timeout/body limits and redirect refusal. A compiled combination can contain exactly two non-conditional, same-origin GETs. Validation precedes I/O; execution is sequential and non-recursive inside the existing operation deadline. The second failure fails the operation, preserving last-good data. See [ADR 001](adr-001-weather-execution.md).

Weather normalization is split into mapping validation, row/column extraction, scalar parsing and typed forecast projection in `backend/internal/integrations/capabilities/weather/`. `weather_forecast/v1` retains optional quantities, explicit units/timezone, ordered bounded forecasts, stable identity/revision, provider attribution and independent capture/check times. Missing data never becomes zero. Data becomes stale after one hour; current conditions are omitted when outside their validity window. Open-Meteo daily dates respect its declared fixed offset across daylight-saving changes.

## Generic consumer boundaries

Home's API compares opaque `SourceReference` descriptors discovered through `capabilities.SourceProvider`. Each capability validates/decodes its own payload and exposes valid source identities. Home settings check structure, while new placement admission checks capability source membership; accepted missing references remain placeholders. No weather/waste schema checks remain in generic Home code.

Frontend presenters under `frontend/src/integrations/` own setup copy, source references, source/picker presentation, icons, Connections widgets and results components. A composition registry installs them. Generic Home/Connections render those interfaces; capability-specific wrappers narrow resources before calling waste/weather components. Home layout preserves the existing placement/reference scheme and Panel behavior. Mobile daily forecasts show five rows per page; desktop cards scroll. Available fields follow provider availability.

Setup methods orchestrate one existing repository transaction and lock. Focused helpers perform capacity checks, target/replay/selection validation, instance creation and operation admission. AST cyclomatic measurement against main: SaveSetup 37 → 7; CreateSetup 17 → 3; SearchSetup 13 → 5; makeOperation 12 → 7. Largest extracted setup helper: 11. No exception is requested.

## Contracts, persistence and recovery

- Existing authenticated integrations endpoints remain the transport; no public route, MQTT, MCP or metrics contract is added.
- `InstanceDTO.sources` is additive and contains opaque kind/resource/sub-source references. Resource payloads retain the existing flat envelope and versioned common header. The frontend validates weather before shared-state ingress.
- Existing waste definitions and saved resources remain readable. The optional selected-label checkpoint is additive; older records fall back to coordinates when a provider omits a label.
- Weather Home references use the existing layout version. Before rollback to an older binary, back up Home settings and restore a pre-weather layout if necessary; old binaries reject unknown reference kinds rather than silently deleting data.
- Definition changes continue through existing version/hash/destination review and previews. Weather-specific selectors/recipes are editable in YAML mode; arbitrary custom-provider tooling is outside scope.

## Constitution check

Design checks below reflect self-review, not maintainer ratification. Final execution evidence is recorded under verification.

| Rule | Evidence | Design review |
| --- | --- | --- |
| NH-01 | Spec, tasks, provider contracts, ADR; user authorized implementation and corrections | PASS |
| NH-02 | Information only; no device/event pipeline changes | N/A |
| NH-03 | Existing auth; validated payloads, private key slots, origin isolation and safe failures | PASS |
| NH-04 | Capability/source/presenter interfaces; shared lifecycle and documented additive contracts | PASS |
| NH-05 | Isolated fixtures, acceptance tests, frontend and backend checks below | PASS |
| NH-06 | Atomic last-good publication, bounded calls, missing/freshness semantics, rollback guidance | PASS |
| BE-01 | Domain adapters do no I/O/storage; existing repository boundaries | PASS |
| BE-02 | Device lifetime unchanged | N/A |
| BE-03 | Existing worker/context/shutdown ownership; sequential companion requests; race checks | PASS |
| BE-04 | Additive flow metadata/resources; compatibility and restart/recovery regression suites | PASS |
| BE-05 | Existing authenticated routes; safe provider errors; no assistant changes | PASS |
| FE-01 | Existing Vuex/services; presenters separate domain mapping from generic components | PASS |
| FE-02 | Validated ingress, shared fixture and mock-provider browser flow | PASS |
| FE-03 | Loading/failure/stale/missing handling; no automatic provider refresh | PASS |
| FE-04 | Shared controls/theme, focus/keyboard pagination and responsive viewport checks | PASS |
| FE-05 | Credentials remain outside YAML, bundles and presentation resources | PASS |

## Traceability and verification

| Acceptance | Evidence | Result |
| --- | --- | --- |
| AC-01–03 | Three compiled YAML providers; native fake-provider setup; no-key/key-link browser cases; origin/key/recipe rejection tests | Passed isolated checks |
| AC-04–07 | Weather/source ingress tests, opaque Home source test, source identity/placeholder tests, mobile pagination and desktop/browser checks | Passed isolated checks |
| AC-07–08 | Manual setup/refresh fixture, read-without-I/O assertion, failed-refresh last-good preservation, companion request failures | Passed isolated checks |
| NFR-01–02 | Compiler/transport validation, finite/bounded model tests, unit/timezone checks, 320/375/768/1440 viewport checks | Passed isolated checks |

Actual runs: full backend `go test -count=1 -timeout=5m ./...` passed before the final review refactor. Subsequent targeted integration/model/Home tests passed after the refactor. Frontend 85 suites/486 tests, lint (28 existing warnings), TypeScript/Vite build passed. Ten desktop/mobile browser cases covering weather setup/results/key links and existing bin navigation/Home placement passed; the first browser fixture-envelope/label failures were corrected and rerun. A sandbox port/cache denial required approved reruns. Final backend/race/frontend/browser reruns are recorded below when complete.

Tests use synthetic responses, temporary encrypted stores and in-memory browser hubs. No household device or real provider account was used. Live account/endpoint verification remains unrun; actual free limits and service availability remain provider-controlled. Met Office is initially UK-only and API fields absent in a response remain unavailable. These are documented product limits, not fabricated passed checks.

## Completion

Documentation: [Connections guide](../../docs/connections.md), [spec](spec.md), [tasks](tasks.md), [provider research](provider-research.md), [ADR](adr-001-weather-execution.md). Final diff review includes generic ownership, readable declarations, transaction semantics and imports. No exception or external reviewer approval is claimed. No new dependency is added.

Final reruns after the capability and readability refactor: backend full `go test -count=1 -timeout=5m ./...` passed; affected integration/model/Home/API race tests passed; `go build ./...` and `go vet ./...` passed. Frontend unit tests (85 suites, 486 tests), lint (28 existing warnings) and build passed. Ten selected desktop/mobile browser cases passed after the refactor. Source graphs were refreshed for code; semantic extraction for changed YAML/documents and the merged graph refresh remain outstanding. Live provider access remains unverified.

### Location and operation recovery

Declare an optional postcode search endpoint and response mapping in the location-select recipe; select it for UK postcodes and retain the public cross-origin credential boundary. Add optional geographic context to location labels. Re-read the operation owner before a manual status check and re-read on a revision conflict, preserving operation fencing. Cover postcode routing, location labels and concurrent completion with isolated fixtures.

Handle ISO timestamps with minute precision and an explicit timezone in the weather timestamp parser, preserving strict validation and the existing configured timestamp mapping. Exercise real-shaped hourly and daily Met Office fixtures and malformed timestamps.

Use the existing Home grid's two-column span for standalone integration cards, without capability-specific layout branches. Integration placements inside areas also span the area grid so their text remains readable. Verify alongside device tiles at desktop and mobile widths.
Use a single-select control for integration results in the existing Home picker, retaining multi-select for device values.
Let the waste presenter mark individual services as unavailable for new picker placements while retaining their presentations for existing saved cards.

Apply the existing bounded, full-row panel width to every integration tile, independent of capability and summary rows; preserve the device grid and fit narrow viewports.

Increase the Home grid minimum column width and make area columns responsive. Update the shared EntityTile to wrap labels and device names, allow height to grow and keep controls and status readable. Preserve Home placement and panel behavior.

### Home drag previews

Remove the Home block-drag rule that hides all card grids. Keep Sortable's existing interaction copies and revision-fenced saves; style its existing ghost as a visible drop placeholder without changing placement logic. Extend the isolated browser regression for area dragging to check that card content remains visible before release and that reordering persists at desktop and mobile widths. Verify lint, build and frontend tests; do not use household devices.

### Review corrections

Correct provider code maps and test documented cases. Pass condition-specific icons through the capability-owned presentation. Select future hourly entries, make navigation names accessible, guard setup defaults and expand resource dispatch. Investigate WeatherAPI retention using the existing cleanup lifecycle and record any limitation for stopped installations, retained database pages and backups before asserting compliance. Add focused source-owned tests and run applicable checks.

### Future integration scheduling and retention cleanup

When shared integration scheduling is designed, include scheduled cleanup of stored provider data according to each provider's retention rules, including WeatherAPI.com's current 24-hour forecast limit. Cover cleanup after a restart or prolonged shutdown, retained database pages and backups, and behavior when a provider subscription ends. Cleanup is separate from scheduled provider refresh and must not trigger network requests. The existing persistent snapshot does not currently enforce the WeatherAPI limit; this remains unresolved until the follow-up is implemented and verified.

Update shared integration confirmations and unavailable Home text. Evict deleted instances from the session cache and invalidate older in-flight instance reads so they cannot restore removed data. Keep persisted Home placements and provider definition deletion rules unchanged. Verify cache fencing and shared behavior with isolated tests.

Share an integration status component between Home and board cards, passing explicit error and warning state from the owning presenter. Retain each view's existing inline or stacked layout and neutral checked timestamp.

### Future shared data-policy design and UI

The later integration scheduling work must include a provider-agnostic data-policy service, with a single authoritative backend evaluation for freshness and retention. Provider definitions declare policy defaults and supported data categories; a user's configured preferences may shorten retention but must not extend independently verified provider limits. Do not treat arbitrary editable YAML as an immutable provider maximum: design how trusted provider constraints are sourced and validated, and explicitly handle unknown or changed terms.

The existing provider configuration UI must display the effective policy, its source, and any verified limit; where permitted it must edit freshness and retention through the same validated definition model used by YAML, not a separate frontend setting. Explain the effects of changes on already-saved results before applying them. Fresh, stale, unavailable and expired states and timestamps must be computed server-side and consumed consistently by the board, Home, panel and detail views. Preserve existing waste source verification semantics and avoid resetting content age when a 304 response or settings change occurs. For legacy snapshots, define a conservative, testable migration and explicit behavior when a trusted source timestamp is absent.

Retention applies to independently expiring data categories where required; a provider with differing current and forecast limits must not be modeled as a single arbitrary expiry unless the strictest limit is applied. Define atomic cleanup, instance revision and Home placeholder behavior, active-operation races, restart and failed-cleanup recovery. Periodic and startup cleanup must not submit provider requests or silently create a refresh schedule. Startup cleanup cannot enforce a storage deadline while NodeHerder is stopped; bbolt free pages and backups also outlive logical deletion. The scheduling plan must therefore evaluate an appropriate nonpersistent or other verifiable storage strategy for strict limits and document backup handling before claiming compliance. Add clock-controlled boundary, timezone/DST, migration, crash/restart, offline, invalid-policy, shared UI, accessibility, and provider-contract tests. Do not purge existing user data or enable automatic refresh as part of this planning note.

### Integration code cleanup

Audit all backend integration production packages and newly added frontend composables. Share provider destination URL validation and credential-origin canonicalization at an integration-owned boundary, keeping transport DNS/address checks separate and retaining existing validation behavior. Use the shared UI icon registry and date utilities for weather, retaining timezone/calendar-date semantics and existing public behavior. Extract focused helpers only where responsibilities or measured complexity justify them; preserve transactional fencing, request security, lifecycle ownership and source compatibility. Remove dead code only with verified references. Review the full diff and run frontend tests/lint/build and backend integration tests/build/vet, with focused race tests as applicable. Record complexity evidence and any remaining limits.

### CI browser regression follow-up

Reproduce the failing Home layout and area-drag cases against owned mock servers, preserving the full drag preview, touch support and persisted ordering. Fix layout geometry where it is incorrect and update selectors/assertions that assumed only one integration provider or obsolete confirmation text. Run the affected desktop and mobile cases and the applicable frontend checks; verify the new GitHub checks after pushing.
