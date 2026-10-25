# Connections M1: implementation and verification

Date: 2026-10-25 | Scope: manual-first, personal Docker, Connections-only UI
Status: implemented and verified locally. This is an agent self-review, not maintainer approval,
constitutional ratification, a live-provider certification or a deployment record.

## Delivered behavior

Connections appears after Devices in the existing desktop/mobile menu. Users can add bin
collections, enter a postcode, select an explicit address, save, inspect results, Refresh and
Check status. Pending work survives restart for manual continuation. Every signed-in user has
the same controls; no admin/household/owner hierarchy was introduced.

Configure operations provides Form, editable YAML and human-readable Help. Request names and
links are configured in YAML. One collection mapping handles every bin colour. Supported endpoint,
request-field and JSON Pointer repairs can be saved on the host and applied without a release.
Validation, destination review, stale-edit conflicts, reload, previous-revision restore, missing-file
recreation and supplied-example comparison are included. Unsupported transports, authentication
schemes, operation primitives and capabilities still need application support.

The supplied source is [whenisbins.yaml](../../backend/configs/integrations/waste_collection/whenisbins.yaml).
Runtime files use `configs/integrations/waste_collection/whenisbins.yaml` relative to the backend.
Definitions contain shareable request/mapping configuration; selected addresses and credentials
are encrypted private runtime state. UI saves do not commit or push Git.

Home, Devices, Overview and Panel receive no new content, tiles, counts, links or room placements.
Only shared navigation and Connections-specific routing lifecycle changed. No new public MCP,
device-context or settings-WebSocket data is exposed.

## Source ownership and acceptance evidence

Paths below are relative to the repository; Go companions are next to production files and
frontend unit companions mirror production paths under `frontend/src/__tests__/`.

| Acceptance | Implemented owners and evidence |
| --- | --- |
| AC-01–04, AC-22 | `backend/internal/integrations/setup.go`, `coordinator.go`, `service_test.go`, `coordinator_test.go`: current-candidate membership, duplicate Save receipt, exact request body/idempotency key, pending/ambiguous response continuation, unsupported-input rejection; `frontend/e2e/ConnectionsPage.spec.ts` exercises setup/results. |
| AC-06–08 (M1) | `coordinator.go`, `capabilities/waste.go` and companions: last-good preservation, 304/cache scope, backoff, manual renewal, expiry, restart continuation, local-date/DST projection and late-response fencing. AC-05 automatic work remains M2. |
| AC-03/09/18 | `models/integrations/resource.go`, `capabilities/waste.go`, `identity.go` and companions: typed exact/partial results, unknown fields/categories, missing/empty distinctions, per-service coverage, date-only semantics, stable local IDs and encrypted source notes. Provisional/neighbour results are rejected in M1. |
| AC-11/17/20/22 (supported M1 format) | `configuration.go`, `definitions.go`, `operations/request.go`, model definitions and companions; UI `ProviderForm.vue`, `ProviderEditor.vue`, `RequestValues.vue`: arbitrary request IDs, explicit recipe links, phase/capture validation, shared Form/YAML model, endpoint/field repair, atomic activation, test isolation and recovery. Second-provider proof/replacement is M4. |
| AC-07/12/19 (M1 integrity) | `repository/integrations.go`, `integration_definitions.go` and companions, `definitions_test.go`: encrypted atomic aggregates, schema/AAD checks, rollback, file hash CAS, journal reconciliation, missing file recreation and stable IDs after compatible repair. Full replacement preview/mapping is deferred. |
| AC-13/15 | `internal/api/integrations.go`, existing auth/API middleware companions and `http.go`: authenticated equal access, direct-handler rejection, cookie Origin/custom header, cookie precedence, revision preconditions, origin/slot-pinned credential reassociation, private-IP/DNS/redirect/proxy/response limits. |
| AC-09/13 public boundary | Existing `internal/hub_test.go` installs a sentinel connection and asserts actual public device-context/MCP routes and hub snapshots contain neither its data nor resource references; anonymous protected catalog is denied. |
| AC-14/19 | `repository/integrations.go` companion: AES-GCM, separate versioned keyring, mixed-key rotation, backup/isolated restore, missing/corrupt/newer-schema failure; `bootstrap.go`, `janitor.go`, `main.go` companions: optional degraded startup, retention, cancellation and join-before-close. |
| AC-16/21/23 | Frontend contracts/service/store/domain companions and selected Playwright spec: DTO validation, error/loading states, Form/YAML/Help draft preservation, keyboard focus, desktop/mobile, unknown routes, signed-out deep links and no browser PII persistence. |
| AC-26/27 | Navigation unit companion plus selected Playwright spec checks unchanged Home/Devices/Overview/Panel and no feature promotion. Isolated Home fixture has 18 tiles; the user's separate screenshot showed 15. Neither count is rewritten in production. |

AC-10 consumers, full AC-12 replacement, AC-05 automatic refresh and later-milestone portions of
shared criteria remain outside M1. Retired AC-24/25 and T029/T030 were not implemented.

## Final review corrections

The implementation review found and corrected these paths before handoff:

- Configuration captures must be available on both immediate-completion and pending-completion
  paths before Refresh can use them. Optional/unreachable captures cannot satisfy that check.
- Strict YAML/JSON parity rejects coerced/null booleans and unsafe numeric configuration literals;
  opaque provider numeric captures retain exact `json.Number` values in encrypted persistence.
- Credential destination review pins the active slot plus instance/credential revisions. Changing
  slots cannot transfer an unrelated saved token to a new origin; old unused slots remain removable.
- An unchanged response is valid only with a matching retained snapshot and sent validator. Invalid
  304 state gets one unconditional request; source age/content revision do not advance on 304.
- Ambiguous POST retry retains identical saved bytes/key. Disable pauses pending work; Enable does
  no provider I/O and explicit Check can continue the saved attempt before expiry.
- Failed writes cannot publish uncommitted snapshots. Saved-but-not-applied host files report that
  state and reconcile through their journal. Missing files can be recreated from the last active
  source while preserving instances. Unreadable files are reported separately.
- Shutdown joins owned workers before storage closes, including listener/startup failures. A
  deadline leaves storage open beneath a still-running writer rather than closing it unsafely.
- Help/editor restoration clones saved drafts, failed controls preserve useful errors, route keys
  apply only within Connections, logout/late replies cannot restore old private state, and fixture
  identity/run-nonce checks precede browser mutations.

## Actual checks

Toolchains: Go 1.24.8 (`GOTOOLCHAIN`); Node 24 via `/opt/homebrew/opt/node@24/bin`.
All provider/device-facing tests used synthetic transports, fake MQTT and temporary storage.

| Command / scope | Result |
| --- | --- |
| Backend `go test -count=1 -timeout=5m ./...` | Passed across the repository. |
| Backend `go test -race -count=1 ./internal/integrations/... ./repository ./internal/api ./internal/auth ./internal .` | Passed; final missing-file regression also passed in a subsequent `-race` run of `./internal/integrations/...`. |
| Backend `go build ./...`, `go vet ./...` | Passed. |
| Frontend `npm test -- --runInBand` | 59 suites, 360 tests passed. |
| Frontend `npm run build` | Passed, including the final provider-editor recovery controls. |
| Frontend `npm run lint` | Zero errors; seven existing script warnings remain. |
| Frontend `npm run type-check:e2e` | Passed. |
| Frontend `npm run test:e2e -- ConnectionsPage.spec.ts` | Eight passed across 1280×800 desktop and 390×844 mobile. Includes keyboard, errors, Help/drafts, login denial, privacy and unchanged-page assertions. |
| `node --test scripts/manage-stack.test.cjs scripts/create-integration-keyring.test.mjs` | 66 passed. |
| Isolated Compose configuration tests (`NODEHERDER_COMPOSE_TESTS=1`, standalone Compose binary) | 13 passed. Config-only; Docker daemon deliberately unreachable. |
| `bash -n scripts/manage-stack.sh`; ShellCheck 0.11.0 | Passed. |
| Local artifact/link check and `git diff --check` | Passed before the implementation commit. |

Initial Compose plugin discovery failed in its isolated configuration; rerunning with the installed
standalone Compose binary passed. Temporary desktop/mobile Results and editor screenshots were
visually reviewed; no mock HTML/PNG assets are retained in the repository.

Full historical browser-suite coverage was not run; eight targeted scenarios and the repository's
full unit suites were run. No live provider POST, real account/quota test, household device command,
Docker deployment, user-stack restore, sustained-load benchmark or production security audit was
performed. Synthetic tests establish implementation behavior, not external provider availability.

## Rule review

| Rule | Final M1 self-review |
| --- | --- |
| NH-01 | Existing spec → plan → tasks preceded implementation; confirmed manual/personal-Docker scope retained. |
| NH-02 | No device impersonation, commands or device event/state changes. |
| NH-03 | Protected new routes, deterministic ingress/egress checks, origin-bound credentials, encrypted PII and public-boundary negative assertions. |
| NH-04 | Repository/service/API/Vuex boundaries, typed producer/consumer DTOs, bounded YAML operations and ADR alternatives retained. yaml.v3 is an existing dependency made direct; no new framework. |
| NH-05 | Source-owned tests and acceptance mapping above; actual/unrun evidence distinguished. No live-device testing. |
| NH-06 | Explicit limits/retention, atomic publication/recovery, separate key backups, optional degraded feature and worker shutdown tested. |
| BE-01 | Business logic independent of HTTP handlers and exercised with injected transports/storage/time. |
| BE-02 | N/A: DeviceLifetimeService/event/debounce/automation semantics unchanged. |
| BE-03 | Owned bounded workers/queue/janitor, cancellation, controlled-channel race cases and affected-package race runs. No recurring M1 provider timer. |
| BE-04 | Initial schema fails closed on corrupt/newer/unknown storage; encrypted backup/rotation/restore and file recovery tested. Dates/coverage never become guessed zero data. |
| BE-05 | Auth/error/DTO/precondition contracts and actual public MCP/context exclusion tests. No new assistant tools. |
| FE-01 | Existing Vuex/services/composables/shared controls retained; components present feature data. |
| FE-02 | Runtime parsers and typed service/store ingress; isolated real Go feature API and mock/degraded responses agree. |
| FE-03 | Admission, pending, completion, failure, stale/unknown coverage, disabled and offline states remain distinct. |
| FE-04 | Existing theme/shared controls, labels/focus/touch targets, confirmations and recorded desktop/mobile browser evidence. |
| FE-05 | Existing authenticated API helpers, memory-only private state, logout cleanup, write-only tokens and escaped text rendering. |

No scoped exception was requested or self-approved. Maintainer review remains the normal merge
step; the ADR and constitutions have not been declared ratified by this work.

## Operational handoff and remaining roadmap

[docs/connections.md](../../docs/connections.md) describes explicit external keyring provisioning,
optional `--connections`/Compose overlay, host paths, backups, rotation, recovery and manual use.
Without a readable matching keyring, Connections degrades inside its page while device features
remain available. Installation/page reads never create provider lookup jobs.

Provider availability, current access conditions/quotas and host mount permissions must be checked
on the actual installation. Keys are not auto-generated/replaced at runtime. Logical deletion does
not securely erase bbolt free pages or backups; preserve/protect old keys for retained backups.

Next milestones require separate work: M2 automatic scheduling/pending continuation; M3 protected
assistant/automation consumers; M4 second-provider/replacement mapping and identity review. Widgets,
Home/Panel placements, marketplace, scripts and additional household roles remain excluded.

## Follow-up refactor and review (2026-10-25, AC-28–30)

The initial implementation passed behavior tests but its Vue route/form components still mixed
application orchestration, draft lifecycle and rendering; exported Go APIs/types were insufficiently
documented. The user correctly requested another review. T031–T035 track this same-branch follow-up.

### Responsibilities and abstractions

- Route components now compose `useConnectionsCatalog`, `useConnectionSetup`, `useConnectionDetail`
  and `useProviderEditor`. The editor separately composes `useProviderDraft` (document/Help lifecycle)
  and `useProviderPreview` (temporary sample/test state). HTTP stays in the existing service; Vuex
  owns shared accepted state. There is no new state/UI framework or persistence schema.
- `ProviderForm.vue` is an 81-line section coordinator, down from 344 lines. Identity/inputs,
  individual requests/captures, setup wiring, collection mapping and policy have controlled form
  sections. `useEditableModel` owns each section's JSON draft and emits replacements, rather than
  mutating parent props. Pure request/capture renaming lives in `domain/provider-definition.ts`.
- `ProviderEditor.vue` is 134 lines, down from 492. Comparison, explicit destination review and
  test preview panels are presentation components. Results and settings are separate components;
  `domain/waste-schedule.ts` computes upcoming cards once per accepted local projection.
- Go validation has named metadata/input/recipe/request/phase/output checks. Operation claiming,
  pending-status handling, unchanged-cache completion and normalization context are separate helpers.
  Long composite literals/imports are formatted for review. New Go types/exported APIs explain their
  purpose, private/public boundaries and state invariants. Package docs and JSON Pointer comments
  explain `/result/collections/0/name`, `~0`/`~1`, arrays, missing vs null and borrowed values.
  Inline comments cover credential selection, exact POST replay, identity, encryption AAD and publication.

### Concrete defects corrected

- Renaming a capture previously left `state.<name>` uses and alternate completion-path producers
  behind. The pure edit now updates all declared producers and typed uses atomically, preserving
  fixed strings and rejecting collisions/invalid identifiers.
- Replaced observers could finish late and refresh/schedule again. `createOperationObserver` owns
  cancellation and its one timer, checks ownership after each await and stops at terminal status
  or 30 seconds. Tests cover replacement, disposal mid-refresh and late errors.
- Mutations now use the page/session abort signal too. Store commits check generation and signal;
  page scopes own auth handling once. An expired editor session bypasses the dirty-leave dialog,
  clears private state and reaches Login. Token buffers clear on abort/unmount; pending focus frames
  are cancelled. Date refresh listeners/intervals are removed on unmount and do no provider I/O.
- Full-document synchronous deep cloning on each keystroke is removed. Form sections emit owned
  replacements; editor snapshots occur at explicit validation/navigation points. Memory-only drafts
  keep the 32 most recently visited providers (documented in Help), with logout clearing all drafts.
- Expired provider backoff origins previously accumulated. A shared typed limit record and
  `activeProviderLimits` prune encrypted rows and publish the in-memory map only after commit.
  Cleanup failures now emit a sanitized warning instead of being silently discarded.
- Error-target focusing no longer removes actual form controls from keyboard tab order.

### Verification and scope

The follow-up runs full Go tests/build/vet and affected-package race detection, full frontend unit
checks/build/lint, E2E type checks and the isolated desktop/mobile Connections scenarios. Final
counts/results are recorded below. Focused regressions belong to their production
owners: observer, controlled model, provider-definition edits, waste presentation, API service,
Vuex store, Go JSON Pointer and provider-limit cleanup. Existing public-boundary, storage, shutdown
and manual lifecycle tests remain part of the full suites.

Every NH/BE/FE rule was revisited: NH-01 uses the existing spec/plan/tasks; NH-02/BE-02 device
semantics remain untouched; NH-03/BE-05/FE-05 retain authenticated/private contracts and cleanup;
NH-04/BE-01/FE-01 improve ownership without a framework migration; NH-05 maps regressions to source
owners; NH-06/BE-03/BE-04 cover bounded retention/observation, transactional pruning and race/shutdown
checks; FE-02 preserves DTO/mock contracts; FE-03/04 retain feedback, keyboard and viewport evidence.
No exception, maintainer approval or constitutional ratification is claimed.

This is code/lifecycle review and regression evidence, not a heap-profile, sustained-load benchmark
or guarantee that no undiscovered issue remains. No live provider/device/deployment checks were added.
The M1 scope and later roadmap are unchanged.

### Final follow-up check results

- Go 1.24.8: full `go test -count=1 -timeout=5m ./...`, `go build ./...` and `go vet ./...` passed.
  `go test -race -count=1 ./internal/integrations/... ./models/integrations ./repository ./internal/api
  ./internal/auth ./internal .` passed; the final coordinator simplification also passed the
  `./internal/integrations/...` race run.
- Node 24: 63 Jest suites / **371 tests passed**; frontend build, lint and E2E type checking passed.
  Lint still reports the same seven existing script warnings, zero errors.
- **14 isolated Playwright checks passed**, desktop and mobile. This includes slow instance
  loading before Help focus restoration, name/disable result preservation, catalog expiry and
  expiry while a dirty provider editor is open.
- **13 real Compose config-only tests passed** without a running daemon. The original PR CI's
  omitted-false JSON serialization is now accepted only alongside an explicit source false check.
- Local Markdown links and `git diff --check` passed; the exported Go declaration comment audit
  found no undocumented declarations in the new runtime/model/repository APIs. This audit is a
  review aid, not a replacement for the semantic comments/examples or compiler/test evidence.

During iteration, new expiry tests initially left the shell fixture's cookie valid while rejecting
Connections. Both fixtures now represent actual expiry consistently. Original PR CI also exposed
focus restoration while a loading fieldset was disabled: restoration now follows completion of
that local read, and the browser test deliberately delays it. These failures were resolved locally;
remote CI for the new commit is tracked separately from these local results.

## Native first-run correction (AC-31)

The screenshot investigation found the native backend had no key and defaulted to a Docker-only
/run path. Bootstrap now chooses the OS user-config directory when no explicit key file is set.
repository/integration_keyring.go owns first creation: random 256-bit key, private directory/file,
fsynced staging and atomic no-overwrite publication. Concurrent startups reuse the fully published
winner. Existing encrypted DB + missing key and malformed/unsafe existing files fail closed without
changing data. Explicit key paths remain required, and base Compose explicitly declares its mounted
key path. This supersedes earlier universal manual-provisioning statements for native defaults.

Passed locally: full Go suite/build/vet; repository + integrations affected-package race checks;
13 real Compose config-only tests. Source-owned tests cover first run/restart, encrypted round-trip,
byte-preserving lost-key failure, invalid files/permissions and concurrent creators. User HOME/XDG
were replaced with temp directories in tests. No live provider/device calls or running-backend
restart was performed. Native application users must restart/rebuild once to apply this change.

Rule review: NH-01 updates this same spec/plan/tasks; NH-03/06 retain encryption and lost-key
recovery; NH-04/BE-01 put key persistence in the repository and path policy in bootstrap; NH-05
uses source companions; BE-03 concurrency and BE-04 storage/startup behavior are tested; BE-05
public/API boundaries are unchanged. Device semantics and frontend contracts are unchanged.
No exception/ratification is claimed. Separate backups are still required: an app-managed local
key protects accidental data-file exposure, not a compromise of the same OS account.

## Collection card presentation (AC-32/33)

The card misalignment came from the shared adjacent-card margin and independently sized wrapped
headers. WasteScheduleResults now scopes grid margin overrides and shared header/body tracks to
collection cards. No generic UiCard/style behavior changes. After the user's visual feedback,
headers use a medium 48×60 colour-bin SVG, 1rem Black bin/Blue bin/Brown bin titles. Repeated bin-colour/coverage rows below dates are removed; coverage remains under Details. This supersedes earlier large
illustration/colour-title trials. Full service names remain in an accessible native Details section.

The tall status paragraph block is replaced by a <=48px saved/stale/check-time row. Details reveals
source timestamps, manual-refresh explanation, provider notes and original collection descriptions.
Recent checks moves below results. Stale/error/subscription/date semantics and underlying provider
names/identities are unchanged. Known colours use a fixed palette; unknown text cannot become SVG
markup/style. Icons are decorative and labels remain readable without colour.

Passed: 64 Jest suites / 380 tests; frontend build, lint (seven existing warnings, zero errors),
E2E type checking; 16 isolated Connections browser checks at desktop1280/mobile390. Layout checks
assert aligned tops/dividers/dates, 70–88px header heights, <=48px status row, keyboard-operable
Details, retained full descriptions, per-bin SVG colours and no horizontal overflow. Final light/
dark screenshots were visually reviewed outside the repo. The first run safely refused occupied
standard ports. Subsequent runs used a temporary public source snapshot on 54100/54110/54120 with
owned fixture handshakes, separate Vite cache and actual fonts; the user's app was not reused.

NH-01 retains this spec/plan/tasks set; NH-02/03/06 device/data semantics and public boundaries are
unchanged; NH-04/FE-01 keep presentation separate from pure view-model logic; NH-05 source-owned
unit/E2E evidence; FE-02 DTO unchanged; FE-03 retains stale/pending/error meaning; FE-04 covers
responsive/keyboard/colour-independent labels; FE-05 uses escaped text/static SVG only. No backend
production change or new dependency; backend reruns are not required for these frontend-only edits.
Unrelated runtime automation edits in the working tree are excluded from this change.

Final user preference retained: Black bin/Blue bin/Brown bin are the headers. No General waste/
Paper & card/Recycling replacement headings and no repeated colour/limited-dates text below dates.
The last 16-case run verified this exact state, including medium header/icon bounds and the compact
status disclosure; final desktop/mobile light/dark visuals were reviewed.
