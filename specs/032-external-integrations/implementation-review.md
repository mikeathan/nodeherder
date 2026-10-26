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

## Connections board follow-up — 2026-10-26 (AC-34–37)

Implemented the approved grouped board. ConnectionsPage coordinates presentation only; the new
ConnectionWidget/ConnectionBoard/ConnectionCatalogDialog separate summaries, arrangement and
provider discovery. Pure connection-board.ts reuses UK calendar projections and distinguishes
missing/exhausted/confirmed-empty dates. The real WasteBinIcon SVG/palette is shared with the full
results page. Real stretched RouterLinks open the existing connection page without nested buttons;
Refresh/Check are independent siblings. No Details footer, size selector, weather implementation
or changes to Home/Devices/Overview/Panel were introduced.

useConnectionsBoard owns one bounded operation observer plus a minute/focus local-only read;
requests/listeners/timer/route guard follow page/login lifetime. Vuex generation and per-instance-read
sequence guards reject logged-out, aborted and superseded replies. useConnectionBoardLayout owns
immutable drafts and retains them on failed/conflicting saves. Existing vuedraggable handles and
keyboard/tap arrows support reordering. Save/Cancel, hide/restore and confirmed deletion/re-add work.
New setup returns/focuses the board widget; existing-address setup keeps the detail return. Card
navigation resets the connection page scroll and Back restores the originating focus/board position.
Deleting also clears the in-memory current instance. No PII is added to browser persistence.

Backend BoardPreference is optional private instance data with an independent revision. Board GET/
PUT use existing protected auth/CSRF/no-store boundaries. Membership+preference ETags fence lost
updates; all placements commit atomically. Layout writes preserve lifecycle revision/generation,
credentials, snapshots and timestamps. Old records are visible by creation time; no encrypted
schema/AAD/bucket changes or new dependency were needed. Encrypted repository restart, invalid
proposals/rollback and lifecycle independence are tested. See ADR-002 and the board contract.

Actual verification:
- 390 frontend unit tests across 66 suites passed, including pure projections/layout, failed draft
  retention, API validation/auth headers and late/superseded read guards.
- Frontend build, lint and E2E type checks passed. Lint retains seven existing script warnings;
  Vite retains the existing DeviceCard naming warning. No new lint errors.
- 28 isolated Connections browser checks passed on desktop/mobile. Coverage includes direct result
  cards/shared SVGs, independent Refresh, keyboard navigation/focus, whole-content hit area, drag/
  arrows, cancelled/failed/conflicting layouts, host-persisted hide/restore, confirmed delete/re-add,
  normal setup and existing editor/help/auth flows, 1280/736/390/320px, long names, three colours,
  light/dark and scroll return. Screenshots are temporary QA outputs, not production assets.
- Go build/vet, focused service/API suites and affected-package race checks passed.
- The workspace full Go suite failed in internal/automations/scenarios because the already-modified
  configs/automations/0x001788010d7d9d3f.json references a different dial ID than its committed fixture.
  That unrelated user file was left untouched. The complete Go suite passed in a temporary source
  snapshot with the committed automation fixture and these Connections changes.

Browser servers ran only from a temporary source snapshot on 54100/54110/54120 with matching fake
hub ownership handshakes, temporary databases and provider/MQTT fakes. No live provider lookups or
household device commands were performed. A hit-area test initially targeted text under the
intentional stretched link; it now clicks the actual pointer position. Drag coverage checks real
movement and persistence against displayed order rather than assuming a responsive drop inserts
before a particular neighbour. Both corrected checks pass. Initial temporary font-path serving
warnings were corrected in the QA copy before final visual inspection.

Rule review: NH-01 has AC-34–37/contracts/tasks before code; NH-02/BE-02 preserve device paths;
NH-03/BE-05/FE-05 have authenticated no-store/CSRF tests and no new private browser persistence;
NH-04/BE-01/FE-01–02 keep service/repository/transport/store/presentation boundaries and validate
wire data; NH-05/FE-04 have the source-owned and isolated UI evidence above; NH-06/BE-03–04/FE-03
have atomic writes, independent revisions, bounded cleaned-up local observation, explicit failures,
draft retention and destructive confirmation. No rule exception or ratification is claimed.

The user's running app was not restarted or tested against. A backend restart/rebuild is needed
if the local runner does not automatically reload Go changes. No scheduled refresh milestone was
pulled into this follow-up.

## Navigation flicker and connection icon — 2026-10-26 (AC-38–39)

Confirmed first-detail reproduction: with an isolated instance GET deliberately held pending, the
page displayed `Connection` instead of the already-known `Bin collections` title and omitted its
postcode. The regression failed on both desktop and mobile before the fix. Source inspection also
identified ConnectionFeedback's conditional normal-flow Working paragraph above content; its
insertion/removal moved results/tabs during revalidation and back navigation. Initial instance and
provider reads used separate busy intervals, causing additional loading-state churn.

useConnectionDetail now immediately presents only the matching validated private session DTO
(current detail or board entry), revalidates it, and owns one initial busy interval covering both
reads. Board/detail opt into compact accessible progress (out of normal flow), truthful aria-busy
and a visible cold-load message when no cached data exists. Other ConnectionFeedback callers retain
their previous behavior. No artificial transition/animation, fixed header sizing, global shell
change, new library or request polling was added.

Failed instance reads no longer proceed to provider loading and clear their error. A 404 evicts
that matching cached connection instead of resurrecting it; another connection's current data
never becomes the fallback. Revalidation preserves a name typed while Setup is visible. Timer,
focus listener and token cleanup now also happens immediately on page-signal abort/logout.

Header-image comparison: title-only is compact but loses the consistent icon identity used by
other cards. The shared schedule/calendar icon represents the overall collection connection
without selecting a waste stream or copying the first bin colour. It replaces the header wheelie
bin; individual coloured WasteBinIcon SVGs remain unchanged beside the service dates and in the
full Results page. The same calendar treatment works before dates exist.

Verification: 390 frontend unit tests, frontend build, lint (seven existing script warnings), and
E2E type checks passed. All 32 isolated Connections browser tests passed. The delayed-navigation
regression holds instance/provider/board reads separately, verifies the real title/postcode while
pending, and compares exact header height/position, tab position and board-card geometry before
and after completion. It also verifies a typed Setup name survives revalidation. A second regression
verifies wrong-ID isolation, cache eviction, preserved not-found feedback and zero provider reads
after a failed instance lookup. Existing tests now assert a generic schedule icon in the header,
no header bin SVG and exactly one SVG per displayed service. Existing light/dark, long-name,
320/390/736/1280px, keyboard, layout, persistence, deletion and editor checks still pass.

Desktop/mobile screenshots of the new generic icon were reviewed. Tests used the already-isolated
54100/54110/54120 fake hubs and temporary data. No live provider/device/app-restart checks or backend
suite rerun were needed for this frontend-only follow-up; prior backend evidence is unchanged.
NH-01/04 and FE-01 retain the existing planning/composable/presentation boundaries; NH-02/device
behavior and API contracts are unchanged; NH-03/FE-02/05 retain validated matching session cache,
logout and ingress boundaries; NH-05/FE-04 have before/after browser/geometry evidence;
NH-06/FE-03 retain explicit failures/cold loading and cleaned-up resources. No exception claimed.

## Chronological consumer data — 2026-10-26 (AC-40)

The provider's collection/service order was being preserved even though each service's Dates was
already normalized to a sorted unique list. Frontend tests reproducing Black/Brown/Blue ordering,
undated handling and pre-limit multi-resource ordering failed before the fix. Model tests also
reproduced the unsorted canonical resource and missing UK-midnight reordering.

WasteSchedule.Project now stably orders a copied Services slice by each service's first upcoming
ISO date at the UK local date of evaluated_at, with undated/exhausted services last. Normalization,
Instance/Instances and resource reads already use this projection, so consumers receive ordered
services on collection and on later local reads. No provider request is needed at date rollover.
Historical dates, IDs, content revision, source timestamps and stored snapshots are preserved.
Immutable London timezone data is cached once; standard-library tzdata provides a native/minimal-
container fallback. No new third-party dependency, JSON field, route or encryption migration.

The shared frontend date presentation also sorts copied upcoming-date arrays and projected cards,
using raw ISO keys rather than formatted labels. The widget uses that same comparator after merging
resources, before its three-row limit. Equal-date/undated ties keep source order. Neither operation
mutates DTOs or the user's independent connection layout. The canonical wire contract now explicitly
documents this ordering as of evaluated_at and requires consumers to identify bins by stable ID.

Actual verification: 392 frontend tests passed; frontend build/lint/E2E types passed (seven existing
script warnings). All 34 Connections browser tests passed in the isolated fixture, including both
rendered views with deliberately shuffled service DTOs. After adding the backend guarantee, those
two ordering browser cases were rechecked against the updated Go fixture and passed on desktop and
mobile. Pure model/service tests cover normalization/consumer order, source immutability, same-day
and undated ties, exhausted dates, UK midnight/BST-GMT boundaries and stable identity/revision.
Go build/vet and races for models/integrations, internal/integrations and internal/api passed.
The complete Go suite passed in the temporary source copy with the committed automation fixture;
the user's pre-existing modified household automation JSON remains untouched. The workspace full
suite was not repeated against that already-known mismatched fixture. No live provider/device calls
or app restart. A running native backend needs reload/rebuild to expose the new canonical ordering.

Rule review: NH-01/04 and BE-01/FE-01 place the rule in shared pure domain/presentation functions;
NH-02/device semantics are untouched; NH-03/BE-05/FE-02/05 retain auth and wire/PII boundaries;
NH-05/FE-04 have before/after and rendered evidence; NH-06/BE-03–04/FE-03 preserve calendar timezone,
missing/empty distinctions, identity and stored data, with copy/race checks. No rule exception.

## Calendar ownership, presentation cleanup and documentation — 2026-10-26 (AC-41–42)

Timezone audit found no runtime General timezone setting. Compose sets process TZ=Europe/London;
the browser-test server has its own TIMEZONE constant; the existing automation scheduler hardcodes
London; MCP time-scope accepts an explicit timezone/default UTC. These are different contexts, not
one wired setting. A future General setting should govern display/automation defaults while each
provider/resource retains its source calendar. Existing schedules and DST gap/overlap behavior
need a separate scoped plan; this work does not silently change physical automation timing.

Connections now declares its v1 UK source calendar once in backend/models/integrations/calendar.go.
WasteCalendarDate shares cached immutable zone data across resource ordering and horizon checks;
normalization uses WasteTimezone. The standard-library zoneinfo fallback is retained. Frontend
config/integrations.ts owns the supported contract value used by runtime validation and the DTO
type, while calendar formatting consumes schedule.timezone rather than another hardcoded literal.
Operational check timestamps retain their prior browser-local display until the General setting
is designed. No v1 JSON shape, process TZ, automation/device behavior or host configuration changed.

Refactored waste-schedule.ts into a typed presentation composer with narrow service-card and label
helpers. waste-calendar.ts owns source-day conversion, date-only formatting, bounded upcoming dates
and comparison. ISO keys come from formatToParts rather than locale punctuation/order. The formatter
cache is constrained by the validated contract timezone set. Typed view models describe nullable
next-date keys and outputs. connection-board.ts now separates row construction, merged ordering,
missing-date text and lifecycle status using named helpers/guard clauses instead of nested ternaries.
The existing comparator export is preserved. No new framework or provider-specific UI logic.

Added substantive JSDoc for public calendar/view-model/ingress functions and touched composable/API
methods, including side effects, cancellation, revision fencing, custody, non-mutation and unknown
values. Go calendar/type/function comments explain source-zone versus app defaults, error policy,
copying and stable identity. Existing Vue/Go rules now require documentation for new/changed public
APIs and purposeful comments for non-obvious flows, with a final audit step. Vue guidance explicitly
rejects nested ternaries/large inline map callbacks and requires cohesive single-responsibility
helpers. Rules are .agents/rules/vue-engineer.md and go-engineer.md; AGENTS.md was not changed.

Verification: 398 frontend tests across 67 suites, build/lint/E2E types and all 34 isolated Connections
browser checks passed. New source-owned waste-calendar.test.ts/calendar_test.go cover UK midnight,
BST/GMT, process-zone independence, non-mutating chronology, nullable dates and bounded labels;
contract tests reject using an arbitrary app timezone as a v1 resource calendar. Existing source
presentation/identity/ordering/missing-state tests still pass. Go build/vet, races for models/
integrations, internal/integrations and internal/api, and the complete isolated Go suite passed.
Lint retains seven existing script warnings; Vite retains the prior DeviceCard name warning. A
TypeScript AST audit found no missing JSDoc on touched exported functions; returned public actions
were reviewed separately. Documentation-only comment edits did not change execution paths.

Tests ran against temporary storage and fake hubs on 54100/54110/54120, with committed automation
fixtures in the isolated source copy. The user's modified dial configuration and pre-existing index
staging were left untouched. No live provider/device tests or running-app restart. No new visual
behavior is intended; prior appearance, flicker and chronology browser regressions remain green.

Rule review: NH-01/04 and FE/BE-01 retain the current planning set and focused domain/presentation
boundaries; NH-02/device scheduling is unchanged; NH-03/FE-02/05 retain source-calendar ingress and
private-state handling; NH-05/FE-04 have source-owned and cross-stack evidence; NH-06/BE-03–04/FE-03
preserve timezone/day boundaries, non-mutation, missing semantics and resource ownership. No rule
exception or constitutional adoption is claimed.


## 2026-10-26: capability execution refactor and Integrations naming (AC-43–46)

Completed on `spec/032-connections`, preserving the existing feature work. Root integration
composition is separated from `runtime`, YAML `definitions`, reusable `flows/searchselect`,
`operations`, capability registry/`capabilities/waste`, and restricted `transport/httpclient`.
The coordinator is 175 lines and imports no concrete capability. Root composition is the one
shipped capability registration point and injects bundled provider files; runtime imports
neither waste code nor the bundled provider. Setup/definition commands remain with the runtime's
single lock/transaction owner instead of spreading state ownership across mutually importing packages.

Waste normalization, semantic service identity and coverage/renewal live under the waste
capability. Saved resources and previews use a generic immutable JSON payload with a typed common
header; capability codecs validate/project it. Legacy flat waste JSON and encrypted DB schema/AAD
remain unchanged. Stored v1 field names `postcode` and `service_ids` remain for compatibility;
search validation follows its configured input type and identity mappings are generic in Go.
Separate `doc.go` files were removed; public/complex-flow documentation stays beside code.

Registry dispatch is shared by ordinary execution and offline/live previews. A synthetic reading
capability has no waste fields/dates/timezone and proves search/select setup, atomic publication,
refresh, 304, read-time projection, offline/live preview isolation and last-good retention after
normalization failure. Missing codecs produce visible errors for instances and previews without
rewriting/deleting stored data. Existing fencing/cancellation/idempotent recovery tests moved with
source owners. Legacy flat JSON round-trip/buffer-ownership, mapping, flow/input, renewal/calendar,
registry and outbound-policy checks cover new boundaries.

The visible feature is Integrations throughout navigation, page headings, settings/actions and Help.
New routes use `/integrations`; matching `/connections` aliases retain list/detail/setup/provider/
Help/not-found links and authentication. Private component/CSS identifiers remain stable. Home,
Devices, Overview and Panel layouts/data paths are unchanged. Desktop/mobile/light/dark board
screenshots were inspected; isolated browser checks cover keyboard navigation, failures, layout
stability, old bookmarks and the existing setup/editor/board interactions.

Actual verification:

- `go test ./internal/integrations/... ./models/integrations/...` passed in the workspace.
- Complete `go test ./...`, `go build ./...`, `go vet ./...` passed in a temporary source copy.
- `go test -race ./internal/integrations/... ./models/integrations/... ./repository/... ./internal/api/...`
  passed in that isolated copy.
- `npm test -- --runInBand`: 67 suites / 398 tests passed. Frontend build and E2E type checks passed.
- `npm run lint`: zero errors; seven existing warnings in get-version.cjs/start.mjs.
- `npm run test:e2e -- ConnectionsPage.spec.ts`: final 36 desktop/mobile cases passed.
- `gofmt` and `git diff --check` completed; public comments, package dependencies, lifecycle,
  credential boundaries and error/publication paths were reviewed.

The sandbox initially prevented Go cache writes/local test listeners; checks used a temporary
cache and authorized local fixture servers. Initial isolated runs omitted required docs JSON
fixtures, and a subsequent Vite allow-list correction restarted the test server mid-run; those
harness attempts failed. The final copy includes committed docs fixtures, correctly resolved font
paths and test-owned servers on 54100/54110/54120, and the final suites above pass. The pre-existing
modified household automation JSON was never edited; only the isolated copy uses its committed
fixture. No live devices/provider calls, app restart, deployment or push was performed.

NH/BE/FE review: outcomes/tasks precede code (NH-01); device semantics stay unchanged (NH-02/BE-02);
protected routes, redaction, origin grants and egress restrictions remain (NH-03/BE-05/FE-05);
capability/flow/runtime/repository contracts and compatibility are explicit (NH-04/BE-01/FE-01/02);
source-owned tests and actual isolated evidence are recorded (NH-05); immutable legacy payloads,
last-good retention, atomic publication and existing bounded shutdown remain (NH-06/BE-03/04).
Loading/errors/authentication and accessible responsive controls retain browser coverage (FE-03/04).
Constitutions remain proposed; no ratification/exception approval is claimed.

Limit: this introduces a generic capability boundary for the supported HTTP/search-select recipe,
not arbitrary authentication/protocol execution. Adding a provider within existing supported
operations remains YAML/UI work. A genuinely new capability needs a codec/registration and a UI
renderer/ingress contract; additional setup/authentication protocols remain explicitly scoped work.


## Section navigation and layout stability (AC-47–48, T059–62)

The detail-only navigation disappeared on provider/Help route transitions. Its scoped zero bottom
margin also removed the normal shared spacing, leaving exactly 0px above the schedule panel.
Both cases failed in the existing browser companion before implementation.

One presentational IntegrationSections component now renders typed named routes on all relevant
pages. A saved integration carries explicit instance context into provider/Help links; shared
provider entry does not guess an instance. Results/Setup hashes restore selection on reload/back.
Provider Help retains editor query context; draft leave confirmation still applies to data sections
and other providers. Shared navigation has a 20px bottom margin and native keyboard links, visible
focus and exactly one current section. Header Help/back links no longer duplicate section links.

The later delayed-read regression reproduced a 26px header-height change when provider metadata
arrived and a separate 28px offset carried into Help. The editor also initially rebuilt from empty
state and inserted a Working row between its reads. IntegrationPageHeader reserves the subtitle
line and keeps mobile actions in the available title row. Loading announcements stay out of flow.
Up to 32 accepted provider DTOs are cached by key in session-only Vuex state, with explicit removal
and logout cleanup. Cached documents and dirty buffers hydrate before rendering; fresh metadata
cannot replace a dirty buffer or its conflict token. Revalidation has one owned loading interval.
A cold entry uses an inert placeholder, not an interactive fake form. Draft animation frames are
cancelled/replaced and check page/session ownership before touching focus or scroll.

AppShell previously tried to scroll <main>, while .nh-app-main is the actual scrolling element.
Integration path changes now reset that outer container only. Top-level Help tabs record a top
return position; contextual Help anchors still restore the original field/focus position. Hash
anchors and unrelated view behavior remain unchanged. Canonical and legacy integration page keys
also remount instance/provider owners when path identity changes; hash/query changes retain them.

Verification performed:

- 403 unit tests in 68 suites passed, including source-owned navigation and definition-cache tests.
- Frontend production build, E2E type checks and lint passed. Lint retains seven existing warnings
  in get-version.cjs/start.mjs, with zero errors.
- Final ConnectionsPage browser suite: 44 cases passed at desktop 1280x800 and mobile 390x844.
  Cases cover measured spacing, delayed metadata geometry, immediately cached form/draft content,
  four-section continuity, direct provider entry, reload/history, keyboard use, draft confirmation,
  focus restoration, expired sessions, missing instances and the existing board/editor flows.
- Desktop/mobile results, editor and Help screenshots were inspected. Final git diff --check passed.

Tests ran only against test-owned fake hubs/temporary data on 54100/54110/54120. Early runs failed
on the reproduced defects; the intermediate full suite also needed its missing-instance assertion
updated to check the intentional hidden empty subtitle rather than the absence of a subtitle node.
It still asserts the removed property's value is gone, the missing error remains and no provider
request follows a failed instance read. The final 44-case suite passes. Backend tests were not rerun
for this frontend-only correction; no live provider/device requests, app restart, commit or push.

Final review: NH-01 uses the existing spec/plan/tasks set before code; NH-02 has no device effects;
NH-03/FE-05 retain auth and session cleanup without persisting private caches; NH-04/FE-01 keep
routing, presentation, draft ownership and accepted state separate; NH-05 records failing-before
and passing-after evidence; NH-06 bounds cached data and asynchronous ownership. FE-02 keeps wire
contracts unchanged; FE-03/04 preserve loading/errors, guards, keyboard focus and responsive layout.
No dependencies, backend schemas or provider execution behavior changed. Constitutions remain
proposed and no exception/ratification is claimed.


## YAML editor typography and themed canvas (AC-49, T063)

The code textarea inherited font weight 600 from its wrapping label through `font: inherit`.
The code-specific class now explicitly uses regular 400-weight system monospace, 0.875rem
(14px at default scale), line height 1.6 and two-space tab stops. Labels retain their emphasis.
The prior smaller mobile override was removed, retaining readable, user-scalable code on mobile.
The editor uses the theme page background with a subtle 3% accent mix, the strong theme border,
normal foreground and accent caret. Existing focus/contrast tokens, scrolling, resizing and
editing behaviour remain unchanged. Other ordinary form fields are untouched.

Existing YAML editing/repair browser flows passed on desktop and mobile with isolated fake hubs.
Temporary harness instrumentation (not a permanent implementation-mirroring CSS test) measured
weight 400, 14px size and 22.4px line height at both widths. It sampled all 11 palettes × 2 modes ×
11 default/preset-accent choices: 242 combinations at each viewport, minimum text contrast 12.837:1
(lagoon/light/#495057). Light/dark desktop/mobile editor screenshots were inspected. 403 frontend
unit tests, production build and lint passed; lint retains seven existing script warnings and no
errors. Final git diff --check passed. No new dependencies, state, data/auth contracts, provider
calls or backend changes; no commit/push. FE-04 readability and existing theme consistency are
verified; NH-01/04/05 and the existing frontend boundary rules apply, other runtime rules unchanged.
